package testutils

import (
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
	endpoint          string
	layout            func(content string) string
	logger            *slog.Logger
	registration      bool
	redirectOnSuccess string
}

func (a *authSharedTest) Router() *http.ServeMux { return http.NewServeMux() }

func (a *authSharedTest) IsRegistrationEnabled() bool { return a.registration }

func (a *authSharedTest) WebAuthOrRedirectMiddleware(next http.Handler) http.Handler { return next }

func (a *authSharedTest) ApiAuthOrErrorMiddleware(next http.Handler) http.Handler { return next }

func (a *authSharedTest) WebAppendUserIdIfExistsMiddleware(next http.Handler) http.Handler {
	return next
}

func (a *authSharedTest) GetCurrentUserID(r *http.Request) string { return "" }

func (a *authSharedTest) LinkLogin() string { return "" }

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
