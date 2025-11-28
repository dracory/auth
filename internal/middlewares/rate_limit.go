package middlewares

import "net/http"

// RateLimitConfig holds configuration for rate limiting.
type RateLimitConfig struct {
	// Check should perform the actual rate limit check. It should return true if
	// the request is allowed to proceed. If it returns false, it is assumed to
	// have already written the appropriate HTTP response.
	Check func(w http.ResponseWriter, r *http.Request, endpoint string) bool

	// Endpoint is the logical endpoint name used for rate limiting, e.g.
	// "login", "register", etc.
	Endpoint string
}

// WithRateLimit wraps an http.HandlerFunc with rate limiting logic.
func WithRateLimit(cfg RateLimitConfig, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.Check != nil {
			if ok := cfg.Check(w, r, cfg.Endpoint); !ok {
				return
			}
		}

		next(w, r)
	}
}
