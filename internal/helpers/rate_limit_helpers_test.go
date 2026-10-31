package helpers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dracory/auth/utils"
)

func TestCheckRateLimit_DisabledAlwaysAllows(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	allowed := CheckRateLimit(w, req, "/login", true, nil, nil, nil)
	if !allowed {
		t.Fatalf("expected request to be allowed when rate limiting is disabled")
	}

	if w.Result().StatusCode != 0 && w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected no error status when rate limiting is disabled, got %d", w.Result().StatusCode)
	}
}

func TestCheckRateLimit_UsesCustomFunction(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	w := httptest.NewRecorder()

	custom := func(ip, endpoint string) (bool, time.Duration, error) {
		if endpoint != "/login" {
			t.Fatalf("unexpected endpoint: %s", endpoint)
		}
		return false, 5 * time.Second, nil
	}

	allowed := CheckRateLimit(w, req, "/login", false, custom, nil, nil)
	if allowed {
		t.Fatalf("expected request to be blocked by custom rate limiter")
	}

	res := w.Result()
	if res.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, res.StatusCode)
	}
}

func TestCheckRateLimit_DefaultLimiter_AllowsAndBlocks(t *testing.T) {
	limiter := utils.NewInMemoryRateLimiter(1, time.Second, time.Second)
	defer limiter.Stop()

	endpoint := "/login"

	// First request should pass
	req1 := httptest.NewRequest(http.MethodPost, endpoint, nil)
	w1 := httptest.NewRecorder()
	allowed1 := CheckRateLimit(w1, req1, endpoint, false, nil, limiter, nil)
	if !allowed1 {
		t.Fatalf("expected first request to be allowed")
	}

	// Second request from same IP/endpoint should be blocked
	req2 := httptest.NewRequest(http.MethodPost, endpoint, nil)
	w2 := httptest.NewRecorder()
	allowed2 := CheckRateLimit(w2, req2, endpoint, false, nil, limiter, nil)
	if allowed2 {
		t.Fatalf("expected second request to be rate limited")
	}

	if w2.Result().StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected status %d for rate limited request, got %d", http.StatusTooManyRequests, w2.Result().StatusCode)
	}
}

func TestGetClientIP_IgnoresSpoofableHeaders(t *testing.T) {
	// X-Forwarded-For and X-Real-IP are client-controlled and must NOT be
	// used for rate limiting — RemoteAddr always wins.
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "4.4.4.4:12345"
	req1.Header.Set("X-Forwarded-For", "1.1.1.1, 2.2.2.2")
	if ip := GetClientIP(req1); ip != "4.4.4.4" {
		t.Fatalf("expected RemoteAddr, got %q", ip)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "4.4.4.4:12345"
	req2.Header.Set("X-Real-IP", "3.3.3.3")
	if ip := GetClientIP(req2); ip != "4.4.4.4" {
		t.Fatalf("expected RemoteAddr, got %q", ip)
	}

	// RemoteAddr without a port is returned as-is
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.RemoteAddr = "4.4.4.4"
	if ip := GetClientIP(req3); ip != "4.4.4.4" {
		t.Fatalf("expected IP without port from RemoteAddr, got %q", ip)
	}
}
