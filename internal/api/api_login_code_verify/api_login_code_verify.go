package api_login_code_verify

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/dracory/api"
	"github.com/dracory/auth/internal/rules"
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/req"
	"github.com/dracory/rule"
)

// Dependencies defines dependencies for verifying a passwordless
// login code and authenticating the user.
type Dependencies struct {
	DisableRateLimit bool

	TemporaryKeyGet func(key string) (string, error)

	// TemporaryKeySet is optional; when present the consumed code is
	// invalidated after a successful verification (single-use).
	// It also backs the per-nonce attempt counter — without it the
	// counter never increments and MaxCodeAttempts becomes a no-op.
	TemporaryKeySet func(key string, value string, expiresSeconds int) error

	// TemporaryKeyDelete is optional; when present it is preferred over
	// TemporaryKeySet for invalidating consumed codes (store-agnostic
	// delete vs the "empty value, 0 expiry" convention).
	TemporaryKeyDelete func(key string) error

	// AuthenticateViaUsername is called on successful code verification
	// to perform authentication (token generation, cookies, etc.) and send
	// the final HTTP response.
	AuthenticateViaUsername func(w http.ResponseWriter, r *http.Request, email string)

	// CreateUserOnRegister defers user creation to the register form: when
	// the verified email has no account, the flow issues a verified-email
	// token and the handler redirects to RegisterURL instead of calling
	// AuthenticateViaUsername (which would create a session-first user).
	CreateUserOnRegister bool

	// UserExists reports whether the email already has an account. Used by
	// CreateUserOnRegister; nil falls back to AuthenticateViaUsername.
	UserExists func(ctx context.Context, email string, options types.UserAuthOptions) (bool, error)

	// RegisterURL is the register page the verified-email token redirects
	// to (e.g. "/auth/register"). Only used with CreateUserOnRegister.
	RegisterURL string

	// EmailCaptured is an optional lead-capture hook fired after the code
	// matches, with verified=true.
	EmailCaptured func(ctx context.Context, email string, verified bool, options types.UserAuthOptions) error

	// Logger is used to log internal errors. If nil, slog.Default() is used.
	Logger *slog.Logger
}

// LoginCodeVerifyErrorCode categorizes error sources.
type LoginCodeVerifyErrorCode string

const (
	LoginCodeVerifyErrorCodeNone            LoginCodeVerifyErrorCode = ""
	LoginCodeVerifyErrorCodeValidation      LoginCodeVerifyErrorCode = "validation"
	LoginCodeVerifyErrorCodeCodeExpired     LoginCodeVerifyErrorCode = "code_expired"
	LoginCodeVerifyErrorCodeTooManyAttempts LoginCodeVerifyErrorCode = "too_many_attempts"
)

// LoginCodeVerifyError represents a structured error in the login code
// verification flow.
type LoginCodeVerifyError struct {
	Code    LoginCodeVerifyErrorCode
	Message string
	Err     error
}

func (e *LoginCodeVerifyError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Code)
}

// LoginCodeVerifyResult represents a successful verification.
type LoginCodeVerifyResult struct {
	Email string
	// VerifiedEmailToken is set under CreateUserOnRegister when the email
	// has no account — the handler redirects to the register form with it
	// instead of authenticating. Empty means proceed to authentication.
	VerifiedEmailToken string
}

// ApiLoginCodeVerify is the HTTP-level helper that wires request/response
// handling to the core LoginCodeVerify business logic using the provided
// dependencies.
func ApiLoginCodeVerify(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}

	result, perr := LoginCodeVerify(r.Context(), r, deps)
	if perr != nil {
		errText := perr.Error()
		if perr.Err != nil {
			errText = perr.Err.Error()
		}
		logger.Error("login code verify failed",
			slog.String("error", errText),
			slog.String("code", string(perr.Code)),
		)
		switch perr.Code {
		case LoginCodeVerifyErrorCodeValidation,
			LoginCodeVerifyErrorCodeCodeExpired,
			LoginCodeVerifyErrorCodeTooManyAttempts:
			api.Respond(w, r, api.Error(perr.Message))
			return
		default:
			api.Respond(w, r, api.Error(types.MsgVerificationCodeExpired))
			return
		}
	}

	// Register-first: no account for this verified email — hand the client
	// the verified-email token and the register URL instead of logging in.
	if result.VerifiedEmailToken != "" {
		registerURL := deps.RegisterURL
		if registerURL == "" {
			registerURL = "/register"
		}
		separator := "?"
		if strings.Contains(registerURL, "?") {
			separator = "&"
		}
		api.Respond(w, r, api.SuccessWithData(types.MsgVerificationSuccess, map[string]any{
			"redirect": registerURL + separator + "verified_key=" + url.QueryEscape(result.VerifiedEmailToken),
		}))
		return
	}

	if deps.AuthenticateViaUsername == nil {
		api.Respond(w, r, api.Error(types.MsgFailedToProcess))
		return
	}

	deps.AuthenticateViaUsername(w, r, result.Email)
}

// ApiLoginCodeVerifyWithAuth is a convenience wrapper that allows callers to
// pass a types.AuthSharedInterface (such as authImplementation) instead of
// manually wiring Dependencies. It constructs the Dependencies struct using
// the interface accessors and preserves the existing behaviour.
func ApiLoginCodeVerifyWithAuth(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
	deps := Dependencies{
		DisableRateLimit: a.GetDisableRateLimit(),
		TemporaryKeyGet:  a.GetFuncTemporaryKeyGet(),
		TemporaryKeySet:  a.GetFuncTemporaryKeySet(),
		Logger:           a.GetLogger(),
	}
	if deleter, ok := a.(interface {
		GetFuncTemporaryKeyDelete() func(key string) error
	}); ok {
		deps.TemporaryKeyDelete = deleter.GetFuncTemporaryKeyDelete()
	}

	// Register-first wiring — all via optional interfaces so
	// AuthSharedInterface stays stable.
	if pol, ok := a.(interface {
		GetUserCreationPolicy() types.UserCreationPolicy
	}); ok {
		deps.CreateUserOnRegister = pol.GetUserCreationPolicy() == types.CreateUserOnRegister
	}
	if exists, ok := a.(interface {
		GetFuncUserExists() func(ctx context.Context, email string, options types.UserAuthOptions) (bool, error)
	}); ok {
		deps.UserExists = exists.GetFuncUserExists()
	}
	if deps.UserExists == nil {
		// Fall back to the passwordless lookup: a found userID means the
		// account exists, anything else (error or empty) means it does not.
		if fn := a.GetPasswordlessUserFindByEmail(); fn != nil {
			deps.UserExists = func(ctx context.Context, email string, options types.UserAuthOptions) (bool, error) {
				userID, err := fn(ctx, email, options)
				return err == nil && userID != "", nil
			}
		}
	}
	if capt, ok := a.(interface {
		GetFuncEmailCaptured() func(ctx context.Context, email string, verified bool, options types.UserAuthOptions) error
	}); ok {
		if fn := capt.GetFuncEmailCaptured(); fn != nil {
			deps.EmailCaptured = func(ctx context.Context, email string, verified bool, options types.UserAuthOptions) error {
				if err := fn(ctx, email, verified, options); err != nil {
					a.GetLogger().Error("email capture hook failed", slog.String("error", err.Error()))
					return err
				}
				return nil
			}
		}
	}
	deps.RegisterURL = a.LinkRegister()

	deps.AuthenticateViaUsername = func(w http.ResponseWriter, r *http.Request, email string) {
		a.AuthenticateViaUsername(w, r, email, "", "")
	}

	ApiLoginCodeVerify(w, r, deps)
}

// LoginCodeVerify encapsulates the core business logic for verifying a
// passwordless login code. It mirrors the original validation rules and error
// messages but does not perform authentication or write HTTP responses.
func LoginCodeVerify(ctx context.Context, r *http.Request, deps Dependencies) (*LoginCodeVerifyResult, *LoginCodeVerifyError) {
	verificationCode := req.GetStringTrimmed(r, "verification_code")
	options := types.UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	}
	nonce := req.GetStringTrimmed(r, "nonce")
	email := req.GetStringTrimmed(r, "email")

	for _, check := range []rule.RuleInterface{
		rules.NewVerificationCodeRule(verificationCode, deps.DisableRateLimit),
		rules.NewNonceRule(nonce),
		rules.NewRequiredFieldRule(email, types.MsgEmailRequired),
	} {
		if check.Fails() {
			return nil, &LoginCodeVerifyError{
				Code:    LoginCodeVerifyErrorCodeValidation,
				Message: check.FailMessageFirst(),
			}
		}
	}

	if deps.TemporaryKeyGet == nil {
		return nil, &LoginCodeVerifyError{
			Code:    LoginCodeVerifyErrorCodeCodeExpired,
			Message: types.MsgVerificationCodeExpired,
		}
	}

	// The nonce — not the code — is the storage key. The stored value is
	// "email:code" and must match exactly what the client presents.
	// verifyMu serializes the get-compare-invalidate sequence so two
	// concurrent requests cannot both pass the compare before either
	// deletes the key (TOCTOU). Cross-process deployments still need an
	// atomic store; this is best-effort within one process.
	cacheKey := utils.LoginCodeKeyPrefix + nonce
	attemptsKey := utils.LoginCodeAttemptsKeyPrefix + nonce
	verifyMu.Lock()
	defer verifyMu.Unlock()

	// Per-nonce attempt limit, enforced before the code lookup so a known
	// nonce cannot be used to probe the code space until expiry.
	attemptsLimit := rules.NewCodeAttemptsLimitRule(
		deps.TemporaryKeyGet, deps.TemporaryKeyDelete, deps.TemporaryKeySet,
		cacheKey, attemptsKey)
	if attemptsLimit.Fails() {
		return nil, &LoginCodeVerifyError{
			Code:    LoginCodeVerifyErrorCodeTooManyAttempts,
			Message: attemptsLimit.FailMessageFirst(),
		}
	}
	attempts := attemptsLimit.Attempts()

	stored, errCode := deps.TemporaryKeyGet(cacheKey)
	if errCode != nil || stored == "" {
		return nil, &LoginCodeVerifyError{
			Code:    LoginCodeVerifyErrorCodeCodeExpired,
			Message: types.MsgVerificationCodeExpired,
		}
	}

	// Count every attempt against the nonce before the compare — including
	// the (rare) successful one, which deletes the keys anyway.
	// The fixed 3600 TTL mirrors the send fallback; the counter only needs
	// to live as long as the code, and both are deleted together.
	if deps.TemporaryKeySet != nil {
		if err := deps.TemporaryKeySet(attemptsKey, fmt.Sprintf("%d", attempts+1), 3600); err != nil {
			logger := deps.Logger
			if logger == nil {
				logger = slog.Default()
			}
			logger.Error("login code verify: failed to update attempts counter",
				slog.String("error", err.Error()))
		}
	}

	expected := email + ":" + verificationCode
	if check := rules.NewCodeMatchRule(stored, expected); check.Fails() {
		return nil, &LoginCodeVerifyError{
			Code:    LoginCodeVerifyErrorCodeCodeExpired,
			Message: check.FailMessageFirst(),
		}
	}

	// Single-use: invalidate the consumed code and its attempt counter.
	// Prefer a real delete when available; otherwise fall back to the
	// "empty value, 0 expiry" convention the module already uses.
	invalidateErr := utils.InvalidateTempKey(deps.TemporaryKeyDelete, deps.TemporaryKeySet, cacheKey)
	_ = utils.InvalidateTempKey(deps.TemporaryKeyDelete, deps.TemporaryKeySet, attemptsKey)
	if invalidateErr != nil {
		logger := deps.Logger
		if logger == nil {
			logger = slog.Default()
		}
		logger.Error("login code verify: failed to invalidate consumed code",
			slog.String("error", invalidateErr.Error()))
	}

	// Lead capture: the email just proved ownership of the inbox. The
	// hook is best-effort — wiring sites log failures.
	if deps.EmailCaptured != nil {
		_ = deps.EmailCaptured(ctx, email, true, options)
	}

	result := &LoginCodeVerifyResult{Email: email}

	// Register-first policy: accounts are created by the register form,
	// not at verify time. Verified-but-new emails get a token instead of
	// a session. Misconfiguration or an existence-check failure falls
	// through to the normal authenticate path so a store hiccup cannot
	// lock users out — logged loudly since it also means an incomplete
	// user row the policy exists to prevent.
	if deps.CreateUserOnRegister {
		logger := deps.Logger
		if logger == nil {
			logger = slog.Default()
		}
		switch {
		case deps.UserExists == nil:
			logger.Warn("login code verify: CreateUserOnRegister set but UserExists is not configured — falling back to session-first registration")
		case deps.TemporaryKeySet == nil:
			logger.Warn("login code verify: CreateUserOnRegister set but TemporaryKeySet is not configured — falling back to session-first registration")
		default:
			exists, errExists := deps.UserExists(ctx, email, options)
			switch {
			case errExists != nil:
				logger.Warn("login code verify: UserExists failed — falling back to session-first registration",
					slog.String("error", errExists.Error()))
			case !exists:
				token, errToken := utils.IssueVerifiedEmailToken(deps.TemporaryKeySet, email, 0)
				if errToken != nil {
					logger.Error("login code verify: verified-email token issue failed",
						slog.String("error", errToken.Error()))
				} else {
					result.VerifiedEmailToken = token
				}
			}
		}
	}

	return result, nil
}

// verifyMu serializes the get-compare-invalidate sequence in
// LoginCodeVerify (see comment at the lookup site).
var verifyMu sync.Mutex
