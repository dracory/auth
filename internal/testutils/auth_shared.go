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
	endpoint                string
	layout                  func(content string) string
	logger                  *slog.Logger
	registration            bool
	passwordless            bool
	verification            bool
	temporaryKeyGet         func(key string) (string, error)
	funcUserFindByAuthToken func(ctx context.Context, token string, options types.UserAuthOptions) (string, error)
	redirectOnSuccess       string
	loginURL                string
	useCookies              bool
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

func (a *authSharedTest) TemporaryKeyGet(token string) (string, error) { return "", nil }

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
