package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithRateLimit_NoCheckFunc(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	cfg := RateLimitConfig{Endpoint: "login"}
	handler := WithRateLimit(cfg, next)

	req, _ := http.NewRequest("POST", "/api/login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called when Check is nil")
	}
}

func TestWithRateLimit_CheckPasses(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	cfg := RateLimitConfig{
		Endpoint: "login",
		Check:    func(w http.ResponseWriter, r *http.Request, endpoint string) bool { return true },
	}
	handler := WithRateLimit(cfg, next)

	req, _ := http.NewRequest("POST", "/api/login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected next handler to be called when rate limit check passes")
	}
}

func TestWithRateLimit_CheckFails(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	checkCalled := false
	cfg := RateLimitConfig{
		Endpoint: "login",
		Check: func(w http.ResponseWriter, r *http.Request, endpoint string) bool {
			checkCalled = true
			if endpoint != "login" {
				t.Fatalf("expected endpoint 'login', got %q", endpoint)
			}
			w.WriteHeader(http.StatusTooManyRequests)
			return false
		},
	}
	handler := WithRateLimit(cfg, next)

	req, _ := http.NewRequest("POST", "/api/login", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !checkCalled {
		t.Fatal("expected Check function to be called")
	}
	if called {
		t.Fatal("expected next handler NOT to be called when rate limit check fails")
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, rec.Code)
	}
}
