package auth

import (
	"context"
	"net/http"

	"github.com/dracory/api"
	apipwd "github.com/dracory/auth/internal/api"
	"github.com/dracory/req"
)

func (a authImplementation) apiPasswordRestore(w http.ResponseWriter, r *http.Request) {
	deps := apipwd.PasswordRestoreDeps{
		UserFindByUsername: func(ctx context.Context, email, firstName, lastName string) (string, error) {
			return a.funcUserFindByUsername(ctx, email, firstName, lastName, UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		},
		TemporaryKeySet: a.funcTemporaryKeySet,
		ExpiresSeconds:  int(DefaultPasswordResetExpiration.Seconds()),
		EmailTemplate: func(ctx context.Context, userID, token string) string {
			return a.funcEmailTemplatePasswordRestore(ctx, userID, a.LinkPasswordReset(token), UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		},
		EmailSend: func(ctx context.Context, userID, subject, body string) error {
			return a.funcEmailSend(ctx, userID, subject, body)
		},
	}

	result, perr := apipwd.PasswordRestore(r.Context(), r, deps)
	if perr != nil {
		logger := a.GetLogger()
		ip := req.GetIP(r)
		userAgent := r.UserAgent()

		switch perr.Code {
		case apipwd.PasswordRestoreErrorCodeValidation:
			api.Respond(w, r, api.Error(perr.Message))
			return
		case apipwd.PasswordRestoreErrorCodeUserLookup:
			logger.Error("password restore user lookup failed",
				"error", perr.Err,
				"email", perr.Email,
				"first_name", perr.FirstName,
				"last_name", perr.LastName,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_password_restore",
			)
			api.Respond(w, r, api.Error(perr.Message))
			return
		case apipwd.PasswordRestoreErrorCodeCodeGenerate:
			authErr := NewCodeGenerationError(perr.Err)
			logger.Error("password reset token generation failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"user_id", perr.UserID,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_password_restore",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		case apipwd.PasswordRestoreErrorCodeTokenStore:
			authErr := NewTokenStoreError(perr.Err)
			logger.Error("password reset token store failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"user_id", perr.UserID,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_password_restore",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		case apipwd.PasswordRestoreErrorCodeEmailSend:
			authErr := NewEmailSendError(perr.Err)
			logger.Error("password restore email send failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"user_id", perr.UserID,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_password_restore",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		default:
			authErr := NewInternalError(perr.Err)
			logger.Error("password restore internal error",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"user_id", perr.UserID,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_password_restore",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		}
	}

	api.Respond(w, r, api.Success(result.SuccessMessage))
}
