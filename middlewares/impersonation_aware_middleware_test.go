package middlewares

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dracory/auth/types"
)

type mockAuthShared struct {
	useCookies          bool
	temporaryKeyGet     func(key string) (string, error)
	userFindByAuthToken func(ctx context.Context, token string, options types.UserAuthOptions) (string, error)
}

func (m *mockAuthShared) Router() *http.ServeMux                                     { return http.NewServeMux() }
func (m *mockAuthShared) IsRegistrationEnabled() bool                                { return false }
func (m *mockAuthShared) IsPasswordless() bool                                       { return false }
func (m *mockAuthShared) IsVerificationEnabled() bool                                { return false }
func (m *mockAuthShared) WebAuthOrRedirectMiddleware(next http.Handler) http.Handler { return next }
func (m *mockAuthShared) WebAppendUserIdIfExistsMiddleware(next http.Handler) http.Handler {
	return next
}
func (m *mockAuthShared) GetCurrentUserID(r *http.Request) string { return "" }
func (m *mockAuthShared) LinkLogin() string                       { return "" }
func (m *mockAuthShared) LinkLogout() string                      { return "" }
func (m *mockAuthShared) LinkRegister() string                    { return "" }
func (m *mockAuthShared) LinkRegisterCodeVerify() string          { return "" }
func (m *mockAuthShared) LinkRedirectOnSuccess() string           { return "" }
func (m *mockAuthShared) LinkApiLogin() string                    { return "" }
func (m *mockAuthShared) LinkApiLogout() string                   { return "" }
func (m *mockAuthShared) LinkApiRegister() string                 { return "" }
func (m *mockAuthShared) LinkApiRegisterCodeVerify() string       { return "" }
func (m *mockAuthShared) LinkApiImpersonateStart() string         { return "" }
func (m *mockAuthShared) LinkApiImpersonateStop() string          { return "" }
func (m *mockAuthShared) GetEndpoint() string                     { return "" }
func (m *mockAuthShared) SetEndpoint(string)                      {}
func (m *mockAuthShared) GetLogger() *slog.Logger                 { return nil }
func (m *mockAuthShared) SetLogger(*slog.Logger)                  {}
func (m *mockAuthShared) GetLayout() func(string) string          { return nil }
func (m *mockAuthShared) SetLayout(func(string) string)           {}
func (m *mockAuthShared) GetFuncTemporaryKeyGet() func(key string) (string, error) {
	return m.temporaryKeyGet
}
func (m *mockAuthShared) SetFuncTemporaryKeyGet(func(key string) (string, error)) {}
func (m *mockAuthShared) GetFuncTemporaryKeySet() func(key string, value string, expiresSeconds int) error {
	return nil
}
func (m *mockAuthShared) SetFuncTemporaryKeySet(func(key string, value string, expiresSeconds int) error) {
}
func (m *mockAuthShared) GetUseCookies() bool { return m.useCookies }
func (m *mockAuthShared) SetUseCookies(bool)  {}
func (m *mockAuthShared) GetFuncUserFindByAuthToken() func(ctx context.Context, token string, options types.UserAuthOptions) (string, error) {
	return m.userFindByAuthToken
}
func (m *mockAuthShared) SetFuncUserFindByAuthToken(func(ctx context.Context, token string, options types.UserAuthOptions) (string, error)) {
}
func (m *mockAuthShared) GetDisableRateLimit() bool                          { return false }
func (m *mockAuthShared) SetDisableRateLimit(bool)                           {}
func (m *mockAuthShared) GetPasswordStrength() *types.PasswordStrengthConfig { return nil }
func (m *mockAuthShared) SetPasswordStrength(*types.PasswordStrengthConfig)  {}
func (m *mockAuthShared) GetFuncUserLogin() func(context.Context, string, string, types.UserAuthOptions) (string, error) {
	return nil
}
func (m *mockAuthShared) SetFuncUserLogin(func(context.Context, string, string, types.UserAuthOptions) (string, error)) {
}
func (m *mockAuthShared) GetPasswordlessUserRegister() func(context.Context, string, string, string, types.UserAuthOptions) error {
	return nil
}
func (m *mockAuthShared) SetPasswordlessUserRegister(func(context.Context, string, string, string, types.UserAuthOptions) error) {
}
func (m *mockAuthShared) GetFuncUserRegister() func(context.Context, string, string, string, string, types.UserAuthOptions) error {
	return nil
}
func (m *mockAuthShared) SetFuncUserRegister(func(context.Context, string, string, string, string, types.UserAuthOptions) error) {
}
func (m *mockAuthShared) GetFuncUserPasswordChange() func(context.Context, string, string, types.UserAuthOptions) error {
	return nil
}
func (m *mockAuthShared) SetFuncUserPasswordChange(func(context.Context, string, string, types.UserAuthOptions) error) {
}
func (m *mockAuthShared) GetFuncUserLogout() func(context.Context, string, types.UserAuthOptions) error {
	return nil
}
func (m *mockAuthShared) SetFuncUserLogout(func(context.Context, string, types.UserAuthOptions) error) {
}
func (m *mockAuthShared) GetPasswordlessUserFindByEmail() func(context.Context, string, types.UserAuthOptions) (string, error) {
	return nil
}
func (m *mockAuthShared) SetPasswordlessUserFindByEmail(func(context.Context, string, types.UserAuthOptions) (string, error)) {
}
func (m *mockAuthShared) GetFuncUserFindByUsername() func(context.Context, string, string, string, types.UserAuthOptions) (string, error) {
	return nil
}
func (m *mockAuthShared) SetFuncUserFindByUsername(func(context.Context, string, string, string, types.UserAuthOptions) (string, error)) {
}
func (m *mockAuthShared) GetFuncEmailTemplatePasswordRestore() func(context.Context, string, string, types.UserAuthOptions) string {
	return nil
}
func (m *mockAuthShared) SetFuncEmailTemplatePasswordRestore(func(context.Context, string, string, types.UserAuthOptions) string) {
}
func (m *mockAuthShared) GetFuncEmailTemplateRegisterCode() func(context.Context, string, string, types.UserAuthOptions) string {
	return nil
}
func (m *mockAuthShared) SetFuncEmailTemplateRegisterCode(func(context.Context, string, string, types.UserAuthOptions) string) {
}
func (m *mockAuthShared) GetFuncEmailSend() func(context.Context, string, string, string) error {
	return nil
}
func (m *mockAuthShared) SetFuncEmailSend(func(context.Context, string, string, string) error) {}
func (m *mockAuthShared) GetPasswordlessFuncEmailTemplateLoginCode() func(context.Context, string, string, types.UserAuthOptions) string {
	return nil
}
func (m *mockAuthShared) SetPasswordlessFuncEmailTemplateLoginCode(func(context.Context, string, string, types.UserAuthOptions) string) {
}
func (m *mockAuthShared) GetPasswordlessFuncEmailTemplateRegisterCode() func(context.Context, string, string, types.UserAuthOptions) string {
	return nil
}
func (m *mockAuthShared) SetPasswordlessFuncEmailTemplateRegisterCode(func(context.Context, string, string, types.UserAuthOptions) string) {
}
func (m *mockAuthShared) GetPasswordlessFuncEmailSend() func(context.Context, string, string, string) error {
	return nil
}
func (m *mockAuthShared) SetPasswordlessFuncEmailSend(func(context.Context, string, string, string) error) {
}
func (m *mockAuthShared) GetFuncUserStoreAuthToken() func(context.Context, string, string, types.UserAuthOptions) error {
	return nil
}
func (m *mockAuthShared) SetFuncUserStoreAuthToken(func(context.Context, string, string, types.UserAuthOptions) error) {
}
func (m *mockAuthShared) SetAuthCookie(http.ResponseWriter, *http.Request, string) {}
func (m *mockAuthShared) RemoveAuthCookie(http.ResponseWriter, *http.Request)      {}
func (m *mockAuthShared) GetObservabilityHooks() types.ObservabilityHooks          { return nil }
func (m *mockAuthShared) SetObservabilityHooks(types.ObservabilityHooks)           {}
func (m *mockAuthShared) AuthenticateViaUsername(http.ResponseWriter, *http.Request, string, string, string) {
}
func (m *mockAuthShared) IsImpersonationEnabled() bool { return false }
func (m *mockAuthShared) GetFuncCanImpersonate() func(context.Context, string, string) (bool, error) {
	return nil
}
func (m *mockAuthShared) SetFuncCanImpersonate(func(context.Context, string, string) (bool, error)) {}
func (m *mockAuthShared) GetFuncImpersonationStart() func(context.Context, string, string) error {
	return nil
}
func (m *mockAuthShared) SetFuncImpersonationStart(func(context.Context, string, string) error) {}
func (m *mockAuthShared) GetFuncImpersonationStop() func(context.Context, string, string) error {
	return nil
}
func (m *mockAuthShared) SetFuncImpersonationStop(func(context.Context, string, string) error) {}
func (m *mockAuthShared) LinkPasswordRestore() string                                          { return "" }
func (m *mockAuthShared) LinkPasswordReset(string) string                                      { return "" }
func (m *mockAuthShared) LinkApiPasswordRestore() string                                       { return "" }
func (m *mockAuthShared) LinkApiPasswordReset() string                                         { return "" }
func (m *mockAuthShared) LinkLoginCodeVerify() string                                          { return "" }
func (m *mockAuthShared) LinkApiLoginCodeVerify() string                                       { return "" }

func TestImpersonationAwareMiddlewareImpersonationActive(t *testing.T) {
	mock := &mockAuthShared{
		useCookies: true,
		temporaryKeyGet: func(key string) (string, error) {
			if key == "imp:current-token" {
				return "original-admin-token", nil
			}
			return "", nil
		},
		userFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
			if token == "original-admin-token" {
				return "admin-123", nil
			}
			return "", nil
		},
	}

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		imp := r.Context().Value(types.IsImpersonating{})
		if imp != true {
			t.Fatal("expected IsImpersonating to be true in context")
		}
		impID := r.Context().Value(types.ImpersonatorUserID{})
		if impID != "admin-123" {
			t.Fatalf("expected ImpersonatorUserID to be admin-123, got %v", impID)
		}
	})

	req, _ := http.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: types.CookieName, Value: "current-token"})
	recorder := httptest.NewRecorder()

	middleware := ImpersonationAwareMiddleware(next, mock)
	middleware.ServeHTTP(recorder, req)

	if !called {
		t.Fatal("expected next handler to be called")
	}
}

func TestImpersonationAwareMiddlewareNotImpersonating(t *testing.T) {
	mock := &mockAuthShared{
		useCookies: true,
		temporaryKeyGet: func(key string) (string, error) {
			return "", nil
		},
		userFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
			return "", nil
		},
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		imp := r.Context().Value(types.IsImpersonating{})
		if imp != nil {
			t.Fatal("expected IsImpersonating to be nil when not impersonating")
		}
	})

	req, _ := http.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: types.CookieName, Value: "some-token"})
	recorder := httptest.NewRecorder()

	middleware := ImpersonationAwareMiddleware(next, mock)
	middleware.ServeHTTP(recorder, req)
}

func TestImpersonationAwareMiddlewareNoAuthToken(t *testing.T) {
	mock := &mockAuthShared{
		useCookies:          true,
		temporaryKeyGet:     func(key string) (string, error) { return "", nil },
		userFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) { return "", nil },
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		imp := r.Context().Value(types.IsImpersonating{})
		if imp != nil {
			t.Fatal("expected IsImpersonating to be nil when no auth token")
		}
	})

	req, _ := http.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()

	middleware := ImpersonationAwareMiddleware(next, mock)
	middleware.ServeHTTP(recorder, req)
}
