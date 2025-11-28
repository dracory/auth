package api_login

// import (
// 	"log/slog"
// 	"net/http"

// 	"github.com/dracory/api"
// 	"github.com/dracory/auth/types"
// 	authutils "github.com/dracory/auth/utils"
// 	"github.com/dracory/req"
// )

// func ApiLogin(w http.ResponseWriter, r *http.Request, dependencies Dependencies) {
// 	if dependencies.Passwordless {
// 		apiLoginPasswordless(w, r)
// 	} else {
// 		apiLoginUsernameAndPassword(w, r)
// 	}
// }

// func apiLoginPasswordless(w http.ResponseWriter, r *http.Request) {
// 	email := req.GetStringTrimmed(r, "email")

// 	if email == "" {
// 		api.Respond(w, r, api.Error("Email is required field"))
// 		return
// 	}

// 	if msg := authutils.ValidateEmailFormat(email); msg != "" {
// 		api.Respond(w, r, api.Error(msg))
// 		return
// 	}

// 	// Server generates code, not client
// 	verificationCode, err := authutils.GenerateVerificationCode(a.disableRateLimit)
// 	if err != nil {
// 		authErr := NewCodeGenerationError(err)
// 		logger := a.logger
// 		if logger == nil {
// 			logger = slog.Default()
// 		}
// 		logger.Error("login code generation failed",
// 			"error", authErr.InternalErr,
// 			"error_code", authErr.Code,
// 			"email", email,
// 			"ip", req.GetIP(r),
// 			"user_agent", r.UserAgent(),
// 			"endpoint", "api_login_passwordless",
// 		)
// 		api.Respond(w, r, api.Error(authErr.Message))
// 		return
// 	}

// 	errTempTokenSave := a.funcTemporaryKeySet(verificationCode, email, int(DefaultVerificationCodeExpiration.Seconds()))

// 	if errTempTokenSave != nil {
// 		authErr := NewTokenStoreError(errTempTokenSave)
// 		logger := a.logger
// 		if logger == nil {
// 			logger = slog.Default()
// 		}
// 		logger.Error("login code token store failed",
// 			"error", authErr.InternalErr,
// 			"error_code", authErr.Code,
// 			"email", email,
// 			"ip", req.GetIP(r),
// 			"user_agent", r.UserAgent(),
// 			"endpoint", "api_login_passwordless",
// 		)
// 		api.Respond(w, r, api.Error(authErr.Message))
// 		return
// 	}

// 	emailContent := a.passwordlessFuncEmailTemplateLoginCode(r.Context(), email, verificationCode, UserAuthOptions{
// 		UserIp:    req.GetIP(r),
// 		UserAgent: r.UserAgent(),
// 	})

// 	errEmailSent := a.passwordlessFuncEmailSend(r.Context(), email, "Login Code", emailContent)

// 	if errEmailSent != nil {
// 		authErr := NewEmailSendError(errEmailSent)
// 		logger := a.logger
// 		if logger == nil {
// 			logger = slog.Default()
// 		}
// 		logger.Error("login code email send failed",
// 			"error", authErr.InternalErr,
// 			"error_code", authErr.Code,
// 			"email", email,
// 			"ip", req.GetIP(r),
// 			"user_agent", r.UserAgent(),
// 			"endpoint", "api_login_passwordless",
// 		)
// 		api.Respond(w, r, api.Error(authErr.Message))
// 		return
// 	}

// 	api.Respond(w, r, api.Success("Login code was sent successfully"))

// }

// func apiLoginUsernameAndPassword(w http.ResponseWriter, r *http.Request) {
// 	email := req.GetStringTrimmed(r, "email")
// 	password := req.GetStringTrimmed(r, "password")

// 	response := LoginWithUsernameAndPassword(r.Context(), email, password, types.UserAuthOptions{
// 		UserIp:    req.GetIP(r),
// 		UserAgent: r.UserAgent(),
// 	})

// 	if response.ErrorMessage != "" {
// 		api.Respond(w, r, api.Error(response.ErrorMessage))
// 		return
// 	}

// 	if a.useCookies {
// 		a.setAuthCookie(w, r, response.Token)
// 	}

// 	api.Respond(w, r, api.SuccessWithData(response.SuccessMessage, map[string]any{
// 		"token": response.Token,
// 	}))
// }
