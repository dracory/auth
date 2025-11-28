package auth

import (
	"fmt"
	"net/http"

	"github.com/dracory/api"
)

// checkRateLimit verifies if a request should be allowed based on rate limiting rules
// Returns true if allowed, false if rate limited
func (a authImplementation) checkRateLimit(w http.ResponseWriter, r *http.Request, endpoint string) bool {
	// If rate limiting is disabled, allow all requests
	if a.disableRateLimit {
		return true
	}

	ip := getClientIP(r)

	// Use custom rate limit function if provided
	if a.funcCheckRateLimit != nil {
		allowed, retryAfter, err := a.funcCheckRateLimit(ip, endpoint)
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
	if a.rateLimiter == nil {
		// This shouldn't happen if properly initialized, but fail open for safety
		return true
	}

	result := a.rateLimiter.Check(ip, endpoint)
	if !result.Allowed {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Header().Set("Retry-After", fmt.Sprintf("%.0f", result.RetryAfter.Seconds()))
		api.Respond(w, r, api.Error("Too many requests. Please try again later."))
		return false
	}

	return true
}

// getClientIP extracts the client IP from the request
// Checks X-Forwarded-For and X-Real-IP headers first, then falls back to RemoteAddr
func getClientIP(r *http.Request) string {
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
