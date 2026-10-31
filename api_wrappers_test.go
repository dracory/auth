package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth/internal/api/api_authknight_callback"
	"github.com/dracory/auth/internal/api/api_login"
	"github.com/dracory/auth/internal/api/api_logout"
	"github.com/dracory/auth/internal/api/api_password_restore"
	"github.com/dracory/auth/internal/api/api_register"
	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/types"
)

func post(t *testing.T, path string, values url.Values) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return httptest.NewRecorder(), req
}

func TestApiLoginWithAuth(t *testing.T) {
	a, err := NewUsernameAndPasswordAuth(testutils.NewUsernameAndPasswordConfigForTest())
	if err != nil {
		t.Fatal(err)
	}
	rec, req := post(t, "/api/login", url.Values{})
	api_login.ApiLoginWithAuth(rec, req, a)

	if !strings.Contains(rec.Body.String(), `"status":`) {
		t.Fatalf("expected JSON response, got %d %q", rec.Code, rec.Body.String())
	}
}

func TestApiLogoutWithAuth(t *testing.T) {
	a, err := NewUsernameAndPasswordAuth(testutils.NewUsernameAndPasswordConfigForTest())
	if err != nil {
		t.Fatal(err)
	}
	rec, req := post(t, "/api/logout", url.Values{})
	api_logout.ApiLogoutWithAuth(rec, req, a)

	if !strings.Contains(rec.Body.String(), `"status":`) {
		t.Fatalf("expected JSON response, got %d %q", rec.Code, rec.Body.String())
	}
}

func TestApiPasswordRestoreWithAuth(t *testing.T) {
	a, err := NewUsernameAndPasswordAuth(testutils.NewUsernameAndPasswordConfigForTest())
	if err != nil {
		t.Fatal(err)
	}
	rec, req := post(t, "/api/restore-password", url.Values{})
	api_password_restore.ApiPasswordRestoreWithAuth(rec, req, a)

	if !strings.Contains(rec.Body.String(), `"status":`) {
		t.Fatalf("expected JSON response, got %d %q", rec.Code, rec.Body.String())
	}
}

func TestApiRegisterWithAuth(t *testing.T) {
	a, err := NewUsernameAndPasswordAuth(testutils.NewUsernameAndPasswordConfigForTest())
	if err != nil {
		t.Fatal(err)
	}
	rec, req := post(t, "/api/register", url.Values{})
	api_register.ApiRegisterWithAuth(rec, req, a)

	if !strings.Contains(rec.Body.String(), `"status":`) {
		t.Fatalf("expected JSON response, got %d %q", rec.Code, rec.Body.String())
	}
}

func TestApiAuthKnightCallbackWithAuth(t *testing.T) {
	a, err := NewAuthKnightAuth(newValidAuthKnightConfig())
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/authknight-callback", nil)
	api_authknight_callback.ApiAuthKnightCallbackWithAuth(rec, req, a)

	// Missing once token → legacy path redirects back to the success URL.
	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected redirect, got %d %q", rec.Code, rec.Body.String())
	}
}

func TestApiAuthKnightCallbackWithAuth_PostAuthHooks(t *testing.T) {
	a, err := NewAuthKnightAuth(newValidAuthKnightConfig())
	if err != nil {
		t.Fatal(err)
	}
	setter, ok := any(a).(interface{ SetPostAuthHooks(types.PostAuthHooks) })
	if !ok {
		t.Fatal("expected impl to expose SetPostAuthHooks")
	}
	setter.SetPostAuthHooks(types.PostAuthHooks{
		EmailAllowed: func(ctx context.Context, email string) (bool, string) {
			return true, ""
		},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/authknight-callback", nil)
	api_authknight_callback.ApiAuthKnightCallbackWithAuth(rec, req, a)

	// Post-auth path also redirects to the fallback on a missing once token.
	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected redirect, got %d %q", rec.Code, rec.Body.String())
	}
}
