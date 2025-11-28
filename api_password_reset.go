package auth

import (
	"context"
	"net/http"

	"github.com/dracory/api"
	apipwd "github.com/dracory/auth/internal/api"
	"github.com/dracory/req"
)

func (a authImplementation) apiPasswordReset(w http.ResponseWriter, r *http.Request) {
	deps := apipwd.PasswordResetDeps{
		PasswordStrength: a.passwordStrength,
		TemporaryKeyGet:  a.funcTemporaryKeyGet,
		UserPasswordChange: func(ctx context.Context, userID, password string) error {
			return a.funcUserPasswordChange(ctx, userID, password, UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		},
		LogoutUser: func(ctx context.Context, userID string) error {
			return a.funcUserLogout(ctx, userID, UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		},
	}

	result, perr := apipwd.PasswordReset(r.Context(), r, deps)
	if perr != nil {
		logger := a.GetLogger()
		ip := req.GetIP(r)
		userAgent := r.UserAgent()

		switch perr.Code {
		case apipwd.PasswordResetErrorCodeValidation,
			apipwd.PasswordResetErrorCodeTokenLookup,
			apipwd.PasswordResetErrorCodeTokenInvalid:
			api.Respond(w, r, api.Error(perr.Message))
			return
		case apipwd.PasswordResetErrorCodePasswordStrength:
			authErr := AuthError{
				Code:        ErrCodeValidationFailed,
				Message:     perr.Err.Error(),
				InternalErr: perr.Err,
			}
			logger.Error("password validation failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_password_reset",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		case apipwd.PasswordResetErrorCodePasswordChange:
			authErr := NewPasswordResetError(perr.Err)
			logger.Error("password change failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"user_id", perr.UserID,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_password_reset",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		case apipwd.PasswordResetErrorCodeLogout:
			authErr := NewLogoutError(perr.Err)
			logger.Error("session invalidation after password change failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"user_id", perr.UserID,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_password_reset",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		default:
			authErr := NewInternalError(perr.Err)
			logger.Error("password reset internal error",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"user_id", perr.UserID,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_password_reset",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		}
	}

	api.Respond(w, r, api.SuccessWithData(result.SuccessMessage, map[string]any{
		"token": result.Token,
	}))
}
