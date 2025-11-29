package types

import (
	"log/slog"
	"net/http"
)

// AuthSharedInterface defines the common behavior shared by all auth modes.
// It includes routing helpers, middleware, current user access, and the
// primary login/register URL helpers.
type AuthSharedInterface interface {
	// Router returns an HTTP mux that serves all auth routes.
	Router() *http.ServeMux

	IsRegistrationEnabled() bool

	// Middlewares for protecting or enriching routes.
	WebAuthOrRedirectMiddleware(next http.Handler) http.Handler
	ApiAuthOrErrorMiddleware(next http.Handler) http.Handler
	WebAppendUserIdIfExistsMiddleware(next http.Handler) http.Handler

	// Current user lookup from the request context.
	GetCurrentUserID(r *http.Request) string

	// Web URL helpers
	LinkLogin() string
	LinkLogout() string
	LinkRegister() string
	LinkRegisterCodeVerify() string
	LinkRedirectOnSuccess() string

	// API URL helpers
	LinkApiLogin() string
	LinkApiLogout() string
	LinkApiRegister() string
	LinkApiRegisterCodeVerify() string

	// Accessors
	GetEndpoint() string
	SetEndpoint(endpoint string)

	GetLogger() *slog.Logger
	SetLogger(logger *slog.Logger)

	GetLayout() func(content string) string
	SetLayout(layout func(content string) string)
}

// AuthPasswordInterface represents username/password based authentication.
// It extends the shared interface with password-reset specific helpers.
type AuthPasswordInterface interface {
	AuthSharedInterface

	// Password reset URLs (web and API).
	LinkPasswordRestore() string
	LinkPasswordReset(token string) string
	LinkApiPasswordRestore() string
	LinkApiPasswordReset() string
}

// AuthPasswordlessInterface represents passwordless authentication flows.
// It extends the shared interface with login/verification code helpers.
type AuthPasswordlessInterface interface {
	AuthSharedInterface

	// Passwordless-only URL helpers.
	LinkLoginCodeVerify() string
	LinkApiLoginCodeVerify() string
}
