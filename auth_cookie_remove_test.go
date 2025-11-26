package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthCookieRemove_HTTP_NotSecure(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	AuthCookieRemove(w, r)

	res := w.Result()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected a cookie to be set")
	}

	var c *http.Cookie
	for _, ck := range cookies {
		if ck.Name == CookieName {
			c = ck
			break
		}
	}

	if c == nil {
		t.Fatalf("expected cookie %q to be set", CookieName)
	}

	if c.Value != "none" {
		t.Fatalf("expected cookie value %q, got %q", "none", c.Value)
	}

	if c.Path != "/" {
		t.Fatalf("expected cookie path %q, got %q", "/", c.Path)
	}

	if c.HttpOnly {
		t.Fatalf("expected HttpOnly to be false")
	}

	if c.Secure {
		t.Fatalf("expected Secure to be false for HTTP request")
	}

	if !c.Expires.Before(time.Now()) {
		t.Fatalf("expected cookie to be expired, got Expires=%v", c.Expires)
	}
}

func TestAuthCookieRemove_HTTPS_Secure(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)

	AuthCookieRemove(w, r)

	res := w.Result()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected a cookie to be set")
	}

	var c *http.Cookie
	for _, ck := range cookies {
		if ck.Name == CookieName {
			c = ck
			break
		}
	}

	if c == nil {
		t.Fatalf("expected cookie %q to be set", CookieName)
	}

	if !c.Secure {
		t.Fatalf("expected Secure to be true for HTTPS request")
	}
}
