package helpers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dracory/api"
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
)

// CheckRateLimit verifies if a request should be allowed based on rate limiting rules.
// It returns true if allowed, false if rate limited.
func CheckRateLimit(
	w http.ResponseWriter,
	r *http.Request,
	endpoint string,
	disableRateLimit bool,
	customCheck func(ip string, endpoint string) (allowed bool, retryAfter time.Duration, err error),
	limiter *utils.InMemoryRateLimiter,
	hooks types.ObservabilityHooks,
) bool {
	// If rate limiting is disabled, allow all requests
	if disableRateLimit {
		return true
	}

	ip := GetClientIP(r)

	// Use custom rate limit function if provided
	if customCheck != nil {
		allowed, retryAfter, err := customCheck(ip, endpoint)
		if err != nil {
			// Log error but don't block request on rate limiter errors
			// In production, you might want to handle this differently
			return true
		}
		if !allowed {
			if hooks != nil {
				hooks.RecordRateLimitHit(endpoint, ip)
			}
			w.WriteHeader(http.StatusTooManyRequests)
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
			api.Respond(w, r, api.Error(types.MsgTooManyRequests))
			return false
		}
		return true
	}

	// Use default in-memory rate limiter
	if limiter == nil {
		// This shouldn't happen if properly initialized, but fail open for safety
		return true
	}

	result := limiter.Check(ip, endpoint)
	if !result.Allowed {
		if hooks != nil {
			hooks.RecordRateLimitHit(endpoint, ip)
		}
		w.WriteHeader(http.StatusTooManyRequests)
		w.Header().Set("Retry-After", fmt.Sprintf("%.0f", result.RetryAfter.Seconds()))
		api.Respond(w, r, api.Error(types.MsgTooManyRequests))
		return false
	}

	return true
}

// GetClientIP returns the client IP for rate limiting.
// It deliberately uses RemoteAddr and ignores X-Forwarded-For / X-Real-IP:
// those headers are client-controlled and would let an attacker rotate the
// header on each request to obtain a fresh rate-limit bucket, silently
// disabling brute-force protection. Deployments behind a trusted proxy that
// need XFF-aware limits should provide a custom FuncCheckRateLimit that
// resolves the real client IP from their known proxy chain.
func GetClientIP(r *http.Request) string {

	// Fall back to RemoteAddr
	// RemoteAddr is in format "IP:port", so we need to strip the port
	remoteAddr := r.RemoteAddr
	for i := len(remoteAddr) - 1; i >= 0; i-- {
		if remoteAddr[i] == ':' {
			return remoteAddr[:i]
		}
	}

	return remoteAddr
}
