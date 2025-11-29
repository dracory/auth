package auth

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// )

// func TestAuthTokenRetrieve_UsesCookieWhenEnabled(t *testing.T) {
// 	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
// 	req.AddCookie(&http.Cookie{Name: CookieName, Value: "cookie-token"})

// 	// Even if other sources are present, cookie should win when useCookies is true
// 	req.Header.Set("Authorization", "Bearer header-token")
// 	q := req.URL.Query()
// 	q.Set("api_key", "api-key-token")
// 	q.Set("token", "param-token")
// 	req.URL.RawQuery = q.Encode()

// 	token := AuthTokenRetrieve(req, true)
// 	if token != "cookie-token" {
// 		t.Fatalf("expected token %q, got %q", "cookie-token", token)
// 	}
// }

// func TestAuthTokenRetrieve_CookieEnabled_NoCookieReturnsEmpty(t *testing.T) {
// 	req := httptest.NewRequest(http.MethodGet, "http://example.com/?api_key=api-key-token", nil)
// 	req.Header.Set("Authorization", "Bearer header-token")

// 	token := AuthTokenRetrieve(req, true)
// 	if token != "" {
// 		t.Fatalf("expected empty token when cookie is enabled but missing, got %q", token)
// 	}
// }

// func TestAuthTokenRetrieve_BearerHeaderWhenCookiesDisabled(t *testing.T) {
// 	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
// 	req.Header.Set("Authorization", "Bearer header-token")

// 	token := AuthTokenRetrieve(req, false)
// 	if token != "header-token" {
// 		t.Fatalf("expected token %q, got %q", "header-token", token)
// 	}
// }

// func TestAuthTokenRetrieve_ApiKeyParamWhenNoHeader(t *testing.T) {
// 	req := httptest.NewRequest(http.MethodGet, "http://example.com/?api_key=api-key-token", nil)

// 	token := AuthTokenRetrieve(req, false)
// 	if token != "api-key-token" {
// 		t.Fatalf("expected token %q, got %q", "api-key-token", token)
// 	}
// }

// func TestAuthTokenRetrieve_TokenParamFallback(t *testing.T) {
// 	req := httptest.NewRequest(http.MethodGet, "http://example.com/?token=param-token", nil)

// 	token := AuthTokenRetrieve(req, false)
// 	if token != "param-token" {
// 		t.Fatalf("expected token %q, got %q", "param-token", token)
// 	}
// }

// func TestAuthTokenRetrieve_ReturnsEmptyWhenNoSources(t *testing.T) {
// 	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

// 	token := AuthTokenRetrieve(req, false)
// 	if token != "" {
// 		t.Fatalf("expected empty token, got %q", token)
// 	}
// }
