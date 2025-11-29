package testutils

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dracory/auth/types"
)

func NewAuthSharedForTest() types.AuthSharedInterface {
	a := &authSharedTest{}
	a.SetEndpoint("http://localhost/auth")
	a.SetRedirectOnSuccess("http://localhost/dashboard")
	a.SetLayout(func(content string) string { return content })
	a.SetLogger(slog.Default())
	return a
}

type authSharedTest struct {
	endpoint                              string
	layout                                func(content string) string
	logger                                *slog.Logger
	registration                          bool
	passwordless                          bool
	verification                          bool
	temporaryKeyGet                       func(key string) (string, error)
	temporaryKeySet                       func(key string, value string, expiresSeconds int) error
	funcUserFindByAuthToken               func(ctx context.Context, token string, options types.UserAuthOptions) (string, error)
	redirectOnSuccess                     string
	loginURL                              string
	useCookies                            bool
	disableRateLimit                      bool
	passwordStrength                      *types.PasswordStrengthConfig
	passwordlessUserRegister              func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) error
	funcUserRegister                      func(ctx context.Context, username, password, firstName, lastName string, options types.UserAuthOptions) error
	funcUserPasswordChange                func(ctx context.Context, userID, password string, options types.UserAuthOptions) error
	funcUserLogout                        func(ctx context.Context, userID string, options types.UserAuthOptions) error
	passwordlessUserFindByEmail           func(ctx context.Context, email string, options types.UserAuthOptions) (string, error)
	funcUserFindByUsername                func(ctx context.Context, username, firstName, lastName string, options types.UserAuthOptions) (string, error)
	funcUserStoreAuthToken                func(ctx context.Context, token, userID string, options types.UserAuthOptions) error
	emailTemplatePasswordRestore          func(ctx context.Context, userID string, passwordRestoreLink string, options types.UserAuthOptions) string
	emailSend                             func(ctx context.Context, userID, emailSubject, emailBody string) error
	passwordlessEmailTemplateLoginCode    func(ctx context.Context, email string, passwordRestoreLink string, options types.UserAuthOptions) string
	passwordlessEmailTemplateRegisterCode func(ctx context.Context, email string, passwordRestoreLink string, options types.UserAuthOptions) string
	passwordlessEmailSend                 func(ctx context.Context, email string, emailSubject, emailBody string) error
}

func (a *authSharedTest) Router() *http.ServeMux { return http.NewServeMux() }

func (a *authSharedTest) IsRegistrationEnabled() bool { return a.registration }

func (a *authSharedTest) IsPasswordless() bool { return a.passwordless }

func (a *authSharedTest) IsVerificationEnabled() bool { return a.verification }

func (a *authSharedTest) WebAuthOrRedirectMiddleware(next http.Handler) http.Handler { return next }

func (a *authSharedTest) ApiAuthOrErrorMiddleware(next http.Handler) http.Handler { return next }

func (a *authSharedTest) WebAppendUserIdIfExistsMiddleware(next http.Handler) http.Handler {
	return next
}

func (a *authSharedTest) GetCurrentUserID(r *http.Request) string { return "" }

func (a *authSharedTest) GetUseCookies() bool { return a.useCookies }

func (a *authSharedTest) SetUseCookies(useCookies bool) { a.useCookies = useCookies }

func (a *authSharedTest) GetFuncTemporaryKeyGet() func(key string) (string, error) {
	return a.temporaryKeyGet
}

func (a *authSharedTest) GetFuncTemporaryKeySet() func(key string, value string, expiresSeconds int) error {
	return a.temporaryKeySet
}

func (a *authSharedTest) GetFuncUserFindByAuthToken() func(ctx context.Context, token string, options types.UserAuthOptions) (string, error) {
	if a.funcUserFindByAuthToken != nil {
		return a.funcUserFindByAuthToken
	}
	return func(ctx context.Context, token string, options types.UserAuthOptions) (string, error) {
		return "", nil
	}
}

func (a *authSharedTest) SetFuncUserFindByAuthToken(fn func(ctx context.Context, token string, options types.UserAuthOptions) (string, error)) {
	a.funcUserFindByAuthToken = fn
}

func (a *authSharedTest) SetFuncTemporaryKeyGet(fn func(key string) (string, error)) {
	a.temporaryKeyGet = fn
}

func (a *authSharedTest) SetFuncTemporaryKeySet(fn func(key string, value string, expiresSeconds int) error) {
	a.temporaryKeySet = fn
}

func (a *authSharedTest) GetDisableRateLimit() bool { return a.disableRateLimit }

func (a *authSharedTest) SetDisableRateLimit(disable bool) { a.disableRateLimit = disable }

func (a *authSharedTest) GetPasswordStrength() *types.PasswordStrengthConfig {
	return a.passwordStrength
}

func (a *authSharedTest) SetPasswordStrength(cfg *types.PasswordStrengthConfig) {
	a.passwordStrength = cfg
}

func (a *authSharedTest) GetPasswordlessUserRegister() func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) error {
	return a.passwordlessUserRegister
}

func (a *authSharedTest) SetPasswordlessUserRegister(fn func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) error) {
	a.passwordlessUserRegister = fn
}

func (a *authSharedTest) GetPasswordlessFuncEmailTemplateLoginCode() func(ctx context.Context, email string, passwordRestoreLink string, options types.UserAuthOptions) string {
	return a.passwordlessEmailTemplateLoginCode
}

func (a *authSharedTest) SetPasswordlessFuncEmailTemplateLoginCode(fn func(ctx context.Context, email string, passwordRestoreLink string, options types.UserAuthOptions) string) {
	a.passwordlessEmailTemplateLoginCode = fn
}

func (a *authSharedTest) GetPasswordlessFuncEmailTemplateRegisterCode() func(ctx context.Context, email string, passwordRestoreLink string, options types.UserAuthOptions) string {
	return a.passwordlessEmailTemplateRegisterCode
}

func (a *authSharedTest) SetPasswordlessFuncEmailTemplateRegisterCode(fn func(ctx context.Context, email string, passwordRestoreLink string, options types.UserAuthOptions) string) {
	a.passwordlessEmailTemplateRegisterCode = fn
}

func (a *authSharedTest) GetPasswordlessFuncEmailSend() func(ctx context.Context, email string, emailSubject, emailBody string) error {
	return a.passwordlessEmailSend
}

func (a *authSharedTest) SetPasswordlessFuncEmailSend(fn func(ctx context.Context, email string, emailSubject, emailBody string) error) {
	a.passwordlessEmailSend = fn
}

func (a *authSharedTest) RegisterUserWithPassword(ctx context.Context, email, password, firstName, lastName string, options types.UserAuthOptions) (string, string, string) {
	// Default test double: no-op registration.
	return "", "", ""
}

func (a *authSharedTest) LoginUserWithPassword(ctx context.Context, email, password string, options types.UserAuthOptions) (string, string, string) {
	// Default test double: no-op login.
	return "", "", ""
}

func (a *authSharedTest) GetFuncUserRegister() func(ctx context.Context, username, password, firstName, lastName string, options types.UserAuthOptions) error {
	return a.funcUserRegister
}

func (a *authSharedTest) SetFuncUserRegister(fn func(ctx context.Context, username, password, firstName, lastName string, options types.UserAuthOptions) error) {
	a.funcUserRegister = fn
}

func (a *authSharedTest) GetFuncUserPasswordChange() func(ctx context.Context, userID, password string, options types.UserAuthOptions) error {
	return a.funcUserPasswordChange
}

func (a *authSharedTest) SetFuncUserPasswordChange(fn func(ctx context.Context, userID, password string, options types.UserAuthOptions) error) {
	a.funcUserPasswordChange = fn
}

func (a *authSharedTest) GetFuncUserLogout() func(ctx context.Context, userID string, options types.UserAuthOptions) error {
	return a.funcUserLogout
}

func (a *authSharedTest) SetFuncUserLogout(fn func(ctx context.Context, userID string, options types.UserAuthOptions) error) {
	a.funcUserLogout = fn
}

func (a *authSharedTest) GetPasswordlessUserFindByEmail() func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
	return a.passwordlessUserFindByEmail
}

func (a *authSharedTest) SetPasswordlessUserFindByEmail(fn func(ctx context.Context, email string, options types.UserAuthOptions) (string, error)) {
	a.passwordlessUserFindByEmail = fn
}

func (a *authSharedTest) GetFuncEmailTemplatePasswordRestore() func(ctx context.Context, userID string, passwordRestoreLink string, options types.UserAuthOptions) string {
	return a.emailTemplatePasswordRestore
}

func (a *authSharedTest) SetFuncEmailTemplatePasswordRestore(fn func(ctx context.Context, userID string, passwordRestoreLink string, options types.UserAuthOptions) string) {
	a.emailTemplatePasswordRestore = fn
}

func (a *authSharedTest) GetFuncEmailSend() func(ctx context.Context, userID, emailSubject, emailBody string) error {
	return a.emailSend
}

func (a *authSharedTest) SetFuncEmailSend(fn func(ctx context.Context, userID, emailSubject, emailBody string) error) {
	a.emailSend = fn
}

func (a *authSharedTest) GetFuncUserFindByUsername() func(ctx context.Context, username, firstName, lastName string, options types.UserAuthOptions) (string, error) {
	return a.funcUserFindByUsername
}

func (a *authSharedTest) SetFuncUserFindByUsername(fn func(ctx context.Context, username, firstName, lastName string, options types.UserAuthOptions) (string, error)) {
	a.funcUserFindByUsername = fn
}

func (a *authSharedTest) GetFuncUserStoreAuthToken() func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
	return a.funcUserStoreAuthToken
}

func (a *authSharedTest) SetFuncUserStoreAuthToken(fn func(ctx context.Context, token, userID string, options types.UserAuthOptions) error) {
	a.funcUserStoreAuthToken = fn
}

func (a *authSharedTest) SetAuthCookie(w http.ResponseWriter, r *http.Request, token string) {
	// test double: no-op
}

func (a *authSharedTest) RemoveAuthCookie(w http.ResponseWriter, r *http.Request) {
	// test double: no-op
}

func (a *authSharedTest) TemporaryKeyGet(token string) (string, error) { return "", nil }

func (a *authSharedTest) AuthenticateViaUsername(w http.ResponseWriter, r *http.Request, email, firstName, lastName string) {
	// test double: no-op or can be extended per test needs
}

func (a *authSharedTest) LinkLogin() string { return a.loginURL }

func (a *authSharedTest) LinkLogout() string { return "" }

func (a *authSharedTest) LinkRegister() string { return "" }

func (a *authSharedTest) LinkRegisterCodeVerify() string { return "" }

func (a *authSharedTest) LinkRedirectOnSuccess() string { return a.redirectOnSuccess }

func (a *authSharedTest) LinkApiLogin() string { return "" }

func (a *authSharedTest) LinkApiLogout() string { return "" }

func (a *authSharedTest) LinkApiRegister() string { return "" }

func (a *authSharedTest) LinkApiRegisterCodeVerify() string { return "" }

func (a *authSharedTest) GetEndpoint() string { return a.endpoint }

func (a *authSharedTest) SetEndpoint(endpoint string) { a.endpoint = endpoint }

func (a *authSharedTest) GetLogger() *slog.Logger {
	if a.logger != nil {
		return a.logger
	}
	return slog.Default()
}

func (a *authSharedTest) SetLogger(logger *slog.Logger) { a.logger = logger }

func (a *authSharedTest) GetLayout() func(content string) string {
	if a.layout != nil {
		return a.layout
	}
	return func(content string) string { return content }
}

func (a *authSharedTest) SetLayout(layout func(content string) string) { a.layout = layout }

func (a *authSharedTest) SetRedirectOnSuccess(url string) { a.redirectOnSuccess = url }

// Test helpers to configure additional flags on the shared auth test double.
func SetRegistrationForTest(a types.AuthSharedInterface, registration bool) {
	if v, ok := a.(*authSharedTest); ok {
		v.registration = registration
	}
}

func SetPasswordlessForTest(a types.AuthSharedInterface, passwordless bool) {
	if v, ok := a.(*authSharedTest); ok {
		v.passwordless = passwordless
	}
}

func SetVerificationForTest(a types.AuthSharedInterface, verification bool) {
	if v, ok := a.(*authSharedTest); ok {
		v.verification = verification
	}
}

// SetFuncUserFindByAuthTokenForTest allows tests to control auth-token lookup behaviour.
func SetFuncUserFindByAuthTokenForTest(a types.AuthSharedInterface, fn func(ctx context.Context, token string, options types.UserAuthOptions) (string, error)) {
	if v, ok := a.(*authSharedTest); ok {
		v.funcUserFindByAuthToken = fn
	}
}

// SetLoginURLForTest allows tests to configure the login URL used by LinkLogin.
func SetLoginURLForTest(a types.AuthSharedInterface, url string) {
	if v, ok := a.(*authSharedTest); ok {
		v.loginURL = url
	}
}

// SetUseCookiesForTest allows tests to configure whether cookies are used.
func SetUseCookiesForTest(a types.AuthSharedInterface, useCookies bool) {
	if v, ok := a.(*authSharedTest); ok {
		v.useCookies = useCookies
	}
}
