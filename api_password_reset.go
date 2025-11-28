package auth

import (
	"crypto/subtle"
	"log/slog"
	"net/http"

	"github.com/dracory/api"
	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

func (a Auth) apiPasswordReset(w http.ResponseWriter, r *http.Request) {
	// Check rate limit
	if !a.checkRateLimit(w, r, "password_reset") {
		return
	}

	// Check CSRF token
	if a.enableCSRFProtection && !a.funcCSRFTokenValidate(r) {
		api.Respond(w, r, api.Forbidden("Invalid CSRF token"))
		return
	}

	token := req.GetStringTrimmed(r, "token")
	password := req.GetStringTrimmed(r, "password")
	passwordConfirm := req.GetStringTrimmed(r, "password_confirm")

	if token == "" {
		api.Respond(w, r, api.Error("Token is required field"))
		return
	}

	if password == "" {
		api.Respond(w, r, api.Error("Password is required field"))
		return
	}

	if subtle.ConstantTimeCompare([]byte(password), []byte(passwordConfirm)) != 1 {
		api.Respond(w, r, api.Error("Passwords do not match"))
		return
	}

	if err := authutils.ValidatePasswordStrength(password, a.passwordStrength); err != nil {
		authErr := AuthError{
			Code:        ErrCodeValidationFailed,
			Message:     err.Error(),
			InternalErr: err,
		}
		logger := a.logger
		if logger == nil {
			logger = slog.Default()
		}
		logger.Error("password validation failed",
			"error", authErr.InternalErr,
			"error_code", authErr.Code,
			"ip", req.GetIP(r),
			"user_agent", r.UserAgent(),
			"endpoint", "api_password_reset",
		)
		api.Respond(w, r, api.Error(authErr.Message))
		return
	}

	userID, errToken := a.funcTemporaryKeyGet(token)

	if errToken != nil {
		api.Respond(w, r, api.Error("Link not valid or expired"))
		return
	}

	if userID == "" {
		api.Respond(w, r, api.Error("Link not valid or expired"))
		return
	}

	errPasswordChange := a.funcUserPasswordChange(r.Context(), userID, password, UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	if errPasswordChange != nil {
		authErr := NewPasswordResetError(errPasswordChange)
		logger := a.logger
		if logger == nil {
			logger = slog.Default()
		}
		logger.Error("password change failed",
			"error", authErr.InternalErr,
			"error_code", authErr.Code,
			"user_id", userID,
			"ip", req.GetIP(r),
			"user_agent", r.UserAgent(),
			"endpoint", "api_password_reset",
		)
		api.Respond(w, r, api.Error(authErr.Message))
		return
	}

	errLogout := a.funcUserLogout(r.Context(), userID, UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	if errLogout != nil {
		authErr := NewLogoutError(errLogout)
		logger := a.logger
		if logger == nil {
			logger = slog.Default()
		}
		logger.Error("session invalidation after password change failed",
			"error", authErr.InternalErr,
			"error_code", authErr.Code,
			"user_id", userID,
			"ip", req.GetIP(r),
			"user_agent", r.UserAgent(),
			"endpoint", "api_password_reset",
		)
		api.Respond(w, r, api.Error(authErr.Message))
		return
	}

	api.Respond(w, r, api.SuccessWithData("login success", map[string]interface{}{
		"token": token,
	}))
}
