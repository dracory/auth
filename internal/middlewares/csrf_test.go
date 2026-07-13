package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithCSRF_Disabled(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	cfg := CSRFConfig{Enabled: false}
	handler := WithCSRF(cfg, next)

	req, _ := http.NewRequest("POST", "/api/login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called when CSRF is disabled")
	}
}

func TestWithCSRF_EnabledValidToken(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	cfg := CSRFConfig{
		Enabled:  true,
		Validate: func(r *http.Request) bool { return true },
	}
	handler := WithCSRF(cfg, next)

	req, _ := http.NewRequest("POST", "/api/login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called when CSRF validation passes")
	}
}

func TestWithCSRF_EnabledInvalidToken(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	cfg := CSRFConfig{
		Enabled:  true,
		Validate: func(r *http.Request) bool { return false },
	}
	handler := WithCSRF(cfg, next)

	req, _ := http.NewRequest("POST", "/api/login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if called {
		t.Fatal("expected next handler NOT to be called when CSRF validation fails")
	}
	body := rec.Body.String()
	if !strings.Contains(body, "forbidden") {
		t.Fatalf("expected response body to contain 'forbidden', got %q", body)
	}
}

func TestWithCSRF_EnabledNilValidate(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	cfg := CSRFConfig{
		Enabled:  true,
		Validate: nil,
	}
	handler := WithCSRF(cfg, next)

	req, _ := http.NewRequest("POST", "/api/login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called when Validate is nil (no-op)")
	}
}
