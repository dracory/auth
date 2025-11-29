package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dracory/auth/internal/api/api_authenticate_via_username"
	"github.com/dracory/auth/internal/api/api_login"
	"github.com/dracory/auth/internal/api/api_login_code_verify"
	"github.com/dracory/auth/internal/api/api_logout"
	"github.com/dracory/auth/internal/api/api_password_reset"
	"github.com/dracory/auth/internal/api/api_password_restore"
	"github.com/dracory/auth/internal/api/api_register"
	"github.com/dracory/auth/internal/api/api_register_code_verify"
	"github.com/dracory/auth/types"
	"github.com/dracory/req"
)

func (a authImplementation) apiLogin(w http.ResponseWriter, r *http.Request) {
	dependencies := api_login.Dependencies{
		Passwordless: a.passwordless,
		PasswordlessDependencies: api_login.LoginPasswordlessDeps{
			DisableRateLimit: a.disableRateLimit,
			TemporaryKeySet:  a.funcTemporaryKeySet,
			ExpiresSeconds:   int(DefaultVerificationCodeExpiration.Seconds()),
			EmailTemplate: func(ctx context.Context, email string, verificationCode string) string {
				return a.passwordlessFuncEmailTemplateLoginCode(ctx, email, verificationCode, types.UserAuthOptions{
					UserIp:    req.GetIP(r),
					UserAgent: r.UserAgent(),
				})
			},
			EmailSend: a.passwordlessFuncEmailSend,
		},
		LoginWithUsernameAndPassword: func(ctx context.Context, email, password, ip, userAgent string) (string, string, string) {
			response := a.LoginWithUsernameAndPassword(ctx, email, password, types.UserAuthOptions{
				UserIp:    ip,
				UserAgent: userAgent,
			})
			return response.SuccessMessage, response.Token, response.ErrorMessage
		},
		UseCookies: a.useCookies,
		SetAuthCookie: func(w http.ResponseWriter, r *http.Request, token string) {
			a.setAuthCookie(w, r, token)
		},
	}

	api_login.ApiLogin(w, r, dependencies)
}

func (a authImplementation) apiRegister(w http.ResponseWriter, r *http.Request) {
	dependencies := api_register.Dependencies{
		Passwordless: a.passwordless,
		RegisterPasswordlessInitDependencies: api_register.RegisterPasswordlessInitDependencies{
			DisableRateLimit: a.disableRateLimit,
			TemporaryKeySet:  a.funcTemporaryKeySet,
			ExpiresSeconds:   int(DefaultVerificationCodeExpiration.Seconds()),
			EmailTemplate: func(ctx context.Context, email string, verificationCode string) string {
				return a.passwordlessFuncEmailTemplateRegisterCode(ctx, email, verificationCode, types.UserAuthOptions{
					UserIp:    req.GetIP(r),
					UserAgent: r.UserAgent(),
				})
			},
			EmailSend: func(ctx context.Context, email string, subject string, body string) error {
				return a.passwordlessFuncEmailSend(ctx, email, subject, body)
			},
		},
		RegisterWithUsernameAndPassword: func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (string, string) {
			resp := a.RegisterWithUsernameAndPassword(ctx, email, password, firstName, lastName, types.UserAuthOptions{
				UserIp:    ip,
				UserAgent: userAgent,
			})
			return resp.SuccessMessage, resp.ErrorMessage
		},
	}

	api_register.ApiRegister(w, r, dependencies)
}

func (a authImplementation) apiLogout(w http.ResponseWriter, r *http.Request) {
	api_logout.ApiLogoutWithAuth(w, r, &a)
}

func (a authImplementation) apiPasswordRestore(w http.ResponseWriter, r *http.Request) {
	dependencies, err := api_password_restore.NewDependencies(
		func(ctx context.Context, email, firstName, lastName string) (string, error) {
			return a.funcUserFindByUsername(ctx, email, firstName, lastName, types.UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		},
		a.funcTemporaryKeySet,
		int(DefaultPasswordResetExpiration.Seconds()),
		func(ctx context.Context, userID, token string) string {
			return a.funcEmailTemplatePasswordRestore(ctx, userID, a.LinkPasswordReset(token), types.UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		},
		func(ctx context.Context, userID, subject, body string) error {
			return a.funcEmailSend(ctx, userID, subject, body)
		},
		a.GetLogger(),
	)
	if err != nil {
		a.GetLogger().Error("password restore dependencies misconfigured",
			slog.String("error", err.Error()),
		)
		http.Error(w, "Internal server error. Please try again later", http.StatusInternalServerError)
		return
	}

	api_password_restore.ApiPasswordRestore(w, r, dependencies)
}

func (a authImplementation) apiPasswordReset(w http.ResponseWriter, r *http.Request) {
	api_password_reset.ApiPasswordResetWithAuth(w, r, &a)
}

func (a authImplementation) apiLoginCodeVerify(w http.ResponseWriter, r *http.Request) {
	api_login_code_verify.ApiLoginCodeVerifyWithAuth(w, r, &a)
}

func (a authImplementation) apiRegisterCodeVerify(w http.ResponseWriter, r *http.Request) {
	api_register_code_verify.ApiRegisterCodeVerifyWithAuth(w, r, &a)
}

func (a authImplementation) authenticateViaUsername(w http.ResponseWriter, r *http.Request, username string, firstName string, lastName string) {
	api_authenticate_via_username.ApiAuthenticateViaUsernameWithAuth(w, r, username, firstName, lastName, &a)
}
