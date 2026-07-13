package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dracory/auth/types"
)

func TestAuthCookieSet_HTTP_SecureByDefault(t *testing.T) {
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
		if ck.Name == types.CookieName {
			c = ck
			break
		}
	}

	if c == nil {
		t.Fatalf("expected cookie %q to be set", types.CookieName)
	}

	if c.Value != "test-token" {
		t.Fatalf("expected cookie value %q, got %q", "test-token", c.Value)
	}

	if c.Path != "/" {
		t.Fatalf("expected cookie path %q, got %q", "/", c.Path)
	}

	if !c.Secure {
		t.Fatalf("expected Secure to be true by default (cfg.Secure=true), even for HTTP request")
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

// TestAuthCookieSet_SecureRespectedBehindReverseProxy is a regression test for SEC-02:
// When cfg.Secure=true and the request is plain HTTP (as behind a reverse proxy),
// the Secure flag must still be set on the cookie.
func TestAuthCookieSet_SecureRespectedBehindReverseProxy(t *testing.T) {
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
		if ck.Name == types.CookieName {
			c = ck
			break
		}
	}

	if c == nil {
		t.Fatalf("expected cookie %q to be set", types.CookieName)
	}

	if !c.Secure {
		t.Fatalf("SEC-02 REGRESSION: expected Secure=true for plain HTTP request when cfg.Secure=true (reverse proxy scenario)")
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
		if ck.Name == types.CookieName {
			c = ck
			break
		}
	}

	if c == nil {
		t.Fatalf("expected cookie %q to be set", types.CookieName)
	}

	if !c.Secure {
		t.Fatalf("expected Secure to be true for HTTPS request")
	}
}

func TestAuthCookieRemove_HTTP_SecureByDefault(t *testing.T) {
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
		if ck.Name == types.CookieName {
			c = ck
			break
		}
	}

	if c == nil {
		t.Fatalf("expected cookie %q to be set", types.CookieName)
	}

	if c.Value != "none" {
		t.Fatalf("expected cookie value %q, got %q", "none", c.Value)
	}

	if c.Path != "/" {
		t.Fatalf("expected cookie path %q, got %q", "/", c.Path)
	}

	if !c.Secure {
		t.Fatalf("expected Secure to be true by default (cfg.Secure=true), even for HTTP request")
	}

	if !c.HttpOnly {
		t.Fatalf("expected HttpOnly to be true")
	}

	if c.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected SameSite to be Lax, got %v", c.SameSite)
	}

	if !c.Expires.Before(time.Now()) {
		t.Fatalf("expected cookie to be expired, got Expires=%v", c.Expires)
	}
}

// TestAuthCookieRemove_SecureRespectedBehindReverseProxy is a regression test for SEC-02:
// When cfg.Secure=true and the request is plain HTTP (as behind a reverse proxy),
// the Secure flag must still be set on the removal cookie.
func TestAuthCookieRemove_SecureRespectedBehindReverseProxy(t *testing.T) {
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
		if ck.Name == types.CookieName {
			c = ck
			break
		}
	}

	if c == nil {
		t.Fatalf("expected cookie %q to be set", types.CookieName)
	}

	if !c.Secure {
		t.Fatalf("SEC-02 REGRESSION: expected Secure=true for plain HTTP removal cookie when cfg.Secure=true (reverse proxy scenario)")
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
		if ck.Name == types.CookieName {
			c = ck
			break
		}
	}

	if c == nil {
		t.Fatalf("expected cookie %q to be set", types.CookieName)
	}

	if !c.Secure {
		t.Fatalf("expected Secure to be true for HTTPS request")
	}
}

func TestAuthCookieSet_SecureFalse(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	AuthCookieSet(w, r, "test-token", types.WithSecure(false))

	res := w.Result()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected a cookie to be set")
	}

	var c *http.Cookie
	for _, ck := range cookies {
		if ck.Name == types.CookieName {
			c = ck
			break
		}
	}

	if c == nil {
		t.Fatalf("expected cookie %q to be set", types.CookieName)
	}

	if c.Secure {
		t.Fatalf("expected Secure to be false when cfg.Secure=false")
	}

	if c.Value != "test-token" {
		t.Fatalf("expected cookie value %q, got %q", "test-token", c.Value)
	}
}

func TestAuthCookieSet_SecureTrue(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	AuthCookieSet(w, r, "test-token", types.WithSecure(true))

	res := w.Result()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected a cookie to be set")
	}

	var c *http.Cookie
	for _, ck := range cookies {
		if ck.Name == types.CookieName {
			c = ck
			break
		}
	}

	if c == nil {
		t.Fatalf("expected cookie %q to be set", types.CookieName)
	}

	if !c.Secure {
		t.Fatalf("expected Secure to be true when cfg.Secure=true")
	}
}

func TestAuthCookieRemove_SecureFalse(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	AuthCookieRemove(w, r, types.WithSecure(false))

	res := w.Result()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected a cookie to be set")
	}

	var c *http.Cookie
	for _, ck := range cookies {
		if ck.Name == types.CookieName {
			c = ck
			break
		}
	}

	if c == nil {
		t.Fatalf("expected cookie %q to be set", types.CookieName)
	}

	if c.Secure {
		t.Fatalf("expected Secure to be false when cfg.Secure=false")
	}

	if c.Value != "none" {
		t.Fatalf("expected cookie value %q, got %q", "none", c.Value)
	}
}

func TestAuthCookieSet_WithCookieConfig(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	cfg := types.CookieConfig{
		HttpOnly: types.CookieHttpWritable,
		Secure:   types.CookieInsecure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   600,
		Path:     "/custom",
		Domain:   "example.com",
	}

	AuthCookieSet(w, r, "test-token", types.WithCookieConfig(cfg))

	res := w.Result()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected a cookie to be set")
	}

	var c *http.Cookie
	for _, ck := range cookies {
		if ck.Name == types.CookieName {
			c = ck
			break
		}
	}

	if c == nil {
		t.Fatalf("expected cookie %q to be set", types.CookieName)
	}

	if c.HttpOnly {
		t.Fatalf("expected HttpOnly=false when WithCookieConfig overrides")
	}

	if c.Secure {
		t.Fatalf("expected Secure=false when WithCookieConfig overrides")
	}

	if c.SameSite != http.SameSiteStrictMode {
		t.Fatalf("expected SameSite Strict, got %v", c.SameSite)
	}

	if c.Path != "/custom" {
		t.Fatalf("expected Path '/custom', got %q", c.Path)
	}

	if c.Domain != "example.com" {
		t.Fatalf("expected Domain 'example.com', got %q", c.Domain)
	}

	if c.MaxAge != 600 {
		t.Fatalf("expected MaxAge 600, got %d", c.MaxAge)
	}
}

func TestAuthCookieGetReturnsValueWhenPresent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: types.CookieName, Value: "test-token"})

	got := AuthCookieGet(req)

	if got != "test-token" {
		t.Fatalf("expected %q, got %q", "test-token", got)
	}
}

func TestAuthCookieGetReturnsEmptyWhenMissing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	got := AuthCookieGet(req)

	if got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestAuthCookieGetWithName_CustomName(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "custom-auth", Value: "custom-token"})

	if got := AuthCookieGetWithName(req, "custom-auth"); got != "custom-token" {
		t.Fatalf("expected %q, got %q", "custom-token", got)
	}
}

func TestAuthCookieGetWithName_FallsBackToDefault(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: types.CookieName, Value: "default-token"})

	if got := AuthCookieGetWithName(req, ""); got != "default-token" {
		t.Fatalf("expected %q, got %q", "default-token", got)
	}
}

func TestAuthCookieSet_WithCookieName(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	AuthCookieSet(w, r, "test-token", types.WithCookieName("custom-cookie"))

	res := w.Result()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected a cookie to be set")
	}

	if cookies[0].Name != "custom-cookie" {
		t.Fatalf("expected cookie name %q, got %q", "custom-cookie", cookies[0].Name)
	}

	if cookies[0].Value != "test-token" {
		t.Fatalf("expected cookie value %q, got %q", "test-token", cookies[0].Value)
	}
}

func TestAuthCookieRemove_WithCookieName(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	AuthCookieRemove(w, r, types.WithCookieName("custom-cookie"))

	res := w.Result()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected a cookie to be set")
	}

	if cookies[0].Name != "custom-cookie" {
		t.Fatalf("expected cookie name %q, got %q", "custom-cookie", cookies[0].Name)
	}

	if cookies[0].Value != "none" {
		t.Fatalf("expected cookie value %q, got %q", "none", cookies[0].Value)
	}
}
