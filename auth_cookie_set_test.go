package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthCookieSet_HTTP_NotSecure(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	AuthCookieSet(w, r, "test-token")

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

	if c.Value != "test-token" {
		t.Fatalf("expected cookie value %q, got %q", "test-token", c.Value)
	}

	if c.Path != "/" {
		t.Fatalf("expected cookie path %q, got %q", "/", c.Path)
	}

	if c.Secure {
		t.Fatalf("expected Secure to be false for HTTP request")
	}

	if !c.HttpOnly {
		t.Fatalf("expected HttpOnly to be true")
	}

	if c.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected SameSite to be Lax, got %v", c.SameSite)
	}

	if !c.Expires.After(time.Now()) {
		t.Fatalf("expected cookie expiration in the future, got Expires=%v", c.Expires)
	}
}

func TestAuthCookieSet_HTTPS_Secure(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)

	AuthCookieSet(w, r, "secure-token")

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
