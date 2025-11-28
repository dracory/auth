package api_password_restore

import (
	"context"
	"errors"
	"html"
	"net/http"

	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

// PasswordRestoreDeps defines the dependencies required for the password
// restore flow (issuing a password reset link).
type PasswordRestoreDeps struct {
	UserFindByUsername func(ctx context.Context, email, firstName, lastName string) (userID string, err error)

	TemporaryKeySet func(key string, value string, expiresSeconds int) error
	ExpiresSeconds  int

	EmailTemplate func(ctx context.Context, userID, token string) string
	EmailSend     func(ctx context.Context, userID, subject, body string) error
}

// PasswordRestoreErrorCode categorizes error sources in the password restore
// flow.
type PasswordRestoreErrorCode string

const (
	PasswordRestoreErrorCodeNone          PasswordRestoreErrorCode = ""
	PasswordRestoreErrorCodeValidation    PasswordRestoreErrorCode = "validation"
	PasswordRestoreErrorCodeUserLookup    PasswordRestoreErrorCode = "user_lookup"
	PasswordRestoreErrorCodeCodeGenerate  PasswordRestoreErrorCode = "code_generation"
	PasswordRestoreErrorCodeTokenStore    PasswordRestoreErrorCode = "token_store"
	PasswordRestoreErrorCodeEmailSend     PasswordRestoreErrorCode = "email_send"
	PasswordRestoreErrorCodeInternalError PasswordRestoreErrorCode = "internal"
)

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

// PasswordRestore encapsulates core business logic for issuing a password
// reset token and sending an email. It does not log or write HTTP responses.
func PasswordRestore(ctx context.Context, r *http.Request, deps PasswordRestoreDeps) (*PasswordRestoreResult, *PasswordRestoreError) {
	email := req.GetStringTrimmed(r, "email")
	firstName := html.EscapeString(req.GetStringTrimmed(r, "first_name"))
	lastName := html.EscapeString(req.GetStringTrimmed(r, "last_name"))

	if email == "" {
		return nil, &PasswordRestoreError{
			Code:    PasswordRestoreErrorCodeValidation,
			Message: "Email is required field",
		}
	}

	if msg := authutils.ValidateEmailFormat(email); msg != "" {
		return nil, &PasswordRestoreError{
			Code:    PasswordRestoreErrorCodeValidation,
			Message: msg,
		}
	}

	if firstName == "" {
		return nil, &PasswordRestoreError{
			Code:    PasswordRestoreErrorCodeValidation,
			Message: "First name is required field",
		}
	}

	if lastName == "" {
		return nil, &PasswordRestoreError{
			Code:    PasswordRestoreErrorCodeValidation,
			Message: "Last name is required field",
		}
	}

	if deps.UserFindByUsername == nil {
		return nil, &PasswordRestoreError{
			Code:      PasswordRestoreErrorCodeUserLookup,
			Message:   "Internal server error",
			Email:     email,
			FirstName: firstName,
			LastName:  lastName,
		}
	}

	userID, errUser := deps.UserFindByUsername(ctx, email, firstName, lastName)
	if errUser != nil {
		return nil, &PasswordRestoreError{
			Code:      PasswordRestoreErrorCodeUserLookup,
			Message:   "Internal server error",
			Err:       errUser,
			Email:     email,
			FirstName: firstName,
			LastName:  lastName,
		}
	}

	if userID == "" {
		return nil, &PasswordRestoreError{
			Code:    PasswordRestoreErrorCodeValidation,
			Message: "User not found",
		}
	}

	resetToken, errToken := authutils.GeneratePasswordResetToken()
	if errToken != nil {
		return nil, &PasswordRestoreError{
			Code:   PasswordRestoreErrorCodeCodeGenerate,
			Err:    errToken,
			UserID: userID,
		}
	}

	if deps.TemporaryKeySet == nil {
		return nil, &PasswordRestoreError{
			Code:   PasswordRestoreErrorCodeTokenStore,
			Err:    errors.New("temporary key store is not configured"),
			UserID: userID,
		}
	}

	expires := deps.ExpiresSeconds
	if expires <= 0 {
		// default: one hour
		expires = 3600
	}

	if errTemp := deps.TemporaryKeySet(resetToken, userID, expires); errTemp != nil {
		return nil, &PasswordRestoreError{
			Code:   PasswordRestoreErrorCodeTokenStore,
			Err:    errTemp,
			UserID: userID,
		}
	}

	if deps.EmailTemplate == nil || deps.EmailSend == nil {
		return nil, &PasswordRestoreError{
			Code:   PasswordRestoreErrorCodeInternalError,
			Err:    errors.New("email template or sender is not configured"),
			UserID: userID,
		}
	}

	emailContent := deps.EmailTemplate(ctx, userID, resetToken)

	if errEmail := deps.EmailSend(ctx, userID, "Password Restore", emailContent); errEmail != nil {
		return nil, &PasswordRestoreError{
			Code:   PasswordRestoreErrorCodeEmailSend,
			Err:    errEmail,
			UserID: userID,
		}
	}

	return &PasswordRestoreResult{SuccessMessage: "Password reset link was sent to your e-mail"}, nil
}
