package middlewares

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracory/auth/types"
)

type mockAuthSharedForAPI struct {
	useCookies          bool
	userFindByAuthToken func(ctx context.Context, token string, options types.UserAuthOptions) (string, error)
}

func (m *mockAuthSharedForAPI) Router() *http.ServeMux      { return http.NewServeMux() }
func (m *mockAuthSharedForAPI) IsRegistrationEnabled() bool { return false }
func (m *mockAuthSharedForAPI) IsPasswordless() bool        { return false }
func (m *mockAuthSharedForAPI) IsVerificationEnabled() bool { return false }
func (m *mockAuthSharedForAPI) WebAuthOrRedirectMiddleware(next http.Handler) http.Handler {
	return next
}
func (m *mockAuthSharedForAPI) WebAppendUserIdIfExistsMiddleware(next http.Handler) http.Handler {
	return next
}
func (m *mockAuthSharedForAPI) GetCurrentUserID(r *http.Request) string                 { return "" }
func (m *mockAuthSharedForAPI) LinkLogin() string                                       { return "" }
func (m *mockAuthSharedForAPI) LinkLogout() string                                      { return "" }
func (m *mockAuthSharedForAPI) LinkRegister() string                                    { return "" }
func (m *mockAuthSharedForAPI) LinkRegisterCodeVerify() string                          { return "" }
func (m *mockAuthSharedForAPI) LinkRedirectOnSuccess() string                           { return "" }
func (m *mockAuthSharedForAPI) LinkApiLogin() string                                    { return "" }
func (m *mockAuthSharedForAPI) LinkApiLogout() string                                   { return "" }
func (m *mockAuthSharedForAPI) LinkApiRegister() string                                 { return "" }
func (m *mockAuthSharedForAPI) LinkApiRegisterCodeVerify() string                       { return "" }
func (m *mockAuthSharedForAPI) LinkApiImpersonateStart() string                         { return "" }
func (m *mockAuthSharedForAPI) LinkApiImpersonateStop() string                          { return "" }
func (m *mockAuthSharedForAPI) GetEndpoint() string                                     { return "" }
func (m *mockAuthSharedForAPI) SetEndpoint(string)                                      {}
func (m *mockAuthSharedForAPI) GetLogger() *slog.Logger                                 { return nil }
func (m *mockAuthSharedForAPI) SetLogger(*slog.Logger)                                  {}
func (m *mockAuthSharedForAPI) GetLayout() func(string) string                          { return nil }
func (m *mockAuthSharedForAPI) SetLayout(func(string) string)                           {}
func (m *mockAuthSharedForAPI) GetFuncTemporaryKeyGet() func(string) (string, error)    { return nil }
func (m *mockAuthSharedForAPI) SetFuncTemporaryKeyGet(func(string) (string, error))     {}
func (m *mockAuthSharedForAPI) GetFuncTemporaryKeySet() func(string, string, int) error { return nil }
func (m *mockAuthSharedForAPI) SetFuncTemporaryKeySet(func(string, string, int) error)  {}
func (m *mockAuthSharedForAPI) GetUseCookies() bool                                     { return m.useCookies }
func (m *mockAuthSharedForAPI) SetUseCookies(bool)                                      {}
func (m *mockAuthSharedForAPI) GetFuncUserFindByAuthToken() func(context.Context, string, types.UserAuthOptions) (string, error) {
	return m.userFindByAuthToken
}
func (m *mockAuthSharedForAPI) SetFuncUserFindByAuthToken(func(context.Context, string, types.UserAuthOptions) (string, error)) {
}
func (m *mockAuthSharedForAPI) GetDisableRateLimit() bool                          { return false }
func (m *mockAuthSharedForAPI) SetDisableRateLimit(bool)                           {}
func (m *mockAuthSharedForAPI) GetPasswordStrength() *types.PasswordStrengthConfig { return nil }
func (m *mockAuthSharedForAPI) SetPasswordStrength(*types.PasswordStrengthConfig)  {}
func (m *mockAuthSharedForAPI) GetFuncUserLogin() func(context.Context, string, string, types.UserAuthOptions) (string, error) {
	return nil
}
func (m *mockAuthSharedForAPI) SetFuncUserLogin(func(context.Context, string, string, types.UserAuthOptions) (string, error)) {
}
func (m *mockAuthSharedForAPI) GetPasswordlessUserRegister() func(context.Context, string, string, string, types.UserAuthOptions) error {
	return nil
}
func (m *mockAuthSharedForAPI) SetPasswordlessUserRegister(func(context.Context, string, string, string, types.UserAuthOptions) error) {
}
func (m *mockAuthSharedForAPI) GetFuncUserRegister() func(context.Context, string, string, string, string, types.UserAuthOptions) error {
	return nil
}
func (m *mockAuthSharedForAPI) SetFuncUserRegister(func(context.Context, string, string, string, string, types.UserAuthOptions) error) {
}
func (m *mockAuthSharedForAPI) GetFuncUserPasswordChange() func(context.Context, string, string, types.UserAuthOptions) error {
	return nil
}
func (m *mockAuthSharedForAPI) SetFuncUserPasswordChange(func(context.Context, string, string, types.UserAuthOptions) error) {
}
func (m *mockAuthSharedForAPI) GetFuncUserLogout() func(context.Context, string, types.UserAuthOptions) error {
	return nil
}
func (m *mockAuthSharedForAPI) SetFuncUserLogout(func(context.Context, string, types.UserAuthOptions) error) {
}
func (m *mockAuthSharedForAPI) GetPasswordlessUserFindByEmail() func(context.Context, string, types.UserAuthOptions) (string, error) {
	return nil
}
func (m *mockAuthSharedForAPI) SetPasswordlessUserFindByEmail(func(context.Context, string, types.UserAuthOptions) (string, error)) {
}
func (m *mockAuthSharedForAPI) GetFuncUserFindByUsername() func(context.Context, string, string, string, types.UserAuthOptions) (string, error) {
	return nil
}
func (m *mockAuthSharedForAPI) SetFuncUserFindByUsername(func(context.Context, string, string, string, types.UserAuthOptions) (string, error)) {
}
func (m *mockAuthSharedForAPI) GetFuncEmailTemplatePasswordRestore() func(context.Context, string, string, types.UserAuthOptions) string {
	return nil
}
func (m *mockAuthSharedForAPI) SetFuncEmailTemplatePasswordRestore(func(context.Context, string, string, types.UserAuthOptions) string) {
}
func (m *mockAuthSharedForAPI) GetFuncEmailTemplateRegisterCode() func(context.Context, string, string, types.UserAuthOptions) string {
	return nil
}
func (m *mockAuthSharedForAPI) SetFuncEmailTemplateRegisterCode(func(context.Context, string, string, types.UserAuthOptions) string) {
}
func (m *mockAuthSharedForAPI) GetFuncEmailSend() func(context.Context, string, string, string) error {
	return nil
}
func (m *mockAuthSharedForAPI) SetFuncEmailSend(func(context.Context, string, string, string) error) {
}
func (m *mockAuthSharedForAPI) GetPasswordlessFuncEmailTemplateLoginCode() func(context.Context, string, string, types.UserAuthOptions) string {
	return nil
}
func (m *mockAuthSharedForAPI) SetPasswordlessFuncEmailTemplateLoginCode(func(context.Context, string, string, types.UserAuthOptions) string) {
}
func (m *mockAuthSharedForAPI) GetPasswordlessFuncEmailTemplateRegisterCode() func(context.Context, string, string, types.UserAuthOptions) string {
	return nil
}
func (m *mockAuthSharedForAPI) SetPasswordlessFuncEmailTemplateRegisterCode(func(context.Context, string, string, types.UserAuthOptions) string) {
}
func (m *mockAuthSharedForAPI) GetPasswordlessFuncEmailSend() func(context.Context, string, string, string) error {
	return nil
}
func (m *mockAuthSharedForAPI) SetPasswordlessFuncEmailSend(func(context.Context, string, string, string) error) {
}
func (m *mockAuthSharedForAPI) GetFuncUserStoreAuthToken() func(context.Context, string, string, types.UserAuthOptions) error {
	return nil
}
func (m *mockAuthSharedForAPI) SetFuncUserStoreAuthToken(func(context.Context, string, string, types.UserAuthOptions) error) {
}
func (m *mockAuthSharedForAPI) SetAuthCookie(http.ResponseWriter, *http.Request, string) {}
func (m *mockAuthSharedForAPI) RemoveAuthCookie(http.ResponseWriter, *http.Request)      {}
func (m *mockAuthSharedForAPI) GetObservabilityHooks() types.ObservabilityHooks          { return nil }
func (m *mockAuthSharedForAPI) SetObservabilityHooks(types.ObservabilityHooks)           {}
func (m *mockAuthSharedForAPI) AuthenticateViaUsername(http.ResponseWriter, *http.Request, string, string, string) {
}
func (m *mockAuthSharedForAPI) IsImpersonationEnabled() bool { return false }
func (m *mockAuthSharedForAPI) GetFuncCanImpersonate() func(context.Context, string, string) (bool, error) {
	return nil
}
func (m *mockAuthSharedForAPI) SetFuncCanImpersonate(func(context.Context, string, string) (bool, error)) {
}
func (m *mockAuthSharedForAPI) GetFuncImpersonationStart() func(context.Context, string, string) error {
	return nil
}
func (m *mockAuthSharedForAPI) SetFuncImpersonationStart(func(context.Context, string, string) error) {
}
func (m *mockAuthSharedForAPI) GetFuncImpersonationStop() func(context.Context, string, string) error {
	return nil
}
func (m *mockAuthSharedForAPI) SetFuncImpersonationStop(func(context.Context, string, string) error) {
}
func (m *mockAuthSharedForAPI) LinkPasswordRestore() string     { return "" }
func (m *mockAuthSharedForAPI) LinkPasswordReset(string) string { return "" }
func (m *mockAuthSharedForAPI) LinkApiPasswordRestore() string  { return "" }
func (m *mockAuthSharedForAPI) LinkApiPasswordReset() string    { return "" }
func (m *mockAuthSharedForAPI) LinkLoginCodeVerify() string     { return "" }
func (m *mockAuthSharedForAPI) LinkApiLoginCodeVerify() string  { return "" }

func TestApiAuthOrErrorMiddleware_NoToken(t *testing.T) {
	mock := &mockAuthSharedForAPI{useCookies: true}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called without a token")
	})

	handler := ApiAuthOrErrorMiddleware(next, mock)
	req, _ := http.NewRequest("GET", "/protected", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "unauthenticated") {
		t.Fatalf("expected response body to contain 'unauthenticated', got %q", body)
	}
}

func TestApiAuthOrErrorMiddleware_FindByAuthTokenError(t *testing.T) {
	mock := &mockAuthSharedForAPI{
		useCookies: true,
		userFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
			return "", fmt.Errorf("db error")
		},
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called on auth error")
	})

	handler := ApiAuthOrErrorMiddleware(next, mock)
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: types.CookieName, Value: "some-token"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "unauthenticated") {
		t.Fatalf("expected response body to contain 'unauthenticated', got %q", body)
	}
}

func TestApiAuthOrErrorMiddleware_EmptyUserID(t *testing.T) {
	mock := &mockAuthSharedForAPI{
		useCookies: true,
		userFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
			return "", nil
		},
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called with empty user ID")
	})

	handler := ApiAuthOrErrorMiddleware(next, mock)
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: types.CookieName, Value: "some-token"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "unauthenticated") {
		t.Fatalf("expected response body to contain 'unauthenticated', got %q", body)
	}
}

func TestApiAuthOrErrorMiddleware_Success(t *testing.T) {
	mock := &mockAuthSharedForAPI{
		useCookies: true,
		userFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
			if token == "valid-token" {
				return "user-123", nil
			}
			return "", nil
		},
	}

	called := false
	var ctxUserID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		val := r.Context().Value(types.AuthenticatedUserID{})
		if val != nil {
			ctxUserID = val.(string)
		}
	})

	handler := ApiAuthOrErrorMiddleware(next, mock)
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: types.CookieName, Value: "valid-token"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called")
	}
	if ctxUserID != "user-123" {
		t.Fatalf("expected context user ID 'user-123', got %q", ctxUserID)
	}
}

func TestApiAuthOrErrorMiddleware_SuccessWithBearerToken(t *testing.T) {
	mock := &mockAuthSharedForAPI{
		useCookies: false,
		userFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
			if token == "bearer-token" {
				return "user-456", nil
			}
			return "", nil
		},
	}

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		val := r.Context().Value(types.AuthenticatedUserID{})
		if val == nil || val.(string) != "user-456" {
			t.Fatalf("expected context user ID 'user-456', got %v", val)
		}
	})

	handler := ApiAuthOrErrorMiddleware(next, mock)
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer bearer-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called")
	}
}
