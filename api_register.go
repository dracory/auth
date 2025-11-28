package auth

import (
	"encoding/json"
	"html"
	"log/slog"
	"net/http"

	"github.com/dracory/api"
	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

func (a Auth) apiRegister(w http.ResponseWriter, r *http.Request) {
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

func (a Auth) apiRegisterPasswordless(w http.ResponseWriter, r *http.Request) {
	// Check rate limit
	if !a.checkRateLimit(w, r, "register") {
		return
	}

	email := req.GetStringTrimmed(r, "email")
	first_name := html.EscapeString(req.GetStringTrimmed(r, "first_name"))
	last_name := html.EscapeString(req.GetStringTrimmed(r, "last_name"))

	if first_name == "" {
		api.Respond(w, r, api.Error("First name is required field"))
		return
	}

	if last_name == "" {
		api.Respond(w, r, api.Error("Last name is required field"))
		return
	}

	if email == "" {
		api.Respond(w, r, api.Error("Email is required field"))
		return
	}

	if msg := authutils.ValidateEmailFormat(email); msg != "" {
		api.Respond(w, r, api.Error(msg))
		return
	}

	verificationCode, errRandomFromGamma := authutils.GenerateVerificationCode(a.disableRateLimit)

	if errRandomFromGamma != nil {
		authErr := NewCodeGenerationError(errRandomFromGamma)
		logger := a.logger
		if logger == nil {
			logger = slog.Default()
		}
		logger.Error("registration code generation failed",
			"error", authErr.InternalErr,
			"error_code", authErr.Code,
			"email", email,
			"ip", req.GetIP(r),
			"user_agent", r.UserAgent(),
			"endpoint", "api_register_passwordless",
		)
		api.Respond(w, r, api.Error(authErr.Message))
		return
	}

	json, errJson := json.Marshal(map[string]string{
		"email":      email,
		"first_name": first_name,
		"last_name":  last_name,
	})

	if errJson != nil {
		authErr := NewSerializationError(errJson)
		logger := a.logger
		if logger == nil {
			logger = slog.Default()
		}
		logger.Error("registration data serialization failed",
			"error", authErr.InternalErr,
			"error_code", authErr.Code,
			"email", email,
			"ip", req.GetIP(r),
			"user_agent", r.UserAgent(),
			"endpoint", "api_register_passwordless",
		)
		api.Respond(w, r, api.Error(authErr.Message))
		return
	}

	errTempTokenSave := a.funcTemporaryKeySet(verificationCode, string(json), 3600)

	if errTempTokenSave != nil {
		authErr := NewTokenStoreError(errTempTokenSave)
		logger := a.logger
		if logger == nil {
			logger = slog.Default()
		}
		logger.Error("registration code token store failed",
			"error", authErr.InternalErr,
			"error_code", authErr.Code,
			"email", email,
			"ip", req.GetIP(r),
			"user_agent", r.UserAgent(),
			"endpoint", "api_register_passwordless",
		)
		api.Respond(w, r, api.Error(authErr.Message))
		return
	}

	emailContent := a.passwordlessFuncEmailTemplateRegisterCode(r.Context(), email, verificationCode, UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	errEmailSent := a.passwordlessFuncEmailSend(r.Context(), email, "Registration Code", emailContent)

	if errEmailSent != nil {
		authErr := NewEmailSendError(errEmailSent)
		logger := a.logger
		if logger == nil {
			logger = slog.Default()
		}
		logger.Error("registration code email send failed",
			"error", authErr.InternalErr,
			"error_code", authErr.Code,
			"email", email,
			"ip", req.GetIP(r),
			"user_agent", r.UserAgent(),
			"endpoint", "api_register_passwordless",
		)
		api.Respond(w, r, api.Error(authErr.Message))
		return
	}

	api.Respond(w, r, api.Success("Registration code was sent successfully"))

}

func (a Auth) apiRegisterUsernameAndPassword(w http.ResponseWriter, r *http.Request) {
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
