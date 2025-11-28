package auth

import (
	"context"
	"html"
	"net/http"

	"github.com/dracory/api"
	apilogin "github.com/dracory/auth/internal/api"
	apipwd "github.com/dracory/auth/internal/api"
	apireg "github.com/dracory/auth/internal/api"
	"github.com/dracory/auth/internal/api/api_login"
	"github.com/dracory/auth/internal/api/api_logout"
	"github.com/dracory/req"
	"github.com/dracory/str"
)

func (a authImplementation) apiLogin(w http.ResponseWriter, r *http.Request) {
	api_login.ApiLogin(w, r, api_login.Dependencies{
		Passwordless: a.passwordless,
		PasswordlessDependencies: api_login.LoginPasswordlessDeps{
			DisableRateLimit: a.disableRateLimit,
			TemporaryKeySet:  a.funcTemporaryKeySet,
			ExpiresSeconds:   int(DefaultVerificationCodeExpiration.Seconds()),
			EmailTemplate: func(ctx context.Context, email string, verificationCode string) string {
				return a.passwordlessFuncEmailTemplateLoginCode(ctx, email, verificationCode, UserAuthOptions{
					UserIp:    req.GetIP(r),
					UserAgent: r.UserAgent(),
				})
			},
			EmailSend: a.passwordlessFuncEmailSend,
		},
		LoginWithUsernameAndPassword: func(ctx context.Context, email, password, ip, userAgent string) (string, string, string) {
			response := a.LoginWithUsernameAndPassword(ctx, email, password, UserAuthOptions{
				UserIp:    ip,
				UserAgent: userAgent,
			})
			return response.SuccessMessage, response.Token, response.ErrorMessage
		},
		UseCookies: a.useCookies,
		SetAuthCookie: func(w http.ResponseWriter, r *http.Request, token string) {
			a.setAuthCookie(w, r, token)
		},
	})
}

func (a authImplementation) apiRegister(w http.ResponseWriter, r *http.Request) {
	if a.passwordless {
		a.apiRegisterPasswordless(w, r)
	} else {
		a.apiRegisterUsernameAndPassword(w, r)
	}
}

func (a authImplementation) apiRegisterPasswordless(w http.ResponseWriter, r *http.Request) {
	deps := apireg.RegisterPasswordlessInitDeps{
		DisableRateLimit: a.disableRateLimit,
		TemporaryKeySet:  a.funcTemporaryKeySet,
		ExpiresSeconds:   int(DefaultVerificationCodeExpiration.Seconds()),
		EmailTemplate: func(ctx context.Context, email string, verificationCode string) string {
			return a.passwordlessFuncEmailTemplateRegisterCode(ctx, email, verificationCode, UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		},
		EmailSend: func(ctx context.Context, email string, subject string, body string) error {
			return a.passwordlessFuncEmailSend(ctx, email, subject, body)
		},
	}

	result, perr := apireg.RegisterPasswordlessInit(r.Context(), r, deps)
	if perr != nil {
		email := req.GetStringTrimmed(r, "email")
		logger := a.GetLogger()
		ip := req.GetIP(r)
		userAgent := r.UserAgent()

		switch perr.Code {
		case apireg.RegisterPasswordlessInitErrorCodeValidation:
			api.Respond(w, r, api.Error(perr.Message))
			return
		case apireg.RegisterPasswordlessInitErrorCodeCodeGeneration:
			authErr := NewCodeGenerationError(perr.Err)
			logger.Error("registration code generation failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"email", email,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_register_passwordless",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		case apireg.RegisterPasswordlessInitErrorCodeSerialization:
			authErr := NewSerializationError(perr.Err)
			logger.Error("registration data serialization failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"email", email,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_register_passwordless",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		case apireg.RegisterPasswordlessInitErrorCodeTokenStore:
			authErr := NewTokenStoreError(perr.Err)
			logger.Error("registration code token store failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"email", email,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_register_passwordless",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		case apireg.RegisterPasswordlessInitErrorCodeEmailSend:
			authErr := NewEmailSendError(perr.Err)
			logger.Error("registration code email send failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"email", email,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_register_passwordless",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		default:
			authErr := NewInternalError(perr.Err)
			logger.Error("registration init internal error",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"email", email,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_register_passwordless",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		}
	}

	api.Respond(w, r, api.Success(result.SuccessMessage))
}

func (a authImplementation) apiRegisterUsernameAndPassword(w http.ResponseWriter, r *http.Request) {
	email := req.GetStringTrimmed(r, "email")
	password := req.GetStringTrimmed(r, "password")
	first_name := html.EscapeString(req.GetStringTrimmed(r, "first_name"))
	last_name := html.EscapeString(req.GetStringTrimmed(r, "last_name"))

	response := a.RegisterWithUsernameAndPassword(r.Context(), email, password, first_name, last_name, UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	if response.ErrorMessage != "" {
		api.Respond(w, r, api.Error(response.ErrorMessage))
		return
	}

	api.Respond(w, r, api.Success(response.SuccessMessage))
}

func (a authImplementation) apiLogout(w http.ResponseWriter, r *http.Request) {
	api_logout.ApiLogout(w, r, api_logout.Dependencies{
		UserFromToken: func(ctx context.Context, token string) (string, error) {
			return a.funcUserFindByAuthToken(ctx, token, UserAuthOptions{
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
		UseCookies: a.useCookies,
		AuthTokenRetrieve: func(r *http.Request, useCookies bool) string {
			return AuthTokenRetrieve(r, useCookies)
		},
		RemoveAuthCookie: func(w http.ResponseWriter, r *http.Request) {
			a.removeAuthCookie(w, r)
		},
	})
}

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

func (a authImplementation) apiLoginCodeVerify(w http.ResponseWriter, r *http.Request) {
	deps := apilogin.LoginCodeVerifyDeps{
		DisableRateLimit: a.disableRateLimit,
		TemporaryKeyGet:  a.funcTemporaryKeyGet,
	}

	result, perr := apilogin.LoginCodeVerify(r.Context(), r, deps)
	if perr != nil {
		switch perr.Code {
		case apilogin.LoginCodeVerifyErrorCodeValidation,
			apilogin.LoginCodeVerifyErrorCodeCodeExpired:
			api.Respond(w, r, api.Error(perr.Message))
			return
		default:
			api.Respond(w, r, api.Error("Verification code has expired"))
			return
		}
	}

	a.authenticateViaUsername(w, r, result.Email, "", "")
}

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

func (a authImplementation) authenticateViaUsername(w http.ResponseWriter, r *http.Request, username string, firstName string, lastName string) {
	var userID string
	var errUser error
	if a.passwordless {
		userID, errUser = a.passwordlessFuncUserFindByEmail(r.Context(), username, UserAuthOptions{
			UserIp:    req.GetIP(r),
			UserAgent: r.UserAgent(),
		})
	} else {
		userID, errUser = a.funcUserFindByUsername(r.Context(), username, firstName, lastName, UserAuthOptions{
			UserIp:    req.GetIP(r),
			UserAgent: r.UserAgent(),
		})
	}

	if errUser != nil {
		api.Respond(w, r, api.Error("Invalid credentials"))
		return
	}

	if userID == "" {
		api.Respond(w, r, api.Error("Invalid credentials"))
		return
	}

	token, errRandomFromGamma := str.RandomFromGamma(32, "BCDFGHJKLMNPQRSTVXYZ")

	if errRandomFromGamma != nil {
		authErr := NewCodeGenerationError(errRandomFromGamma)
		logger := a.GetLogger()
		logger.Error("login auth token generation failed",
			"error", authErr.InternalErr,
			"error_code", authErr.Code,
			"user_id", userID,
			"ip", req.GetIP(r),
			"user_agent", r.UserAgent(),
			"endpoint", "authenticate_via_username",
		)
		api.Respond(w, r, api.Error(authErr.Message))
		return
	}

	errSession := a.funcUserStoreAuthToken(r.Context(), token, userID, UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	if errSession != nil {
		authErr := NewTokenStoreError(errSession)
		logger := a.GetLogger()
		logger.Error("login auth token store failed",
			"error", authErr.InternalErr,
			"error_code", authErr.Code,
			"user_id", userID,
			"ip", req.GetIP(r),
			"user_agent", r.UserAgent(),
			"endpoint", "authenticate_via_username",
		)
		api.Respond(w, r, api.Error(authErr.Message))
		return
	}

	if a.useCookies {
		a.setAuthCookie(w, r, token)
	}

	api.Respond(w, r, api.SuccessWithData("login success", map[string]any{
		"token": token,
	}))
}
