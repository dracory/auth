package api_password_restore

import (
	"context"
	"html"
	"log/slog"
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

// PasswordRestoreErrorCode categorizes error sources in the password restore
// flow.
type PasswordRestoreErrorCode string

// PasswordRestoreError represents a structured error for password restore.
type PasswordRestoreError struct {
	Code      PasswordRestoreErrorCode
	Message   string
	Err       error
	UserID    string
	Email     string
	FirstName string
	LastName  string
}

func (e *PasswordRestoreError) Error() string {
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

// PasswordRestoreResult represents a successful password restore operation.
type PasswordRestoreResult struct {
	SuccessMessage string
}

// ApiPasswordRestore is the HTTP-level helper that wires request/response
// handling to the core PasswordRestore business logic using the provided
// dependencies.
func ApiPasswordRestore(w http.ResponseWriter, r *http.Request, deps dependencies) {
	successMessage, errorMessage := passwordRestore(r.Context(), r, deps)

	if errorMessage != "" {
		api.Respond(w, r, api.Error(errorMessage))
		return
	}

	api.Respond(w, r, api.Success(successMessage))
}

// passwordRestore encapsulates core business logic for issuing a password
// reset token and sending an email. It does not log or write HTTP responses
// or perform dependency validation; dependencies are assumed to be valid.
func passwordRestore(ctx context.Context, r *http.Request, dependencies dependencies) (successMessage string, errorMessage string) {
	email := req.GetStringTrimmed(r, "email")
	firstName := html.EscapeString(req.GetStringTrimmed(r, "first_name"))
	lastName := html.EscapeString(req.GetStringTrimmed(r, "last_name"))

	if email == "" {
		return "", ERROR_EMAIL_REQUIRED
	}

	if msg := utils.ValidateEmailFormat(email); msg != "" {
		return "", msg
	}

	if firstName == "" {
		return "", ERROR_FIRST_NAME_REQUIRED
	}

	if lastName == "" {
		return "", ERROR_LAST_NAME_REQUIRED
	}

	userID, err := dependencies.UserFindByUsername(ctx, email, firstName, lastName)
	if err != nil {
		dependencies.Logger.Error(
			"user not found",
			slog.String("error", err.Error()),
			slog.String("email", email),
			slog.String("first_name", firstName),
			slog.String("last_name", lastName),
		)
		return "", ERROR_INTERNAL_SERVER
	}

	if userID == "" {
		return "", ERROR_USER_NOT_FOUND
	}

	resetToken, err := utils.GeneratePasswordResetToken()
	if err != nil {
		dependencies.Logger.Error(
			"failed to generate password reset token",
			slog.String("error", err.Error()),
			slog.String("email", email),
			slog.String("first_name", firstName),
			slog.String("last_name", lastName),
		)
		return "", ERROR_INTERNAL_SERVER
	}

	expires := dependencies.ExpiresSeconds
	if expires <= 0 {
		// default: one hour
		expires = 3600
	}

	if err := dependencies.TemporaryKeySet(resetToken, userID, expires); err != nil {
		dependencies.Logger.Error(
			"failed to store temporary key",
			slog.String("error", err.Error()),
			slog.String("email", email),
			slog.String("first_name", firstName),
			slog.String("last_name", lastName),
		)
		return "", ERROR_INTERNAL_SERVER
	}

	emailContent := dependencies.EmailTemplate(ctx, userID, resetToken)

	if errEmail := dependencies.EmailSend(ctx, userID, "Password Restore", emailContent); errEmail != nil {
		dependencies.Logger.Error(
			"failed to send email",
			slog.String("error", errEmail.Error()),
			slog.String("email", email),
			slog.String("first_name", firstName),
			slog.String("last_name", lastName),
		)
		return "", ERROR_INTERNAL_SERVER
	}

	return "Password reset link was sent to your e-mail", ""
}
