package api_password_restore

import (
	"context"
	"errors"
	"html"
	"net/http"

	"github.com/dracory/api"
	authutils "github.com/dracory/auth/utils"
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
func ApiPasswordRestore(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	result, perr := PasswordRestore(r.Context(), r, deps)
	if perr != nil {
		switch perr.Code {
		case PasswordRestoreErrorCodeValidation:
			api.Respond(w, r, api.Error(perr.Message))
			return
		case PasswordRestoreErrorCodeUserLookup:
			api.Respond(w, r, api.Error(perr.Message))
			return
		case PasswordRestoreErrorCodeCodeGenerate,
			PasswordRestoreErrorCodeTokenStore,
			PasswordRestoreErrorCodeEmailSend,
			PasswordRestoreErrorCodeInternalError:
			api.Respond(w, r, api.Error("Internal server error. Please try again later"))
			return
		default:
			api.Respond(w, r, api.Error("Internal server error. Please try again later"))
			return
		}
	}

	api.Respond(w, r, api.Success(result.SuccessMessage))
}

// PasswordRestore encapsulates core business logic for issuing a password
// reset token and sending an email. It does not log or write HTTP responses.
func PasswordRestore(ctx context.Context, r *http.Request, deps Dependencies) (*PasswordRestoreResult, *PasswordRestoreError) {
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
