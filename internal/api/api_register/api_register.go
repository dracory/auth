package api_register

import (
	"context"
	"encoding/json"
	"errors"
	"html"
	"net/http"
	"time"

	"github.com/dracory/api"
	"github.com/dracory/auth/internal/core"
	"github.com/dracory/auth/types"
	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

// ApiRegister is the HTTP-level helper that routes registration requests to
// either the passwordless or username+password flow based on the provided
// dependencies.
func ApiRegister(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	if deps.Passwordless {
		result, err := RegisterPasswordlessInit(r.Context(), r, deps.RegisterPasswordlessInitDependencies)
		if err != nil {
			switch err.Code {
			case RegisterPasswordlessInitErrorCodeValidation:
				api.Respond(w, r, api.Error(err.Message))
				return
			case RegisterPasswordlessInitErrorCodeTokenStore,
				RegisterPasswordlessInitErrorCodeSerialization:
				api.Respond(w, r, api.Error("Failed to process request. Please try again later"))
				return
			case RegisterPasswordlessInitErrorCodeEmailSend:
				api.Respond(w, r, api.Error("Failed to send email. Please try again later"))
				return
			default:
				api.Respond(w, r, api.Error("Internal server error. Please try again later"))
				return
			}
		}

		api.Respond(w, r, api.Success(result.SuccessMessage))
		return
	}

	if deps.RegisterWithUsernameAndPassword == nil {
		api.Respond(w, r, api.Error("Registration failed. Please try again later"))
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
	passwordAuth, ok := a.(types.AuthPasswordInterface)
	if !ok {
		if logger := a.GetLogger(); logger != nil {
			logger.Error("registration requires AuthPasswordInterface")
		}
		http.Error(w, "Internal server error. Please try again later", http.StatusInternalServerError)
		return
	}

	deps := Dependencies{}
	deps.Passwordless = a.IsPasswordless()

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
			Message: "First name is required field",
		}
	}

	if lastName == "" {
		return nil, &RegisterPasswordlessInitError{
			Code:    RegisterPasswordlessInitErrorCodeValidation,
			Message: "Last name is required field",
		}
	}

	if email == "" {
		return nil, &RegisterPasswordlessInitError{
			Code:    RegisterPasswordlessInitErrorCodeValidation,
			Message: "Email is required field",
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

	if errEmail := deps.EmailSend(ctx, email, "Registration Code", emailContent); errEmail != nil {
		return nil, &RegisterPasswordlessInitError{
			Code: RegisterPasswordlessInitErrorCodeEmailSend,
			Err:  errEmail,
		}
	}

	return &RegisterPasswordlessInitResult{
		SuccessMessage: "Registration code was sent successfully",
	}, nil
}
