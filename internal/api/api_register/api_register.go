package api_register

import (
	"context"
	"encoding/json"
	"errors"
	"html"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/dracory/api"
	"github.com/dracory/auth/internal/core"
	"github.com/dracory/auth/internal/rules"
	"github.com/dracory/auth/types"
	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/req"
	"github.com/dracory/rule"
	"github.com/dracory/str"
)

// registerMu serializes the verified-key resolve→consume sequence in
// ApiRegister's token branch (see comment at the lock site).
var registerMu sync.Mutex

// ApiRegister is the HTTP-level helper that routes registration requests to
// either the passwordless or username+password flow based on the provided
// dependencies.
func ApiRegister(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}

	// Business rule: an authenticated session turns registration into
	// profile completion. The email belongs to the verified session —
	// never read from the form — so no code is generated and no user is
	// created.
	if completion := rules.NewProfileCompletionRule(deps.CurrentUserID, r); !completion.Fails() {
		userID := completion.UserID()
		firstName := html.EscapeString(req.GetStringTrimmed(r, "first_name"))
		lastName := html.EscapeString(req.GetStringTrimmed(r, "last_name"))

		for _, check := range []rule.RuleInterface{
			rules.NewRequiredFieldRule(firstName, types.MsgFirstNameRequired),
			rules.NewRequiredFieldRule(lastName, types.MsgLastNameRequired),
		} {
			if check.Fails() {
				api.Respond(w, r, api.Error(check.FailMessageFirst()))
				return
			}
		}

		if deps.UserProfileUpdate == nil {
			logger.Error("registration completion failed: UserProfileUpdate is not configured")
			api.Respond(w, r, api.Error(types.MsgFailedToProcess))
			return
		}

		err := deps.UserProfileUpdate(r.Context(), userID, firstName, lastName, types.UserAuthOptions{
			UserIp:    req.GetIP(r),
			UserAgent: r.UserAgent(),
		})
		if err != nil {
			logger.Error("registration completion failed",
				slog.String("error", err.Error()),
				slog.String("user_id", userID),
			)
			api.Respond(w, r, api.Error(types.MsgFailedToProcess))
			return
		}

		api.Respond(w, r, api.SuccessWithData(types.MsgProfileSaved, map[string]any{
			"redirect": deps.RedirectOnSuccess,
		}))
		return
	}

	// Verified-email token branch: a posted verified_key (issued by an
	// OTP/magic-link verification or the AuthKnight callback) resolves to
	// a verified email and registers via the token — the unified
	// register-first create path. For AuthKnight instances it also runs
	// token-less (returning the "once required" error), preserving the
	// original behaviour.
	verifiedKey := req.GetStringTrimmed(r, "verified_key")
	if deps.AuthKnight || verifiedKey != "" {
		// Serialize token resolve → register → consume so two concurrent
		// submits cannot both pass the lookup before either clears it
		// (TOCTOU, best-effort within one process — same pattern as
		// verifyMu in the code-verify flows).
		registerMu.Lock()
		defer registerMu.Unlock()

		vDeps := deps.VerifiedEmailRegisterDependencies
		if verifiedKey == "" {
			// Only reachable for AuthKnight instances (the branch requires
			// a posted key for everyone else) — hence the AK message.
			api.Respond(w, r, api.Error(types.MsgAuthKnightOnceRequired))
			return
		}

		if vDeps.TemporaryKeyGet == nil {
			api.Respond(w, r, api.Error(types.MsgRegistrationFailedGeneric))
			return
		}

		// Business rule: the token must resolve to a verified email.
		keyRule := rules.NewVerifiedEmailKeyRule(vDeps.TemporaryKeyGet, verifiedKey)
		if keyRule.Fails() {
			api.Respond(w, r, api.Error(keyRule.FailMessageFirst()))
			return
		}
		email := keyRule.Email()

		firstName := html.EscapeString(req.GetStringTrimmed(r, "first_name"))
		lastName := html.EscapeString(req.GetStringTrimmed(r, "last_name"))

		if firstName == "" {
			api.Respond(w, r, api.Error(types.MsgFirstNameRequired))
			return
		}
		if lastName == "" {
			api.Respond(w, r, api.Error(types.MsgLastNameRequired))
			return
		}

		if vDeps.UserRegister == nil {
			api.Respond(w, r, api.Error(types.MsgRegistrationFailedGeneric))
			return
		}

		options := types.UserAuthOptions{
			UserIp:    req.GetIP(r),
			UserAgent: r.UserAgent(),
		}

		userID, errReg := vDeps.UserRegister(r.Context(), email, firstName, lastName, options)
		if errReg != nil {
			logger.Error("verified-email registration failed",
				slog.String("error", errReg.Error()),
			)
			api.Respond(w, r, api.Error(types.MsgRegistrationFailedGeneric))
			return
		}
		if userID == "" {
			api.Respond(w, r, api.Error(types.MsgRegistrationFailedGeneric))
			return
		}

		token, errToken := str.RandomFromGamma(32, "BCDFGHJKLMNPQRSTVXYZ")
		if errToken != nil {
			logger.Error("verified-email registration: token generation failed",
				slog.String("error", errToken.Error()),
			)
			api.Respond(w, r, api.Error(types.MsgFailedToGenerateCode))
			return
		}

		if vDeps.UserStoreAuthToken == nil {
			api.Respond(w, r, api.Error(types.MsgFailedToProcess))
			return
		}

		if errStore := vDeps.UserStoreAuthToken(r.Context(), token, userID, options); errStore != nil {
			logger.Error("verified-email registration: auth token store failed",
				slog.String("error", errStore.Error()),
			)
			api.Respond(w, r, api.Error(types.MsgFailedToProcess))
			return
		}

		if vDeps.UseCookies && vDeps.SetAuthCookie != nil {
			vDeps.SetAuthCookie(w, r, token)
		}

		// Consume the verified key so it cannot be reused
		if vDeps.TemporaryKeySet != nil {
			_ = vDeps.TemporaryKeySet(verifiedKey, "", 0)
		}

		api.Respond(w, r, api.SuccessWithData(types.MsgRegistrationSuccess, map[string]any{
			"token": token,
		}))
		return
	}

	if deps.Passwordless {
		result, err := RegisterPasswordlessInit(r.Context(), r, deps.RegisterPasswordlessInitDependencies)
		if err != nil {
			errText := err.Error()
			if err.Err != nil {
				errText = err.Err.Error()
			}
			logger.Error("passwordless registration failed",
				slog.String("error", errText),
				slog.String("code", string(err.Code)),
			)
			switch err.Code {
			case RegisterPasswordlessInitErrorCodeValidation,
				RegisterPasswordlessInitErrorCodeTooManySends:
				api.Respond(w, r, api.Error(err.Message))
				return
			case RegisterPasswordlessInitErrorCodeTokenStore,
				RegisterPasswordlessInitErrorCodeSerialization:
				api.Respond(w, r, api.Error(types.MsgFailedToProcess))
				return
			case RegisterPasswordlessInitErrorCodeEmailSend:
				api.Respond(w, r, api.Error(types.MsgFailedToSendEmail))
				return
			default:
				api.Respond(w, r, api.Error(types.MsgInternalServer))
				return
			}
		}

		api.Respond(w, r, api.SuccessWithData(result.SuccessMessage, map[string]any{
			"nonce": result.Nonce,
		}))
		return
	}

	if deps.RegisterWithUsernameAndPassword == nil {
		api.Respond(w, r, api.Error(types.MsgRegistrationFailedGeneric))
		return
	}

	email := req.GetStringTrimmed(r, "email")
	password := req.GetStringTrimmed(r, "password")
	firstName := html.EscapeString(req.GetStringTrimmed(r, "first_name"))
	lastName := html.EscapeString(req.GetStringTrimmed(r, "last_name"))
	ip := req.GetIP(r)
	userAgent := r.UserAgent()

	successMessage, nonce, errorMessage := deps.RegisterWithUsernameAndPassword(r.Context(), email, password, firstName, lastName, ip, userAgent)
	if errorMessage != "" {
		api.Respond(w, r, api.Error(errorMessage))
		return
	}

	api.Respond(w, r, api.SuccessWithData(successMessage, map[string]any{
		"nonce": nonce,
	}))
}

// ApiRegisterWithAuth is a convenience wrapper that allows callers to pass a
// types.AuthSharedInterface (such as authImplementation) instead of manually
// wiring Dependencies. It constructs the Dependencies struct using the
// interface accessors and preserves the existing behaviour.
func ApiRegisterWithAuth(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
	a = types.UnwrapAuth(a)
	deps := Dependencies{}
	deps.Passwordless = a.IsPasswordless()
	deps.Logger = a.GetLogger()
	deps.RedirectOnSuccess = a.LinkRedirectOnSuccess()
	deps.CurrentUserID = func(r *http.Request) string {
		return authutils.ResolveCurrentUserID(r, a)
	}

	// Optional-hook resolution via type assertion — extending
	// AuthSharedInterface would break external implementers.
	if provider, ok := a.(interface {
		GetFuncUserProfileUpdate() func(ctx context.Context, userID, firstName, lastName string, options types.UserAuthOptions) error
	}); ok {
		deps.UserProfileUpdate = provider.GetFuncUserProfileUpdate()
	}

	// Check if this is an AuthKnight instance
	if akAuth, ok := a.(types.AuthAuthKnightInterface); ok && akAuth.IsAuthKnight() {
		deps.AuthKnight = true
		deps.VerifiedEmailRegisterDependencies = VerifiedEmailRegisterDependencies{
			TemporaryKeyGet: a.GetFuncTemporaryKeyGet(),
			TemporaryKeySet: a.GetFuncTemporaryKeySet(),
			UseCookies:      a.GetUseCookies(),
			SetAuthCookie: func(w http.ResponseWriter, r *http.Request, token string) {
				a.SetAuthCookie(w, r, token)
			},
		}

		deps.VerifiedEmailRegisterDependencies.UserRegister = akAuth.GetAuthKnightUserRegister()

		deps.VerifiedEmailRegisterDependencies.UserStoreAuthToken = func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
			fn := a.GetFuncUserStoreAuthToken()
			if fn == nil {
				return errors.New("user store auth token is not configured")
			}
			return fn(ctx, token, userID, options)
		}

		ApiRegister(w, r, deps)
		return
	}

	// Verified-email token support for non-AuthKnight impls (OTP /
	// magic-link register-first): the same deps shape, with UserRegister
	// adapted from the passwordless register + find-by-email pair since
	// the passwordless register hook does not return a userID.
	if a.IsPasswordless() {
		if regFn := a.GetPasswordlessUserRegister(); regFn != nil {
			findFn := a.GetPasswordlessUserFindByEmail()
			deps.VerifiedEmailRegisterDependencies = VerifiedEmailRegisterDependencies{
				TemporaryKeyGet: a.GetFuncTemporaryKeyGet(),
				TemporaryKeySet: a.GetFuncTemporaryKeySet(),
				UseCookies:      a.GetUseCookies(),
				SetAuthCookie: func(w http.ResponseWriter, r *http.Request, token string) {
					a.SetAuthCookie(w, r, token)
				},
				UserRegister: func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (string, error) {
					if err := regFn(ctx, email, firstName, lastName, options); err != nil {
						return "", err
					}
					if findFn == nil {
						return "", errors.New("user find by email is not configured")
					}
					return findFn(ctx, email, options)
				},
				UserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
					fn := a.GetFuncUserStoreAuthToken()
					if fn == nil {
						return errors.New("user store auth token is not configured")
					}
					return fn(ctx, token, userID, options)
				},
			}
		}
	}

	passwordAuth, ok := a.(types.AuthPasswordInterface)
	if !ok {
		if logger := a.GetLogger(); logger != nil {
			logger.Error("registration requires AuthPasswordInterface")
		}
		http.Error(w, types.MsgInternalServer, http.StatusInternalServerError)
		return
	}

	deps.Passwordless = a.IsPasswordless()
	deps.Logger = a.GetLogger()

	// Configure passwordless branch dependencies if enabled.
	if deps.Passwordless {
		// Optional-hook resolution via type assertion — extending
		// AuthSharedInterface would break external implementers.
		var emailEnqueue func(ctx context.Context, deliveryKey string) error
		if enq, ok := a.(interface {
			GetFuncEmailEnqueue() func(ctx context.Context, deliveryKey string) error
		}); ok {
			emailEnqueue = enq.GetFuncEmailEnqueue()
		}

		deps.RegisterPasswordlessInitDependencies = RegisterPasswordlessInitDependencies{
			DisableRateLimit:  a.GetDisableRateLimit(),
			EmailSendThrottle: authutils.EmailSendThrottleFromAuth(a),
			TemporaryKeyGet:   a.GetFuncTemporaryKeyGet(),
			TemporaryKeySet:   a.GetFuncTemporaryKeySet(),
			EmailEnqueue:      emailEnqueue,
			ExpiresSeconds:    0, // let RegisterPasswordlessInit apply default
			EmailTemplate: func(ctx context.Context, email string, verificationCode string) string {
				fn := a.GetPasswordlessFuncEmailTemplateRegisterCode()
				if fn == nil {
					return ""
				}
				return fn(ctx, email, verificationCode, types.UserAuthOptions{
					UserIp:    req.GetIP(r),
					UserAgent: r.UserAgent(),
				})
			},
			EmailSend: func(ctx context.Context, email string, subject string, body string) error {
				fn := a.GetPasswordlessFuncEmailSend()
				if fn == nil {
					return errors.New("Passwordless email sender is not configured")
				}
				return fn(ctx, email, subject, body)
			},
		}
		if capt, ok := a.(interface {
			GetFuncEmailCaptured() func(ctx context.Context, email string, verified bool, options types.UserAuthOptions) error
		}); ok {
			if fn := capt.GetFuncEmailCaptured(); fn != nil {
				deps.RegisterPasswordlessInitDependencies.EmailCaptured = func(ctx context.Context, email string, verified bool, options types.UserAuthOptions) error {
					if err := fn(ctx, email, verified, options); err != nil {
						a.GetLogger().Error("email capture hook failed", slog.String("error", err.Error()))
						return err
					}
					return nil
				}
			}
		}
	}

	// Configure username/password registration handler.
	deps.RegisterWithUsernameAndPassword = func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (string, string, string) {
		// Delegate to the core registration helper and adapt its result.
		res := core.RegisterWithUsernameAndPassword(
			ctx,
			email,
			password,
			firstName,
			lastName,
			types.UserAuthOptions{
				UserIp:    ip,
				UserAgent: userAgent,
			},
			passwordAuth,
			time.Hour,
		)
		return res.SuccessMessage, res.Nonce, res.ErrorMessage
	}

	ApiRegister(w, r, deps)
}

// RegisterPasswordlessInitDeps defines the dependencies required for the
// passwordless registration init (sending verification code).
type RegisterPasswordlessInitDependencies struct {
	DisableRateLimit bool

	// EmailSendThrottle configures the per-email send quota; zero value
	// applies the package defaults.
	EmailSendThrottle authutils.EmailSendThrottle

	TemporaryKeyGet func(key string) (string, error)
	TemporaryKeySet func(key string, value string, expiresSeconds int) error
	ExpiresSeconds  int

	EmailTemplate func(ctx context.Context, email string, verificationCode string) string
	EmailSend     func(ctx context.Context, email string, subject string, body string) error

	// EmailEnqueue, when set, is preferred over EmailSend: only an opaque
	// delivery key enters the queue (see utils.DeliverSecret).
	EmailEnqueue func(ctx context.Context, deliveryKey string) error

	// EmailCaptured is an optional lead-capture hook fired after input
	// validation, with verified=false (email submitted, not yet proven).
	EmailCaptured func(ctx context.Context, email string, verified bool, options types.UserAuthOptions) error
}

// RegisterPasswordlessInitErrorCode categorizes possible error sources.
type RegisterPasswordlessInitErrorCode string

const (
	RegisterPasswordlessInitErrorCodeNone           RegisterPasswordlessInitErrorCode = ""
	RegisterPasswordlessInitErrorCodeValidation     RegisterPasswordlessInitErrorCode = "validation"
	RegisterPasswordlessInitErrorCodeCodeGeneration RegisterPasswordlessInitErrorCode = "code_generation"
	RegisterPasswordlessInitErrorCodeSerialization  RegisterPasswordlessInitErrorCode = "serialization"
	RegisterPasswordlessInitErrorCodeTokenStore     RegisterPasswordlessInitErrorCode = "token_store"
	RegisterPasswordlessInitErrorCodeEmailSend      RegisterPasswordlessInitErrorCode = "email_send"
	RegisterPasswordlessInitErrorCodeInternal       RegisterPasswordlessInitErrorCode = "internal"
	RegisterPasswordlessInitErrorCodeTooManySends   RegisterPasswordlessInitErrorCode = "too_many_sends"
)

// RegisterPasswordlessInitError is a structured error for the registration init flow.
type RegisterPasswordlessInitError struct {
	Code    RegisterPasswordlessInitErrorCode
	Message string
	Err     error
}

func (e *RegisterPasswordlessInitError) Error() string {
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

// RegisterPasswordlessInitResult represents a successful registration init.
type RegisterPasswordlessInitResult struct {
	SuccessMessage string
	// Nonce identifies the temporary-key under which the registration
	// payload (including the code) is stored. The client must present it
	// alongside the code on verification.
	Nonce string
}

// RegisterPasswordlessInit contains the core business logic of the passwordless
// registration init API. It validates input, generates the verification code,
// stores a temporary JSON payload and sends the email. It intentionally does
// not perform any logging; callers are responsible for mapping structured
// errors to their own logging and HTTP responses.
func RegisterPasswordlessInit(ctx context.Context, r *http.Request, deps RegisterPasswordlessInitDependencies) (*RegisterPasswordlessInitResult, *RegisterPasswordlessInitError) {
	email := req.GetStringTrimmed(r, "email")
	firstName := html.EscapeString(req.GetStringTrimmed(r, "first_name"))
	lastName := html.EscapeString(req.GetStringTrimmed(r, "last_name"))

	for _, check := range []rule.RuleInterface{
		rules.NewRequiredFieldRule(firstName, types.MsgFirstNameRequired),
		rules.NewRequiredFieldRule(lastName, types.MsgLastNameRequired),
		rules.NewRequiredFieldRule(email, types.MsgEmailRequired),
		rules.NewEmailFormatRule(email),
	} {
		if check.Fails() {
			return nil, &RegisterPasswordlessInitError{
				Code:    RegisterPasswordlessInitErrorCodeValidation,
				Message: check.FailMessageFirst(),
			}
		}
	}

	// Lead capture on submission — the email is not yet verified. The
	// hook is best-effort — wiring sites log failures.
	if deps.EmailCaptured != nil {
		_ = deps.EmailCaptured(ctx, email, false, types.UserAuthOptions{
			UserIp:    req.GetIP(r),
			UserAgent: r.UserAgent(),
		})
	}

	// Per-email send throttle — prevents the endpoint being used for email
	// bombing; enforced before code generation.
	if check := rules.NewEmailSendThrottleRule(deps.TemporaryKeyGet, email, deps.EmailSendThrottle); check.Fails() {
		return nil, &RegisterPasswordlessInitError{
			Code:    RegisterPasswordlessInitErrorCodeTooManySends,
			Message: check.FailMessageFirst(),
		}
	}

	verificationCode, err := authutils.GenerateVerificationCode(deps.DisableRateLimit)
	if err != nil {
		return nil, &RegisterPasswordlessInitError{
			Code: RegisterPasswordlessInitErrorCodeCodeGeneration,
			Err:  err,
		}
	}

	nonce, err := authutils.GenerateNonce()
	if err != nil {
		return nil, &RegisterPasswordlessInitError{
			Code: RegisterPasswordlessInitErrorCodeCodeGeneration,
			Err:  err,
		}
	}

	payload, errJSON := json.Marshal(map[string]string{
		"email":      email,
		"first_name": firstName,
		"last_name":  lastName,
		"code":       verificationCode,
	})
	if errJSON != nil {
		return nil, &RegisterPasswordlessInitError{
			Code: RegisterPasswordlessInitErrorCodeSerialization,
			Err:  errJSON,
		}
	}

	if deps.TemporaryKeySet == nil {
		return nil, &RegisterPasswordlessInitError{
			Code: RegisterPasswordlessInitErrorCodeTokenStore,
			Err:  errors.New("temporary key store is not configured"),
		}
	}

	expires := deps.ExpiresSeconds
	if expires <= 0 {
		// fallback to one hour
		expires = 3600
	}

	// The nonce — not the code — is the storage key; the payload carries the
	// code for constant-time comparison at verify time.
	if errTemp := deps.TemporaryKeySet(authutils.RegisterCodeKeyPrefix+nonce, string(payload), expires); errTemp != nil {
		return nil, &RegisterPasswordlessInitError{
			Code: RegisterPasswordlessInitErrorCodeTokenStore,
			Err:  errTemp,
		}
	}

	// Seed the per-nonce attempt counter alongside the payload.
	if errTemp := deps.TemporaryKeySet(authutils.RegisterCodeAttemptsKeyPrefix+nonce, "0", expires); errTemp != nil {
		return nil, &RegisterPasswordlessInitError{
			Code: RegisterPasswordlessInitErrorCodeTokenStore,
			Err:  errTemp,
		}
	}

	if deps.EmailTemplate == nil || deps.EmailSend == nil {
		return nil, &RegisterPasswordlessInitError{
			Code: RegisterPasswordlessInitErrorCodeInternal,
			Err:  errors.New("email template or sender is not configured"),
		}
	}

	emailContent := deps.EmailTemplate(ctx, email, verificationCode)

	// Queue-safe delivery preferred: when EmailEnqueue is set only the
	// opaque delivery key enters the queue; the code stays in the TTL-bound
	// temporary-key store for the worker to resolve.
	errEmail := authutils.DeliverSecret(ctx, deps.TemporaryKeyGet, deps.TemporaryKeySet,
		deps.EmailEnqueue, deps.EmailSend, email, verificationCode,
		types.EmailSubjectRegistrationCode, emailContent, expires)
	if errEmail != nil {
		return nil, &RegisterPasswordlessInitError{
			Code: RegisterPasswordlessInitErrorCodeEmailSend,
			Err:  errEmail,
		}
	}

	// Consume send quota only after the email was handed off successfully.
	if deps.TemporaryKeyGet != nil {
		_ = authutils.RecordEmailSend(deps.TemporaryKeyGet, deps.TemporaryKeySet, email, deps.EmailSendThrottle)
	}

	return &RegisterPasswordlessInitResult{
		SuccessMessage: types.MsgRegistrationCodeSent,
		Nonce:          nonce,
	}, nil
}
