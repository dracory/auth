package middlewares

import (
	"net/http"

	"github.com/dracory/api"
)

// CSRFConfig holds configuration for CSRF validation.
type CSRFConfig struct {
	Enabled  bool
	Validate func(r *http.Request) bool
}

// WithCSRF wraps an http.HandlerFunc with CSRF validation logic. If CSRF is
// disabled, the handler is executed directly. If validation fails, it responds
// with the same payload used previously in the handlers and does not call the
// wrapped handler.
func WithCSRF(cfg CSRFConfig, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.Enabled && cfg.Validate != nil && !cfg.Validate(r) {
			api.Respond(w, r, api.Forbidden("Invalid CSRF token"))
			return
		}

		next(w, r)
	}
}
