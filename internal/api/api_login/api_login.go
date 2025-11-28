package api_login

import (
	"context"
	"net/http"

	"github.com/dracory/api"
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

// ApiLogin is the HTTP-level handler that combines passwordless and
// username+password login flows behind a shared interface.
func ApiLogin(w http.ResponseWriter, r *http.Request, dependencies Dependencies) {
	if dependencies.Passwordless {
		result, err := loginPasswordless(r.Context(), r, dependencies.PasswordlessDependencies)
		if err != nil {
			switch err.Code {
			case LoginPasswordlessErrorCodeValidation:
				api.Respond(w, r, api.Error(err.Message))
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

	if dependencies.LoginWithUsernameAndPassword == nil {
		api.Respond(w, r, api.Error("Internal server error. Please try again later"))
		return
	}

	email := req.GetStringTrimmed(r, "email")
	password := req.GetStringTrimmed(r, "password")
	ip := req.GetIP(r)
	userAgent := r.UserAgent()

	successMessage, token, errMessage := dependencies.LoginWithUsernameAndPassword(r.Context(), email, password, ip, userAgent)
	if errMessage != "" {
		api.Respond(w, r, api.Error(errMessage))
		return
	}

	if dependencies.UseCookies && dependencies.SetAuthCookie != nil {
		dependencies.SetAuthCookie(w, r, token)
	}

	api.Respond(w, r, api.SuccessWithData(successMessage, map[string]any{
		"token": token,
	}))
}
