package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	authtypes "github.com/dracory/auth/types"
	authutils "github.com/dracory/auth/utils"
)

type UserAuthOptions struct {
	UserIp    string
	UserAgent string
}

// Auth defines the structure for the authentication
type Auth struct {
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

// GetLogger returns the configured structured logger for this Auth instance.
// If no logger was explicitly provided, it falls back to slog.Default().
// Under normal library usage this method always returns a non-nil *slog.Logger.
func (a Auth) GetLogger() *slog.Logger {
	if a.logger != nil {
		return a.logger
	}
	return slog.Default()
}

// GetCurrentUserID returns the authenticated user ID stored in the request
// context, or an empty string if no user ID is attached.
func (a Auth) GetCurrentUserID(r *http.Request) string {
	authenticatedUserID := r.Context().Value(AuthenticatedUserID{})
	if authenticatedUserID == nil {
		return ""
	}
	return authenticatedUserID.(string)
}

func (a Auth) LinkApiLogin() string {
	return link(a.endpoint, PathApiLogin)
}

func (a Auth) LinkApiLoginCodeVerify() string {
	return link(a.endpoint, PathApiLoginCodeVerify)
}

func (a Auth) LinkApiLogout() string {
	return link(a.endpoint, PathApiLogout)
}

func (a Auth) LinkApiRegister() string {
	return link(a.endpoint, PathApiRegister)
}

func (a Auth) LinkApiRegisterCodeVerify() string {
	return link(a.endpoint, PathApiRegisterCodeVerify)
}

func (a Auth) LinkApiPasswordRestore() string {
	return link(a.endpoint, PathApiRestorePassword)
}

func (a Auth) LinkApiPasswordReset() string {
	return link(a.endpoint, PathApiResetPassword)
}

func (a Auth) LinkLogin() string {
	return link(a.endpoint, PathLogin)
}

func (a Auth) LinkLoginCodeVerify() string {
	return link(a.endpoint, PathLoginCodeVerify)
}

func (a Auth) LinkLogout() string {
	return link(a.endpoint, PathLogout)
}

func (a Auth) LinkPasswordRestore() string {
	return link(a.endpoint, PathPasswordRestore)
}

// LinkPasswordReset - returns the password reset URL
func (a Auth) LinkPasswordReset(token string) string {
	return link(a.endpoint, PathPasswordReset) + "?t=" + token
}

// LinkRegister - returns the registration URL
func (a Auth) LinkRegister() string {
	return link(a.endpoint, PathRegister)
}

// LinkRegisterCodeVerify - returns the registration code verification URL
func (a Auth) LinkRegisterCodeVerify() string {
	return link(a.endpoint, PathRegisterCodeVerify)
}

// LinkRedirectOnSuccess - returns the URL to where the user will be redirected after successful registration
func (a Auth) LinkRedirectOnSuccess() string {
	return a.urlRedirectOnSuccess
}

// link - creates the final URL by combining the provided endpoint with the provided URL
func link(endpoint, uri string) string {
	if strings.HasSuffix(endpoint, "/") {
		return endpoint + uri
	} else {
		return endpoint + "/" + uri
	}
}

// RegistrationEnable - enables registration
func (a *Auth) RegistrationEnable() {
	a.enableRegistration = true
}

// RegistrationDisable - disables registration
func (a *Auth) RegistrationDisable() {
	a.enableRegistration = false
}
