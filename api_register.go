package auth

import (
	"context"
	"html"
	"net/http"

	"github.com/dracory/api"
	apireg "github.com/dracory/auth/internal/api"
	"github.com/dracory/req"
)

func (a authImplementation) apiRegister(w http.ResponseWriter, r *http.Request) {
	// Check CSRF token
	if a.enableCSRFProtection && !a.funcCSRFTokenValidate(r) {
		api.Respond(w, r, api.Forbidden("Invalid CSRF token"))
		return
	}

	if a.passwordless {
		a.apiRegisterPasswordless(w, r)
	} else {
		a.apiRegisterUsernameAndPassword(w, r)
	}
}

func (a authImplementation) apiRegisterPasswordless(w http.ResponseWriter, r *http.Request) {
	// Check rate limit
	if !a.checkRateLimit(w, r, "register") {
		return
	}

	deps := apireg.RegisterPasswordlessInitDeps{
		DisableRateLimit: a.disableRateLimit,
		TemporaryKeySet:  a.funcTemporaryKeySet,
		ExpiresSeconds:   int(DefaultVerificationCodeExpiration.Seconds()),
		EmailTemplate: func(ctx context.Context, email string, verificationCode string) string {
			return a.passwordlessFuncEmailTemplateRegisterCode(ctx, email, verificationCode, UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		},
		EmailSend: func(ctx context.Context, email string, subject string, body string) error {
			return a.passwordlessFuncEmailSend(ctx, email, subject, body)
		},
	}

	result, perr := apireg.RegisterPasswordlessInit(r.Context(), r, deps)
	if perr != nil {
		// Preserve existing logging and error mapping behaviour as closely as possible.
		email := req.GetStringTrimmed(r, "email")
		logger := a.GetLogger()
		ip := req.GetIP(r)
		userAgent := r.UserAgent()

		switch perr.Code {
		case apireg.RegisterPasswordlessInitErrorCodeValidation:
			api.Respond(w, r, api.Error(perr.Message))
			return
		case apireg.RegisterPasswordlessInitErrorCodeCodeGeneration:
			authErr := NewCodeGenerationError(perr.Err)
			logger.Error("registration code generation failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"email", email,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_register_passwordless",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		case apireg.RegisterPasswordlessInitErrorCodeSerialization:
			authErr := NewSerializationError(perr.Err)
			logger.Error("registration data serialization failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"email", email,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_register_passwordless",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		case apireg.RegisterPasswordlessInitErrorCodeTokenStore:
			authErr := NewTokenStoreError(perr.Err)
			logger.Error("registration code token store failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"email", email,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_register_passwordless",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		case apireg.RegisterPasswordlessInitErrorCodeEmailSend:
			authErr := NewEmailSendError(perr.Err)
			logger.Error("registration code email send failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"email", email,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_register_passwordless",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		default:
			authErr := NewInternalError(perr.Err)
			logger.Error("registration init internal error",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"email", email,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_register_passwordless",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		}
	}

	api.Respond(w, r, api.Success(result.SuccessMessage))

}

func (a authImplementation) apiRegisterUsernameAndPassword(w http.ResponseWriter, r *http.Request) {
	// Check rate limit
	if !a.checkRateLimit(w, r, "register") {
		return
	}

	email := req.GetStringTrimmed(r, "email")
	password := req.GetStringTrimmed(r, "password")
	first_name := html.EscapeString(req.GetStringTrimmed(r, "first_name"))
	last_name := html.EscapeString(req.GetStringTrimmed(r, "last_name"))

	response := a.RegisterWithUsernameAndPassword(r.Context(), email, password, first_name, last_name, UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	if response.ErrorMessage != "" {
		api.Respond(w, r, api.Error(response.ErrorMessage))
		return
	}

	api.Respond(w, r, api.Success(response.SuccessMessage))
}
