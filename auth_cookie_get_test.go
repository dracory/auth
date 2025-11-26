package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthCookieGetReturnsValueWhenPresent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: "test-token"})

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
