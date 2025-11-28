package api_password_reset

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"

	"github.com/dracory/api"
	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

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

// ApiPasswordReset is the HTTP-level helper that wires request/response
// handling to the core PasswordReset business logic using the provided
// dependencies.
func ApiPasswordReset(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	result, perr := PasswordReset(r.Context(), r, deps)
	if perr != nil {
		switch perr.Code {
		case PasswordResetErrorCodeValidation,
			PasswordResetErrorCodeTokenLookup,
			PasswordResetErrorCodeTokenInvalid:
			api.Respond(w, r, api.Error(perr.Message))
			return
		case PasswordResetErrorCodePasswordStrength:
			// Preserve existing behavior: return the validation error string.
			if perr.Err != nil {
				api.Respond(w, r, api.Error(perr.Err.Error()))
			} else {
				api.Respond(w, r, api.Error("Password validation failed"))
			}
			return
		case PasswordResetErrorCodePasswordChange:
			// Map to the same user-facing message as NewPasswordResetError.
			api.Respond(w, r, api.Error("Password reset failed. Please try again later"))
			return
		case PasswordResetErrorCodeLogout:
			// Map to the same user-facing message as NewLogoutError.
			api.Respond(w, r, api.Error("Logout failed. Please try again later"))
			return
		default:
			api.Respond(w, r, api.Error("Internal server error. Please try again later"))
			return
		}
	}

	api.Respond(w, r, api.SuccessWithData(result.SuccessMessage, map[string]any{
		"token": result.Token,
	}))
}

// PasswordReset encapsulates the core business logic for resetting a user's
// password based on a reset token. It does not log or write HTTP responses.
func PasswordReset(ctx context.Context, r *http.Request, deps Dependencies) (*PasswordResetResult, *PasswordResetError) {
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
