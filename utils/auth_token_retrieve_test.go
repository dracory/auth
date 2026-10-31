package utils

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth/types"
)

func TestAuthTokenRetrieve_UsesCookieWhenEnabled(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.AddCookie(&http.Cookie{Name: types.CookieName, Value: "cookie-token"})

	// Even if other sources are present, cookie should win when useCookies is true
	req.Header.Set("Authorization", "Bearer header-token")
	q := req.URL.Query()
	q.Set("api_key", "api-key-token")
	q.Set("token", "param-token")
	req.URL.RawQuery = q.Encode()

	token := AuthTokenRetrieve(req, true)
	if token != "cookie-token" {
		t.Fatalf("expected token %q, got %q", "cookie-token", token)
	}
}

func TestAuthTokenRetrieve_CookieEnabled_NoCookieReturnsEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/?api_key=api-key-token", nil)
	req.Header.Set("Authorization", "Bearer header-token")

	token := AuthTokenRetrieve(req, true)
	if token != "" {
		t.Fatalf("expected empty token when cookie is enabled but missing, got %q", token)
	}
}

func TestAuthTokenRetrieve_BearerHeaderWhenCookiesDisabled(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.Header.Set("Authorization", "Bearer header-token")

	token := AuthTokenRetrieve(req, false)
	if token != "header-token" {
		t.Fatalf("expected token %q, got %q", "header-token", token)
	}
}

func TestAuthTokenRetrieve_ApiKeyParamWhenNoHeader(t *testing.T) {
	form := url.Values{"api_key": {"api-key-token"}}
	req := httptest.NewRequest(http.MethodPost, "http://example.com/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	token := AuthTokenRetrieve(req, false)
	if token != "api-key-token" {
		t.Fatalf("expected token %q, got %q", "api-key-token", token)
	}
}

func TestAuthTokenRetrieve_TokenParamFallback(t *testing.T) {
	form := url.Values{"token": {"param-token"}}
	req := httptest.NewRequest(http.MethodPost, "http://example.com/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	token := AuthTokenRetrieve(req, false)
	if token != "param-token" {
		t.Fatalf("expected token %q, got %q", "param-token", token)
	}
}

func TestAuthTokenRetrieve_IgnoresQueryStringTokens(t *testing.T) {
	// Tokens in URLs leak via logs, history and Referer — GET/HEAD
	// query params must never be accepted as credentials.
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		req := httptest.NewRequest(method, "http://example.com/?api_key=api-key-token&token=param-token", nil)
		if token := AuthTokenRetrieve(req, false); token != "" {
			t.Fatalf("expected empty token for %s query params, got %q", method, token)
		}
	}
}

func TestAuthTokenRetrieve_IgnoresQueryTokensOnPost(t *testing.T) {
	// On a POST, a token in the URL query is still a credential in a URL —
	// only body params may carry it.
	req := httptest.NewRequest(http.MethodPost, "http://example.com/?api_key=api-key-token&token=param-token", nil)

	if token := AuthTokenRetrieve(req, false); token != "" {
		t.Fatalf("expected empty token for POST query params, got %q", token)
	}
}

func TestAuthTokenRetrieve_ReturnsEmptyWhenNoSources(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	token := AuthTokenRetrieve(req, false)
	if token != "" {
		t.Fatalf("expected empty token, got %q", token)
	}
}

func TestAuthTokenRetrieveWithName_CustomCookieName(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.AddCookie(&http.Cookie{Name: "custom-auth", Value: "custom-cookie-token"})

	token := AuthTokenRetrieveWithName(req, true, "custom-auth")
	if token != "custom-cookie-token" {
		t.Fatalf("expected token %q, got %q", "custom-cookie-token", token)
	}
}

func TestAuthTokenRetrieveWithName_FallsBackToDefault(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	req.AddCookie(&http.Cookie{Name: types.CookieName, Value: "default-cookie-token"})

	token := AuthTokenRetrieveWithName(req, true, "")
	if token != "default-cookie-token" {
		t.Fatalf("expected token %q, got %q", "default-cookie-token", token)
	}
}
