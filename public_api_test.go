package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Public API shortcut tests to ensure the v0.29.0 exported functions remain available.
func TestAuthCookieSet_SetsCookie(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	AuthCookieSet(w, r, "public-token")

	res := w.Result()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected a cookie to be set")
	}

	if cookies[0].Name != CookieName {
		t.Fatalf("expected cookie name %q, got %q", CookieName, cookies[0].Name)
	}

	if cookies[0].Value != "public-token" {
		t.Fatalf("expected cookie value %q, got %q", "public-token", cookies[0].Value)
	}
}

func TestAuthCookieGet_ReturnsCookieValueWhenPresent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: "public-token"})

	if got := AuthCookieGet(req); got != "public-token" {
		t.Fatalf("expected %q, got %q", "public-token", got)
	}
}

func TestAuthCookieGet_ReturnsEmptyWhenMissing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if got := AuthCookieGet(req); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestAuthCookieRemove_RemovesCookie(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	AuthCookieRemove(w, r)

	res := w.Result()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected a cookie to be set")
	}

	if cookies[0].Value != "none" {
		t.Fatalf("expected cookie value %q, got %q", "none", cookies[0].Value)
	}

	if !cookies[0].Expires.Before(time.Now()) {
		t.Fatalf("expected expired cookie, got Expires=%v", cookies[0].Expires)
	}
}

func TestAuthTokenRetrieve_FromCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: "cookie-token"})

	if got := AuthTokenRetrieve(req, true); got != "cookie-token" {
		t.Fatalf("expected %q, got %q", "cookie-token", got)
	}
}

func TestAuthTokenRetrieve_FromBearer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer bearer-token")

	if got := AuthTokenRetrieve(req, false); got != "bearer-token" {
		t.Fatalf("expected %q, got %q", "bearer-token", got)
	}
}
