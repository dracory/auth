package api_login

import (
	"context"
	"errors"
	"net/http"

	"github.com/dracory/api"
	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

// LoginPasswordlessDeps defines the dependencies required for the passwordless login flow.
// It is intentionally decoupled from the authImplementation type to avoid import cycles.
type LoginPasswordlessDeps struct {
	DisableRateLimit bool

	TemporaryKeySet func(key string, value string, expiresSeconds int) error

	ExpiresSeconds int

	EmailTemplate func(ctx context.Context, email string, verificationCode string) string
	EmailSend     func(ctx context.Context, email string, subject string, body string) error
}

// LoginPasswordlessErrorCode categorizes the possible error sources in the passwordless login flow.
type LoginPasswordlessErrorCode string

const (
	LoginPasswordlessErrorCodeNone           LoginPasswordlessErrorCode = ""
	LoginPasswordlessErrorCodeValidation     LoginPasswordlessErrorCode = "validation"
	LoginPasswordlessErrorCodeCodeGeneration LoginPasswordlessErrorCode = "code_generation"
	LoginPasswordlessErrorCodeTokenStore     LoginPasswordlessErrorCode = "token_store"
	LoginPasswordlessErrorCodeEmailSend      LoginPasswordlessErrorCode = "email_send"
)

// LoginPasswordlessError represents a structured error for the passwordless login flow.
// Message is intended to be user-facing for validation errors. For internal failures,
// callers are expected to wrap Err into their own error types.
type LoginPasswordlessError struct {
	Code    LoginPasswordlessErrorCode
	Message string
	Err     error
}

func (e *LoginPasswordlessError) Error() string {
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

// LoginPasswordlessResult represents a successful passwordless login operation.
type LoginPasswordlessResult struct {
	SuccessMessage string
}

// LoginPasswordless contains the core business logic of the passwordless login API.
// It performs input validation, code generation, temporary token storage and prepares
// the email contents. It does *not* perform logging; callers are responsible for
// logging and mapping structured errors to their own error types.
func LoginPasswordless(ctx context.Context, r *http.Request, deps LoginPasswordlessDeps) (*LoginPasswordlessResult, *LoginPasswordlessError) {
	email := req.GetStringTrimmed(r, "email")

	if email == "" {
		return nil, &LoginPasswordlessError{
			Code:    LoginPasswordlessErrorCodeValidation,
			Message: "Email is required field",
		}
	}

	if msg := authutils.ValidateEmailFormat(email); msg != "" {
		return nil, &LoginPasswordlessError{
			Code:    LoginPasswordlessErrorCodeValidation,
			Message: msg,
		}
	}

	verificationCode, err := authutils.GenerateVerificationCode(deps.DisableRateLimit)
	if err != nil {
		return nil, &LoginPasswordlessError{
			Code: LoginPasswordlessErrorCodeCodeGeneration,
			Err:  err,
		}
	}

	if deps.TemporaryKeySet == nil {
		return nil, &LoginPasswordlessError{
			Code: LoginPasswordlessErrorCodeTokenStore,
			Err:  errors.New("temporary key store is not configured"),
		}
	}

	expires := deps.ExpiresSeconds
	if expires <= 0 {
		// Fallback to one hour if the caller forgets to configure this; this keeps
		// the business logic package independent from the auth package constants
		// while still providing a safe default.
		expires = 3600
	}

	if errTemp := deps.TemporaryKeySet(verificationCode, email, expires); errTemp != nil {
		return nil, &LoginPasswordlessError{
			Code: LoginPasswordlessErrorCodeTokenStore,
			Err:  errTemp,
		}
	}

	if deps.EmailTemplate == nil || deps.EmailSend == nil {
		return nil, &LoginPasswordlessError{
			Code: LoginPasswordlessErrorCodeEmailSend,
			Err:  errors.New("email template or sender is not configured"),
		}
	}

	emailContent := deps.EmailTemplate(ctx, email, verificationCode)

	if errEmail := deps.EmailSend(ctx, email, "Login Code", emailContent); errEmail != nil {
		return nil, &LoginPasswordlessError{
			Code: LoginPasswordlessErrorCodeEmailSend,
			Err:  errEmail,
		}
	}

	return &LoginPasswordlessResult{
		SuccessMessage: "Login code was sent successfully",
	}, nil
}

// Dependencies aggregates all dependencies required for handling the /api/login
// endpoint for both passwordless and username+password flows.
type Dependencies struct {
	// Passwordless controls which flow is executed. When true, the passwordless
	// email-code flow is used; otherwise the username+password flow is used.
	Passwordless bool

	// PasswordlessDeps contains the business-logic dependencies for the
	// passwordless login flow.
	PasswordlessDeps LoginPasswordlessDeps

	// LoginWithUsernameAndPassword performs the username+password login flow
	// and returns success message, token and error message. If error message is
	// non-empty, the operation is considered failed.
	LoginWithUsernameAndPassword func(ctx context.Context, email, password, ip, userAgent string) (successMessage, token, errorMessage string)

	// UseCookies controls whether the auth token should be written as a cookie
	// when the username+password flow succeeds.
	UseCookies bool

	// SetAuthCookie writes the auth cookie. It is only used when UseCookies is
	// true and must be non-nil in that case.
	SetAuthCookie func(w http.ResponseWriter, r *http.Request, token string)
}

// ApiLogin is the HTTP-level handler that combines passwordless and
// username+password login flows behind a shared interface.
func ApiLogin(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	if deps.Passwordless {
		result, perr := LoginPasswordless(r.Context(), r, deps.PasswordlessDeps)
		if perr != nil {
			switch perr.Code {
			case LoginPasswordlessErrorCodeValidation:
				api.Respond(w, r, api.Error(perr.Message))
				return
			case LoginPasswordlessErrorCodeTokenStore:
				api.Respond(w, r, api.Error("Failed to process request. Please try again later"))
				return
			case LoginPasswordlessErrorCodeEmailSend:
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

	if deps.LoginWithUsernameAndPassword == nil {
		api.Respond(w, r, api.Error("Internal server error. Please try again later"))
		return
	}

	email := req.GetStringTrimmed(r, "email")
	password := req.GetStringTrimmed(r, "password")
	ip := req.GetIP(r)
	userAgent := r.UserAgent()

	successMessage, token, errMessage := deps.LoginWithUsernameAndPassword(r.Context(), email, password, ip, userAgent)
	if errMessage != "" {
		api.Respond(w, r, api.Error(errMessage))
		return
	}

	if deps.UseCookies && deps.SetAuthCookie != nil {
		deps.SetAuthCookie(w, r, token)
	}

	api.Respond(w, r, api.SuccessWithData(successMessage, map[string]any{
		"token": token,
	}))
}
