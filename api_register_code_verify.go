package auth

import (
	"context"
	"net/http"

	"github.com/dracory/api"
	apireg "github.com/dracory/auth/internal/api"
	"github.com/dracory/req"
)

func (a authImplementation) apiRegisterCodeVerify(w http.ResponseWriter, r *http.Request) {
	deps := apireg.RegisterCodeVerifyDeps{
		DisableRateLimit: a.disableRateLimit,
		TemporaryKeyGet:  a.funcTemporaryKeyGet,
		PasswordStrength: a.passwordStrength,
		Passwordless:     a.passwordless,
		PasswordlessUserRegister: func(ctx context.Context, email, firstName, lastName string) error {
			return a.passwordlessFuncUserRegister(ctx, email, firstName, lastName, UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		},
		UserRegister: func(ctx context.Context, email, password, firstName, lastName string) error {
			return a.funcUserRegister(ctx, email, password, firstName, lastName, UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		},
	}

	result, perr := apireg.RegisterCodeVerify(r.Context(), r, deps)
	if perr != nil {
		logger := a.GetLogger()
		ip := req.GetIP(r)
		userAgent := r.UserAgent()
		email := ""
		if result != nil {
			email = result.Email
		}

		switch perr.Code {
		case apireg.RegisterCodeVerifyErrorCodeValidation,
			apireg.RegisterCodeVerifyErrorCodeCodeExpired,
			apireg.RegisterCodeVerifyErrorCodeDeserialize:
			// Pure validation/data errors, no structured logging previously.
			api.Respond(w, r, api.Error(perr.Message))
			return
		case apireg.RegisterCodeVerifyErrorCodePasswordValidation:
			authErr := AuthError{
				Code:        ErrCodeValidationFailed,
				Message:     perr.Err.Error(),
				InternalErr: perr.Err,
			}
			logger.Error("password validation failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"email", email,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_register_code_verify",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		case apireg.RegisterCodeVerifyErrorCodeRegister:
			authErr := NewRegistrationError(perr.Err)
			logger.Error("user registration failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"email", email,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_register_code_verify",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		default:
			authErr := NewInternalError(perr.Err)
			logger.Error("registration code verify internal error",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"email", email,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_register_code_verify",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		}
	}

	a.authenticateViaUsername(w, r, result.Email, result.FirstName, result.LastName)
}
