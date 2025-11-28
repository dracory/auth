package auth

import (
	"html"
	"net/http"

	"github.com/dracory/api"
	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

func (a Auth) apiPasswordRestore(w http.ResponseWriter, r *http.Request) {
	// Check rate limit
	if !a.checkRateLimit(w, r, "password_restore") {
		return
	}

	email := req.GetStringTrimmed(r, "email")
	firstName := html.EscapeString(req.GetStringTrimmed(r, "first_name"))
	lastName := html.EscapeString(req.GetStringTrimmed(r, "last_name"))

	if email == "" {
		api.Respond(w, r, api.Error("Email is required field"))
		return
	}

	if msg := authutils.ValidateEmailFormat(email); msg != "" {
		api.Respond(w, r, api.Error(msg))
		return
	}

	if firstName == "" {
		api.Respond(w, r, api.Error("First name is required field"))
		return
	}

	if lastName == "" {
		api.Respond(w, r, api.Error("Last name is required field"))
		return
	}

	userID, err := a.funcUserFindByUsername(r.Context(), email, firstName, lastName, UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	if err != nil {
		logger := a.GetLogger()
		logger.Error("password restore user lookup failed",
			"error", err,
			"email", email,
			"first_name", firstName,
			"last_name", lastName,
			"ip", req.GetIP(r),
			"user_agent", r.UserAgent(),
			"endpoint", "api_password_restore",
		)
		api.Respond(w, r, api.Error("Internal server error"))
		return
	}

	if userID == "" {
		api.Respond(w, r, api.Error("User not found"))
		return
	}

	// if strings.ToLower(user.FirstName) != strings.ToLower(firstName) {
	// 	api.Respond(w, r, api.Error("First or last name not matching"))
	// 	return
	// }

	// if strings.ToLower(user.LastName) != strings.ToLower(lastName) {
	// 	api.Respond(w, r, api.Error("First or last name not matching"))
	// 	return
	// }

	token, errRandomFromGamma := authutils.GeneratePasswordResetToken()

	if errRandomFromGamma != nil {
		authErr := NewCodeGenerationError(errRandomFromGamma)
		logger := a.GetLogger()
		logger.Error("password reset token generation failed",
			"error", authErr.InternalErr,
			"error_code", authErr.Code,
			"user_id", userID,
			"ip", req.GetIP(r),
			"user_agent", r.UserAgent(),
			"endpoint", "api_password_restore",
		)
		api.Respond(w, r, api.Error(authErr.Message))
		return
	}

	errTempTokenSave := a.funcTemporaryKeySet(token, userID, int(DefaultPasswordResetExpiration.Seconds()))

	if errTempTokenSave != nil {
		authErr := NewTokenStoreError(errTempTokenSave)
		logger := a.GetLogger()
		logger.Error("password reset token store failed",
			"error", authErr.InternalErr,
			"error_code", authErr.Code,
			"user_id", userID,
			"ip", req.GetIP(r),
			"user_agent", r.UserAgent(),
			"endpoint", "api_password_restore",
		)
		api.Respond(w, r, api.Error(authErr.Message))
		return
	}

	emailContent := a.funcEmailTemplatePasswordRestore(r.Context(), userID, a.LinkPasswordReset(token), UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	errEmailSent := a.funcEmailSend(r.Context(), userID, "Password Restore", emailContent)

	if errEmailSent != nil {
		authErr := NewEmailSendError(errEmailSent)
		logger := a.GetLogger()
		logger.Error("password restore email send failed",
			"error", authErr.InternalErr,
			"error_code", authErr.Code,
			"user_id", userID,
			"ip", req.GetIP(r),
			"user_agent", r.UserAgent(),
			"endpoint", "api_password_restore",
		)
		api.Respond(w, r, api.Error(authErr.Message))
		return
	}

	api.Respond(w, r, api.Success("Password reset link was sent to your e-mail"))
}
