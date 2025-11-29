package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSetAuthCookie_UsesDefaultConfigWhenZero(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	var a authImplementation
	// a.cookieConfig is zero value; should fall back to defaultCookieConfig.
	a.setAuthCookie(w, r, "token")

	res := w.Result()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected a cookie to be set")
	}

	c := cookies[0]

	if !c.HttpOnly {
		t.Fatalf("expected HttpOnly to be true by default")
	}

	if c.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected SameSite Lax by default, got %v", c.SameSite)
	}

	if c.Path != "/" {
		t.Fatalf("expected default Path '/' got %q", c.Path)
	}

	if c.Secure {
		t.Fatalf("expected Secure to be false for HTTP request")
	}

	if c.MaxAge <= 0 {
		t.Fatalf("expected positive MaxAge, got %d", c.MaxAge)
	}

	if !c.Expires.After(time.Now()) {
		t.Fatalf("expected future Expires, got %v", c.Expires)
	}
}

func TestSetAuthCookie_UsesCustomConfig(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)

	a := authImplementation{
		cookieConfig: CookieConfig{
			HttpOnly: false,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   600,
			Domain:   "example.com",
			Path:     "/auth",
		},
	}

	a.setAuthCookie(w, r, "token")

	res := w.Result()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected a cookie to be set")
	}

	c := cookies[0]

	if c.HttpOnly {
		t.Fatalf("expected HttpOnly to be false from custom config")
	}

	if !c.Secure {
		t.Fatalf("expected Secure to be true for HTTPS with custom config")
	}

	if c.SameSite != http.SameSiteStrictMode {
		t.Fatalf("expected SameSite Strict, got %v", c.SameSite)
	}

	if c.Path != "/auth" {
		t.Fatalf("expected Path '/auth', got %q", c.Path)
	}

	if c.Domain != "example.com" {
		t.Fatalf("expected Domain 'example.com', got %q", c.Domain)
	}

	if c.MaxAge != 600 {
		t.Fatalf("expected MaxAge 600, got %d", c.MaxAge)
	}

	if c.Expires.Before(time.Now()) {
		t.Fatalf("expected future Expires, got %v", c.Expires)
	}
}

func TestRemoveAuthCookie_UsesCustomConfig(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)

	a := authImplementation{
		cookieConfig: CookieConfig{
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Domain:   "example.com",
			Path:     "/auth",
		},
	}

	a.removeAuthCookie(w, r)

	res := w.Result()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected a cookie to be set")
	}

	c := cookies[0]

	if !c.HttpOnly {
		t.Fatalf("expected HttpOnly true")
	}

	if !c.Secure {
		t.Fatalf("expected Secure true for HTTPS with custom config")
	}

	if c.SameSite != http.SameSiteStrictMode {
		t.Fatalf("expected SameSite Strict, got %v", c.SameSite)
	}

	if c.Path != "/auth" {
		t.Fatalf("expected Path '/auth', got %q", c.Path)
	}

	if c.Domain != "example.com" {
		t.Fatalf("expected Domain 'example.com', got %q", c.Domain)
	}

	if c.MaxAge != -1 {
		t.Fatalf("expected MaxAge -1 for removal, got %d", c.MaxAge)
	}

	if !c.Expires.Before(time.Now()) {
		t.Fatalf("expected expired cookie, got %v", c.Expires)
	}
}
