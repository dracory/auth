package auth

import (
	"log/slog"
	"net/http"

	"github.com/dracory/api"
	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

func (a Auth) apiLogin(w http.ResponseWriter, r *http.Request) {
	// Check CSRF token
	if a.enableCSRFProtection && !a.funcCSRFTokenValidate(r) {
		api.Respond(w, r, api.Forbidden("Invalid CSRF token"))
		return
	}

	if a.passwordless {
		a.apiLoginPasswordless(w, r)
	} else {
		a.apiLoginUsernameAndPassword(w, r)
	}
}

func (a Auth) apiLoginPasswordless(w http.ResponseWriter, r *http.Request) {
	// Check rate limit
	if !a.checkRateLimit(w, r, "login") {
		return
	}

	email := req.GetStringTrimmed(r, "email")

	if email == "" {
		api.Respond(w, r, api.Error("Email is required field"))
		return
	}

	if msg := authutils.ValidateEmailFormat(email); msg != "" {
		api.Respond(w, r, api.Error(msg))
		return
	}

	verificationCode := req.GetStringTrimmed(r, "verification_code")

	errTempTokenSave := a.funcTemporaryKeySet(verificationCode, email, 3600)

	if errTempTokenSave != nil {
		logger := a.logger
		if logger == nil {
			logger = slog.Default()
		}
		logger.Error("login code token store failed",
			"error", errTempTokenSave,
			"email", email,
			"ip", req.GetIP(r),
			"user_agent", r.UserAgent(),
			"endpoint", "api_login_passwordless",
		)
		api.Respond(w, r, api.Error("token store failed."))
		return
	}

	emailContent := a.passwordlessFuncEmailTemplateLoginCode(r.Context(), email, verificationCode, UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	errEmailSent := a.passwordlessFuncEmailSend(r.Context(), email, "Login Code", emailContent)

	if errEmailSent != nil {
		logger := a.logger
		if logger == nil {
			logger = slog.Default()
		}
		logger.Error("login code email send failed",
			"error", errEmailSent,
			"email", email,
			"ip", req.GetIP(r),
			"user_agent", r.UserAgent(),
			"endpoint", "api_login_passwordless",
		)
		api.Respond(w, r, api.Error("Login code failed to be send. Please try again later"))
		return
	}

	api.Respond(w, r, api.Success("Login code was sent successfully"))
}

func (a Auth) apiLoginUsernameAndPassword(w http.ResponseWriter, r *http.Request) {
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
		AuthCookieSet(w, r, response.Token)
	}

	api.Respond(w, r, api.SuccessWithData(response.SuccessMessage, map[string]any{
		"token": response.Token,
	}))
}
