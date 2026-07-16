package api_register

import (
	"context"
	"encoding/json"
	"errors"
	"html"
	"log/slog"
	"net/http"
	"time"

	"github.com/dracory/api"
	"github.com/dracory/auth/internal/core"
	"github.com/dracory/auth/types"
	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/req"
	"github.com/dracory/str"
)

// ApiRegister is the HTTP-level helper that routes registration requests to
// either the passwordless or username+password flow based on the provided
// dependencies.
func ApiRegister(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}

	if deps.AuthKnight {
		akDeps := deps.AuthKnightRegisterDependencies
		akKey := req.GetStringTrimmed(r, "ak_key")
		if akKey == "" {
			api.Respond(w, r, api.Error(types.MsgAuthKnightOnceRequired))
			return
		}

		if akDeps.TemporaryKeyGet == nil {
			api.Respond(w, r, api.Error(types.MsgAuthKnightRegistrationFailed))
			return
		}

		email, errKey := akDeps.TemporaryKeyGet(akKey)
		if errKey != nil || email == "" {
			api.Respond(w, r, api.Error(types.MsgLinkNotValidOrExpired))
			return
		}

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

		if akDeps.UserRegister == nil {
			api.Respond(w, r, api.Error(types.MsgAuthKnightRegistrationFailed))
			return
		}

		options := types.UserAuthOptions{
			UserIp:    req.GetIP(r),
			UserAgent: r.UserAgent(),
		}

		userID, errReg := akDeps.UserRegister(r.Context(), email, firstName, lastName, options)
		if errReg != nil {
			logger.Error("authknight registration failed",
				slog.String("error", errReg.Error()),
			)
			api.Respond(w, r, api.Error(types.MsgAuthKnightRegistrationFailed))
			return
		}
		if userID == "" {
			api.Respond(w, r, api.Error(types.MsgAuthKnightRegistrationFailed))
			return
		}

		token, errToken := str.RandomFromGamma(32, "BCDFGHJKLMNPQRSTVXYZ")
		if errToken != nil {
			logger.Error("authknight registration: token generation failed",
				slog.String("error", errToken.Error()),
			)
			api.Respond(w, r, api.Error(types.MsgFailedToGenerateCode))
			return
		}

		if akDeps.UserStoreAuthToken == nil {
			api.Respond(w, r, api.Error(types.MsgFailedToProcess))
			return
		}

		if errStore := akDeps.UserStoreAuthToken(r.Context(), token, userID, options); errStore != nil {
			logger.Error("authknight registration: auth token store failed",
				slog.String("error", errStore.Error()),
			)
			api.Respond(w, r, api.Error(types.MsgFailedToProcess))
			return
		}

		if akDeps.UseCookies && akDeps.SetAuthCookie != nil {
			akDeps.SetAuthCookie(w, r, token)
		}

		// Consume the ak_key so it cannot be reused
		if akDeps.TemporaryKeySet != nil {
			_ = akDeps.TemporaryKeySet(akKey, "", 0)
		}

		api.Respond(w, r, api.SuccessWithData(types.MsgRegistrationSuccess, map[string]any{
			"token": token,
		}))
		return
	}

	if deps.Passwordless {
		result, err := RegisterPasswordlessInit(r.Context(), r, deps.RegisterPasswordlessInitDependencies)
		if err != nil {
			if err.Err != nil {
				logger.Error("passwordless registration failed",
					slog.String("error", err.Err.Error()),
					slog.String("code", string(err.Code)),
				)
			}
			switch err.Code {
			case RegisterPasswordlessInitErrorCodeValidation:
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

		api.Respond(w, r, api.Success(result.SuccessMessage))
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

	successMessage, errorMessage := deps.RegisterWithUsernameAndPassword(r.Context(), email, password, firstName, lastName, ip, userAgent)
	if errorMessage != "" {
		api.Respond(w, r, api.Error(errorMessage))
		return
	}

	api.Respond(w, r, api.Success(successMessage))
}

// ApiRegisterWithAuth is a convenience wrapper that allows callers to pass a
// types.AuthSharedInterface (such as authImplementation) instead of manually
// wiring Dependencies. It constructs the Dependencies struct using the
// interface accessors and preserves the existing behaviour.
func ApiRegisterWithAuth(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
	deps := Dependencies{}
	deps.Passwordless = a.IsPasswordless()
	deps.Logger = a.GetLogger()

	// Check if this is an AuthKnight instance
	if akAuth, ok := a.(types.AuthAuthKnightInterface); ok && akAuth.IsAuthKnight() {
		deps.AuthKnight = true
		deps.AuthKnightRegisterDependencies = AuthKnightRegisterDependencies{
			TemporaryKeyGet: a.GetFuncTemporaryKeyGet(),
			TemporaryKeySet: a.GetFuncTemporaryKeySet(),
			UseCookies:      a.GetUseCookies(),
			SetAuthCookie: func(w http.ResponseWriter, r *http.Request, token string) {
				a.SetAuthCookie(w, r, token)
			},
		}

		deps.AuthKnightRegisterDependencies.UserRegister = akAuth.GetAuthKnightUserRegister()

		deps.AuthKnightRegisterDependencies.UserStoreAuthToken = func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
			fn := a.GetFuncUserStoreAuthToken()
			if fn == nil {
				return errors.New("user store auth token is not configured")
			}
			return fn(ctx, token, userID, options)
		}

		ApiRegister(w, r, deps)
		return
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
		deps.RegisterPasswordlessInitDependencies = RegisterPasswordlessInitDependencies{
			DisableRateLimit: a.GetDisableRateLimit(),
			TemporaryKeySet:  a.GetFuncTemporaryKeySet(),
			ExpiresSeconds:   0, // let RegisterPasswordlessInit apply default
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
	}

	// Configure username/password registration handler.
	deps.RegisterWithUsernameAndPassword = func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (string, string) {
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
		return res.SuccessMessage, res.ErrorMessage
	}

	ApiRegister(w, r, deps)
}

// RegisterPasswordlessInitDeps defines the dependencies required for the
// passwordless registration init (sending verification code).
type RegisterPasswordlessInitDependencies struct {
	DisableRateLimit bool

	TemporaryKeySet func(key string, value string, expiresSeconds int) error
	ExpiresSeconds  int

	EmailTemplate func(ctx context.Context, email string, verificationCode string) string
	EmailSend     func(ctx context.Context, email string, subject string, body string) error
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

	if firstName == "" {
		return nil, &RegisterPasswordlessInitError{
			Code:    RegisterPasswordlessInitErrorCodeValidation,
			Message: types.MsgFirstNameRequired,
		}
	}

	if lastName == "" {
		return nil, &RegisterPasswordlessInitError{
			Code:    RegisterPasswordlessInitErrorCodeValidation,
			Message: types.MsgLastNameRequired,
		}
	}

	if email == "" {
		return nil, &RegisterPasswordlessInitError{
			Code:    RegisterPasswordlessInitErrorCodeValidation,
			Message: types.MsgEmailRequired,
		}
	}

	if msg := authutils.ValidateEmailFormat(email); msg != "" {
		return nil, &RegisterPasswordlessInitError{
			Code:    RegisterPasswordlessInitErrorCodeValidation,
			Message: msg,
		}
	}

	verificationCode, err := authutils.GenerateVerificationCode(deps.DisableRateLimit)
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

	if errTemp := deps.TemporaryKeySet(verificationCode, string(payload), expires); errTemp != nil {
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

	if errEmail := deps.EmailSend(ctx, email, types.EmailSubjectRegistrationCode, emailContent); errEmail != nil {
		return nil, &RegisterPasswordlessInitError{
			Code: RegisterPasswordlessInitErrorCodeEmailSend,
			Err:  errEmail,
		}
	}

	return &RegisterPasswordlessInitResult{
		SuccessMessage: types.MsgRegistrationCodeSent,
	}, nil
}
