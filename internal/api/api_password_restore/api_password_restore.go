package api_password_restore

import (
	"context"
	"errors"
	"html"
	"log/slog"
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/auth/types"
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

// ApiPasswordRestoreWithAuth is a convenience wrapper that allows callers to
// pass a types.AuthSharedInterface (such as authImplementation) instead of
// manually wiring dependencies. It constructs the dependencies struct using the
// interface accessors and preserves the existing behaviour.
func ApiPasswordRestoreWithAuth(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
	passwordAuth, ok := a.(types.AuthPasswordInterface)
	if !ok {
		if logger := a.GetLogger(); logger != nil {
			logger.Error("password restore requires AuthPasswordInterface")
		}
		http.Error(w, "Internal server error. Please try again later", http.StatusInternalServerError)
		return
	}

	userFindByUsername := a.GetFuncUserFindByUsername()
	temporaryKeySet := a.GetFuncTemporaryKeySet()
	emailTemplatePasswordRestore := a.GetFuncEmailTemplatePasswordRestore()
	emailSend := a.GetFuncEmailSend()

	deps, err := NewDependencies(
		func(ctx context.Context, email, firstName, lastName string) (string, error) {
			if userFindByUsername == nil {
				return "", errors.New("UserFindByUsername is not configured")
			}
			return userFindByUsername(ctx, email, firstName, lastName, types.UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		},
		temporaryKeySet,
		0,
		func(ctx context.Context, userID, token string) string {
			if emailTemplatePasswordRestore == nil {
				return ""
			}
			return emailTemplatePasswordRestore(ctx, userID, passwordAuth.LinkPasswordReset(token), types.UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		},
		func(ctx context.Context, userID, subject, body string) error {
			if emailSend == nil {
				return errors.New("EmailSend is not configured")
			}
			return emailSend(ctx, userID, subject, body)
		},
		a.GetLogger(),
	)
	if err != nil {
		if logger := a.GetLogger(); logger != nil {
			logger.Error("password restore dependencies misconfigured",
				slog.String("error", err.Error()),
			)
		}
		http.Error(w, "Internal server error. Please try again later", http.StatusInternalServerError)
		return
	}

	ApiPasswordRestore(w, r, deps)
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
