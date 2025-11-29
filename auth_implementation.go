package auth

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/dracory/auth/internal/links"
	authtypes "github.com/dracory/auth/types"
	authutils "github.com/dracory/auth/utils"
)

type UserAuthOptions struct {
	UserIp    string
	UserAgent string
}

// Auth defines the structure for the authentication
type authImplementation struct {
	endpoint string

	// enableRegistration enables the registration page and endpoint
	enableRegistration bool

	// urlRedirectOnSuccess the endpoint to return to on success
	urlRedirectOnSuccess string

	// ===== START: shared by all implementations
	funcLayout              func(content string) string
	funcTemporaryKeyGet     func(key string) (value string, err error)
	funcTemporaryKeySet     func(key string, value string, expiresSeconds int) (err error)
	funcUserFindByAuthToken func(ctx context.Context, token string, options UserAuthOptions) (userID string, err error)
	funcUserLogout          func(ctx context.Context, userID string, options UserAuthOptions) (err error)
	funcUserStoreAuthToken  func(ctx context.Context, token string, userID string, options UserAuthOptions) error
	// ===== END: shared by all implementations

	// ===== START: username(email) and password options
	enableVerification               bool
	funcEmailTemplatePasswordRestore func(ctx context.Context, userID string, passwordRestoreLink string, options UserAuthOptions) string // optional
	funcEmailTemplateRegisterCode    func(ctx context.Context, email string, passwordRestoreLink string, options UserAuthOptions) string  // optional
	funcEmailSend                    func(ctx context.Context, userID string, emailSubject string, emailBody string) (err error)
	funcUserLogin                    func(ctx context.Context, username string, password string, options UserAuthOptions) (userID string, err error)
	funcUserPasswordChange           func(ctx context.Context, username string, newPassword string, options UserAuthOptions) (err error)
	funcUserRegister                 func(ctx context.Context, username string, password string, first_name string, last_name string, options UserAuthOptions) (err error)
	funcUserFindByUsername           func(ctx context.Context, username string, first_name string, last_name string, options UserAuthOptions) (userID string, err error)
	passwordStrength                 *authtypes.PasswordStrengthConfig
	// ===== END: username(email) and password options

	// ===== START: passwordless options
	passwordless                              bool
	passwordlessFuncUserFindByEmail           func(ctx context.Context, email string, options UserAuthOptions) (userID string, err error)
	passwordlessFuncEmailTemplateLoginCode    func(ctx context.Context, email string, passwordRestoreLink string, options UserAuthOptions) string // optional
	passwordlessFuncEmailTemplateRegisterCode func(ctx context.Context, email string, passwordRestoreLink string, options UserAuthOptions) string // optional
	passwordlessFuncEmailSend                 func(ctx context.Context, email string, emailSubject string, emailBody string) (err error)
	passwordlessFuncUserRegister              func(ctx context.Context, email string, firstName string, lastName string, options UserAuthOptions) (err error)
	// ===== END: passwordless options

	// ===== START: rate limiting
	disableRateLimit   bool
	funcCheckRateLimit func(ip string, endpoint string) (allowed bool, retryAfter time.Duration, err error)
	rateLimiter        *authutils.InMemoryRateLimiter
	// ===== END: rate limiting

	cookieConfig CookieConfig

	// ===== START: CSRF Protection
	enableCSRFProtection  bool
	csrfSecret            string
	funcCSRFTokenGenerate func(r *http.Request) string
	funcCSRFTokenValidate func(r *http.Request) bool
	// ===== END: CSRF Protection

	// labelUsername   string
	useCookies      bool
	useLocalStorage bool
	logger          *slog.Logger
}

func (a authImplementation) GetEndpoint() string {
	return a.endpoint
}

func (a *authImplementation) SetEndpoint(endpoint string) {
	a.endpoint = endpoint
}

func (a authImplementation) GetLayout() func(content string) string {
	return a.funcLayout
}

func (a *authImplementation) SetLayout(layout func(content string) string) {
	a.funcLayout = layout
}

func (a authImplementation) IsRegistrationEnabled() bool {
	return a.enableRegistration
}

func (a authImplementation) IsPasswordless() bool {
	return a.passwordless
}

func (a authImplementation) IsVerificationEnabled() bool {
	return a.enableVerification
}

// GetLogger returns the configured structured logger for this Auth instance.
// If no logger was explicitly provided, it falls back to slog.Default().
// Under normal library usage this method always returns a non-nil *slog.Logger.
func (a authImplementation) GetLogger() *slog.Logger {
	if a.logger != nil {
		return a.logger
	}
	return slog.Default()
}

func (a *authImplementation) SetLogger(logger *slog.Logger) {
	a.logger = logger
}

// GetCurrentUserID returns the authenticated user ID stored in the request
// context, or an empty string if no user ID is attached.
func (a authImplementation) GetCurrentUserID(r *http.Request) string {
	authenticatedUserID := r.Context().Value(AuthenticatedUserID{})
	if authenticatedUserID == nil {
		return ""
	}
	return authenticatedUserID.(string)
}

func (a authImplementation) LinkApiLogin() string {
	return links.ApiLogin(a.endpoint)
}

func (a authImplementation) LinkApiLoginCodeVerify() string {
	return links.ApiLoginCodeVerify(a.endpoint)
}

func (a authImplementation) LinkApiLogout() string {
	return links.ApiLogout(a.endpoint)
}

func (a authImplementation) LinkApiRegister() string {
	return links.ApiRegister(a.endpoint)
}

func (a authImplementation) LinkApiRegisterCodeVerify() string {
	return links.ApiRegisterCodeVerify(a.endpoint)
}

func (a authImplementation) LinkApiPasswordRestore() string {
	return links.ApiPasswordRestore(a.endpoint)
}

func (a authImplementation) LinkApiPasswordReset() string {
	return links.ApiPasswordReset(a.endpoint)
}

func (a authImplementation) LinkLogin() string {
	return links.Login(a.endpoint)
}

func (a authImplementation) LinkLoginCodeVerify() string {
	return links.LoginCodeVerify(a.endpoint)
}

func (a authImplementation) LinkLogout() string {
	return links.Logout(a.endpoint)
}

func (a authImplementation) LinkPasswordRestore() string {
	return links.PasswordRestore(a.endpoint)
}

// LinkPasswordReset - returns the password reset URL
func (a authImplementation) LinkPasswordReset(token string) string {
	return links.PasswordReset(a.endpoint) + "?t=" + token
}

// LinkRegister - returns the registration URL
func (a authImplementation) LinkRegister() string {
	return links.Register(a.endpoint)
}

// LinkRegisterCodeVerify - returns the registration code verification URL
func (a authImplementation) LinkRegisterCodeVerify() string {
	return links.RegisterCodeVerify(a.endpoint)
}

// LinkRedirectOnSuccess - returns the URL to where the user will be redirected after successful registration
func (a authImplementation) LinkRedirectOnSuccess() string {
	return a.urlRedirectOnSuccess
}

// RegistrationEnable - enables registration
func (a *authImplementation) RegistrationEnable() {
	a.enableRegistration = true
}

// RegistrationDisable - disables registration
func (a *authImplementation) RegistrationDisable() {
	a.enableRegistration = false
}
