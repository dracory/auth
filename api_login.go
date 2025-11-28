package auth

import (
	"context"
	"net/http"

	"github.com/dracory/api"
	apilogin "github.com/dracory/auth/internal/api"
	"github.com/dracory/req"
)

func (a authImplementation) apiLogin(w http.ResponseWriter, r *http.Request) {
	if a.passwordless {
		a.apiLoginPasswordless(w, r)
	} else {
		a.apiLoginUsernameAndPassword(w, r)
	}
}

func (a authImplementation) apiLoginPasswordless(w http.ResponseWriter, r *http.Request) {
	// Check rate limit
	if !a.checkRateLimit(w, r, "login") {
		return
	}

	// Delegate core business logic to internal/api while keeping logging and
	// HTTP response behaviour in this package.
	deps := apilogin.LoginPasswordlessDeps{
		DisableRateLimit: a.disableRateLimit,
		TemporaryKeySet:  a.funcTemporaryKeySet,
		ExpiresSeconds:   int(DefaultVerificationCodeExpiration.Seconds()),
		EmailTemplate: func(ctx context.Context, email string, verificationCode string) string {
			return a.passwordlessFuncEmailTemplateLoginCode(ctx, email, verificationCode, UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		},
		EmailSend: func(ctx context.Context, email string, subject string, body string) error {
			return a.passwordlessFuncEmailSend(ctx, email, subject, body)
		},
	}

	result, perr := apilogin.LoginPasswordless(r.Context(), r, deps)
	if perr != nil {
		// Preserve existing logging and error mapping behaviour.
		email := req.GetStringTrimmed(r, "email")
		ip := req.GetIP(r)
		userAgent := r.UserAgent()
		logger := a.GetLogger()

		switch perr.Code {
		case apilogin.LoginPasswordlessErrorCodeValidation:
			api.Respond(w, r, api.Error(perr.Message))
			return
		case apilogin.LoginPasswordlessErrorCodeCodeGeneration:
			authErr := NewCodeGenerationError(perr.Err)
			logger.Error("login code generation failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"email", email,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_login_passwordless",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		case apilogin.LoginPasswordlessErrorCodeTokenStore:
			authErr := NewTokenStoreError(perr.Err)
			logger.Error("login code token store failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"email", email,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_login_passwordless",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		case apilogin.LoginPasswordlessErrorCodeEmailSend:
			authErr := NewEmailSendError(perr.Err)
			logger.Error("login code email send failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"email", email,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_login_passwordless",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		default:
			authErr := NewInternalError(perr.Err)
			logger.Error("login code internal error",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"email", email,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_login_passwordless",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		}
	}

	api.Respond(w, r, api.Success(result.SuccessMessage))

}

func (a authImplementation) apiLoginUsernameAndPassword(w http.ResponseWriter, r *http.Request) {
	// Check rate limit
	if !a.checkRateLimit(w, r, "login") {
		return
	}

	email := req.GetStringTrimmed(r, "email")
	password := req.GetStringTrimmed(r, "password")

	response := a.LoginWithUsernameAndPassword(r.Context(), email, password, UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	if response.ErrorMessage != "" {
		api.Respond(w, r, api.Error(response.ErrorMessage))
		return
	}

	if a.useCookies {
		a.setAuthCookie(w, r, response.Token)
	}

	api.Respond(w, r, api.SuccessWithData(response.SuccessMessage, map[string]any{
		"token": response.Token,
	}))
}
