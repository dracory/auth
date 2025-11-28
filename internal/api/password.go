package api

import (
	"context"
	"crypto/subtle"
	"errors"
	"html"
	"net/http"

	authtypes "github.com/dracory/auth/types"
	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

// ===================== PASSWORD RESTORE =====================

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

// ===================== PASSWORD RESET =====================

// PasswordResetDeps defines the dependencies required for the password reset
// flow (changing the user's password given a valid token).
type PasswordResetDeps struct {
	PasswordStrength *authtypes.PasswordStrengthConfig

	TemporaryKeyGet func(key string) (string, error)

	UserPasswordChange func(ctx context.Context, userID, password string) error
	LogoutUser         func(ctx context.Context, userID string) error
}

// PasswordResetErrorCode categorizes error sources in the password reset flow.
type PasswordResetErrorCode string

const (
	PasswordResetErrorCodeNone             PasswordResetErrorCode = ""
	PasswordResetErrorCodeValidation       PasswordResetErrorCode = "validation"
	PasswordResetErrorCodePasswordStrength PasswordResetErrorCode = "password_strength"
	PasswordResetErrorCodeTokenLookup      PasswordResetErrorCode = "token_lookup"
	PasswordResetErrorCodeTokenInvalid     PasswordResetErrorCode = "token_invalid"
	PasswordResetErrorCodePasswordChange   PasswordResetErrorCode = "password_change"
	PasswordResetErrorCodeLogout           PasswordResetErrorCode = "logout"
	PasswordResetErrorCodeInternal         PasswordResetErrorCode = "internal"
)

// PasswordResetError represents a structured error for password reset.
type PasswordResetError struct {
	Code    PasswordResetErrorCode
	Message string
	Err     error
	UserID  string
}

func (e *PasswordResetError) Error() string {
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

// PasswordResetResult represents a successful password reset.
type PasswordResetResult struct {
	SuccessMessage string
	Token          string
}

// PasswordReset encapsulates the core business logic for resetting a user's
// password based on a reset token. It does not log or write HTTP responses.
func PasswordReset(ctx context.Context, r *http.Request, deps PasswordResetDeps) (*PasswordResetResult, *PasswordResetError) {
	token := req.GetStringTrimmed(r, "token")
	password := req.GetStringTrimmed(r, "password")
	passwordConfirm := req.GetStringTrimmed(r, "password_confirm")

	if token == "" {
		return nil, &PasswordResetError{
			Code:    PasswordResetErrorCodeValidation,
			Message: "Token is required field",
		}
	}

	if password == "" {
		return nil, &PasswordResetError{
			Code:    PasswordResetErrorCodeValidation,
			Message: "Password is required field",
		}
	}

	if subtle.ConstantTimeCompare([]byte(password), []byte(passwordConfirm)) != 1 {
		return nil, &PasswordResetError{
			Code:    PasswordResetErrorCodeValidation,
			Message: "Passwords do not match",
		}
	}

	if deps.PasswordStrength != nil {
		if err := authutils.ValidatePasswordStrength(password, deps.PasswordStrength); err != nil {
			return nil, &PasswordResetError{
				Code: PasswordResetErrorCodePasswordStrength,
				Err:  err,
			}
		}
	}

	if deps.TemporaryKeyGet == nil {
		return nil, &PasswordResetError{
			Code:    PasswordResetErrorCodeTokenLookup,
			Message: "Link not valid or expired",
		}
	}

	userID, errToken := deps.TemporaryKeyGet(token)
	if errToken != nil {
		return nil, &PasswordResetError{
			Code:    PasswordResetErrorCodeTokenLookup,
			Message: "Link not valid or expired",
			Err:     errToken,
		}
	}

	if userID == "" {
		return nil, &PasswordResetError{
			Code:    PasswordResetErrorCodeTokenInvalid,
			Message: "Link not valid or expired",
		}
	}

	if deps.UserPasswordChange == nil {
		return nil, &PasswordResetError{
			Code:   PasswordResetErrorCodePasswordChange,
			Err:    errors.New("password change function is not configured"),
			UserID: userID,
		}
	}

	if errChange := deps.UserPasswordChange(ctx, userID, password); errChange != nil {
		return nil, &PasswordResetError{
			Code:   PasswordResetErrorCodePasswordChange,
			Err:    errChange,
			UserID: userID,
		}
	}

	if deps.LogoutUser == nil {
		return &PasswordResetResult{SuccessMessage: "login success", Token: token}, nil
	}

	if errLogout := deps.LogoutUser(ctx, userID); errLogout != nil {
		return nil, &PasswordResetError{
			Code:   PasswordResetErrorCodeLogout,
			Err:    errLogout,
			UserID: userID,
		}
	}

	return &PasswordResetResult{SuccessMessage: "login success", Token: token}, nil
}
