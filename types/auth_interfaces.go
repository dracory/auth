package types

import (
	"context"
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
	IsPasswordless() bool
	IsVerificationEnabled() bool

	// Middlewares for protecting or enriching routes.
	WebAuthOrRedirectMiddleware(next http.Handler) http.Handler
	// ApiAuthOrErrorMiddleware(next http.Handler) http.Handler
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

	// ======================================================================
	// Accessors (Setters and Getters)
	// ======================================================================

	GetEndpoint() string
	SetEndpoint(endpoint string)

	GetLogger() *slog.Logger
	SetLogger(logger *slog.Logger)

	GetLayout() func(content string) string
	SetLayout(layout func(content string) string)

	GetFuncTemporaryKeyGet() func(key string) (string, error)
	SetFuncTemporaryKeyGet(fn func(key string) (string, error))

	GetFuncTemporaryKeySet() func(key string, value string, expiresSeconds int) error
	SetFuncTemporaryKeySet(fn func(key string, value string, expiresSeconds int) error)

	GetUseCookies() bool
	SetUseCookies(useCookies bool)

	GetFuncUserFindByAuthToken() func(ctx context.Context, token string, options UserAuthOptions) (userID string, err error)
	SetFuncUserFindByAuthToken(fn func(ctx context.Context, token string, options UserAuthOptions) (userID string, err error))

	// Additional accessors used by internal API flows.
	GetDisableRateLimit() bool
	SetDisableRateLimit(disable bool)

	GetPasswordStrength() *PasswordStrengthConfig
	SetPasswordStrength(cfg *PasswordStrengthConfig)

	GetPasswordlessUserRegister() func(ctx context.Context, email, firstName, lastName string, options UserAuthOptions) error
	SetPasswordlessUserRegister(fn func(ctx context.Context, email, firstName, lastName string, options UserAuthOptions) error)

	GetFuncUserRegister() func(ctx context.Context, username, password, firstName, lastName string, options UserAuthOptions) error
	SetFuncUserRegister(fn func(ctx context.Context, username, password, firstName, lastName string, options UserAuthOptions) error)

	GetFuncUserPasswordChange() func(ctx context.Context, userID, password string, options UserAuthOptions) error
	SetFuncUserPasswordChange(fn func(ctx context.Context, userID, password string, options UserAuthOptions) error)

	GetFuncUserLogout() func(ctx context.Context, userID string, options UserAuthOptions) error
	SetFuncUserLogout(fn func(ctx context.Context, userID string, options UserAuthOptions) error)

	GetPasswordlessUserFindByEmail() func(ctx context.Context, email string, options UserAuthOptions) (string, error)
	SetPasswordlessUserFindByEmail(fn func(ctx context.Context, email string, options UserAuthOptions) (string, error))

	GetFuncUserFindByUsername() func(ctx context.Context, username, firstName, lastName string, options UserAuthOptions) (string, error)
	SetFuncUserFindByUsername(fn func(ctx context.Context, username, firstName, lastName string, options UserAuthOptions) (string, error))

	GetFuncEmailTemplatePasswordRestore() func(ctx context.Context, userID string, passwordRestoreLink string, options UserAuthOptions) string
	SetFuncEmailTemplatePasswordRestore(fn func(ctx context.Context, userID string, passwordRestoreLink string, options UserAuthOptions) string)

	GetFuncEmailSend() func(ctx context.Context, userID, emailSubject, emailBody string) error
	SetFuncEmailSend(fn func(ctx context.Context, userID, emailSubject, emailBody string) error)

	GetPasswordlessFuncEmailTemplateLoginCode() func(ctx context.Context, email string, passwordRestoreLink string, options UserAuthOptions) string
	SetPasswordlessFuncEmailTemplateLoginCode(fn func(ctx context.Context, email string, passwordRestoreLink string, options UserAuthOptions) string)

	GetPasswordlessFuncEmailTemplateRegisterCode() func(ctx context.Context, email string, passwordRestoreLink string, options UserAuthOptions) string
	SetPasswordlessFuncEmailTemplateRegisterCode(fn func(ctx context.Context, email string, passwordRestoreLink string, options UserAuthOptions) string)

	GetPasswordlessFuncEmailSend() func(ctx context.Context, email string, emailSubject, emailBody string) error
	SetPasswordlessFuncEmailSend(fn func(ctx context.Context, email string, emailSubject, emailBody string) error)

	GetFuncUserStoreAuthToken() func(ctx context.Context, token, userID string, options UserAuthOptions) error
	SetFuncUserStoreAuthToken(fn func(ctx context.Context, token, userID string, options UserAuthOptions) error)

	SetAuthCookie(w http.ResponseWriter, r *http.Request, token string)
	RemoveAuthCookie(w http.ResponseWriter, r *http.Request)

	// Final authentication step helpers used by internal API flows.
	AuthenticateViaUsername(w http.ResponseWriter, r *http.Request, email, firstName, lastName string)
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

	// Generic username/password registration helper used by API layers.
	RegisterUserWithPassword(ctx context.Context, email, password, firstName, lastName string, options UserAuthOptions) (successMessage, token, errorMessage string)

	// Generic username/password login helper used by API layers.
	LoginUserWithPassword(ctx context.Context, email, password string, options UserAuthOptions) (successMessage, token, errorMessage string)
}

// AuthPasswordlessInterface represents passwordless authentication flows.
// It extends the shared interface with login/verification code helpers.
type AuthPasswordlessInterface interface {
	AuthSharedInterface

	// Passwordless-only URL helpers.
	LinkLoginCodeVerify() string
	LinkApiLoginCodeVerify() string
}
