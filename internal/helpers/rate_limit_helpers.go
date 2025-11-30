package helpers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dracory/api"
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
			w.WriteHeader(http.StatusTooManyRequests)
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
			api.Respond(w, r, api.Error("Too many requests. Please try again later."))
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
		w.WriteHeader(http.StatusTooManyRequests)
		w.Header().Set("Retry-After", fmt.Sprintf("%.0f", result.RetryAfter.Seconds()))
		api.Respond(w, r, api.Error("Too many requests. Please try again later."))
		return false
	}

	return true
}

// GetClientIP extracts the client IP from the request.
// It checks X-Forwarded-For and X-Real-IP headers first, then falls back to RemoteAddr.
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (can contain multiple IPs, use the first one)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		for i, c := range xff {
			if c == ',' || c == ' ' {
				return xff[:i]
			}
		}
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

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
