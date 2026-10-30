package api_login_code_verify

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
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
		if perr.Err != nil {
			logger.Error("login code verify failed",
				slog.String("error", perr.Err.Error()),
				slog.String("code", string(perr.Code)),
			)
		}
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

	return &LoginCodeVerifyResult{Email: email}, nil
}

// verifyMu serializes the get-compare-invalidate sequence in
// LoginCodeVerify (see comment at the lookup site).
var verifyMu sync.Mutex
