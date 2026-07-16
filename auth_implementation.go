package auth

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/middlewares"
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
)

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
	funcUserFindByAuthToken func(ctx context.Context, token string, options types.UserAuthOptions) (userID string, err error)
	funcUserLogout          func(ctx context.Context, userID string, options types.UserAuthOptions) (err error)
	funcUserStoreAuthToken  func(ctx context.Context, token string, userID string, options types.UserAuthOptions) error
	// ===== END: shared by all implementations

	// ===== START: username(email) and password options
	enableVerification               bool
	funcEmailTemplatePasswordRestore func(ctx context.Context, userID string, passwordRestoreLink string, options types.UserAuthOptions) string // optional
	funcEmailTemplateRegisterCode    func(ctx context.Context, email string, registerLink string, options types.UserAuthOptions) string         // optional
	funcEmailSend                    func(ctx context.Context, userID string, emailSubject string, emailBody string) (err error)
	funcUserLogin                    func(ctx context.Context, username string, password string, options types.UserAuthOptions) (userID string, err error)
	funcUserPasswordChange           func(ctx context.Context, username string, newPassword string, options types.UserAuthOptions) (err error)
	funcUserRegister                 func(ctx context.Context, username string, password string, firstName string, lastName string, options types.UserAuthOptions) (err error)
	funcUserFindByUsername           func(ctx context.Context, username string, firstName string, lastName string, options types.UserAuthOptions) (userID string, err error)
	passwordStrength                 *types.PasswordStrengthConfig
	// ===== END: username(email) and password options

	// ===== START: passwordless options
	passwordless                              bool
	passwordlessFuncUserFindByEmail           func(ctx context.Context, email string, options types.UserAuthOptions) (userID string, err error)
	passwordlessFuncEmailTemplateLoginCode    func(ctx context.Context, email string, loginLink string, options types.UserAuthOptions) string    // optional
	passwordlessFuncEmailTemplateRegisterCode func(ctx context.Context, email string, registerLink string, options types.UserAuthOptions) string // optional
	passwordlessFuncEmailSend                 func(ctx context.Context, email string, emailSubject string, emailBody string) (err error)
	passwordlessFuncUserRegister              func(ctx context.Context, email string, firstName string, lastName string, options types.UserAuthOptions) (err error)
	// ===== END: passwordless options

	// ===== START: rate limiting
	disableRateLimit   bool
	funcCheckRateLimit func(ip string, endpoint string) (allowed bool, retryAfter time.Duration, err error)
	rateLimiter        *utils.InMemoryRateLimiter
	// ===== END: rate limiting

	// ===== START: cookie config
	cookieConfig CookieConfig
	cookieName   string
	useCookies   bool
	// ===== END: cookie config

	// ===== START: CSRF Protection
	enableCSRFProtection  bool
	csrfSecret            string
	funcCSRFTokenGenerate func(r *http.Request) string
	funcCSRFTokenValidate func(r *http.Request) bool
	// ===== END: CSRF Protection

	// ===== START: impersonation
	enableImpersonation    bool
	funcCanImpersonate     func(ctx context.Context, adminUserID string, targetUserID string) (bool, error)
	funcImpersonationStart func(ctx context.Context, adminUserID string, targetUserID string) error
	funcImpersonationStop  func(ctx context.Context, adminUserID string, targetUserID string) error
	// ===== END: impersonation

	// ===== START: AuthKnight options
	authKnight                    bool
	authKnightFuncUserFindByEmail func(ctx context.Context, email string, options types.UserAuthOptions) (userID string, err error)
	authKnightFuncUserRegister    func(ctx context.Context, email string, firstName string, lastName string, options types.UserAuthOptions) (userID string, err error)
	authKnightFuncRedirectURL     func(ctx context.Context, userID string) string
	authKnightHTTPTimeout         time.Duration
	// ===== END: AuthKnight options

	// labelUsername   string
	useLocalStorage bool
	logger          *slog.Logger

	observabilityHooks types.ObservabilityHooks
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

// SetLogger allows overriding the internal logger used by the Auth instance.
// This is useful for integrating with application logging systems or for testing.
func (a *authImplementation) SetLogger(logger *slog.Logger) {
	a.logger = logger
}

// GetObservabilityHooks returns the configured observability hooks for this Auth instance.
// If no hooks were explicitly provided, it falls back to NoopObservabilityHooks.
// Under normal library usage this method always returns a non-nil types.ObservabilityHooks.
func (a authImplementation) GetObservabilityHooks() types.ObservabilityHooks {
	if a.observabilityHooks == nil {
		return types.NoopObservabilityHooks{}
	}
	return a.observabilityHooks
}

// SetObservabilityHooks allows overriding the internal observability hooks used by the Auth instance.
// This is useful for integrating with application observability systems or for testing.
func (a *authImplementation) SetObservabilityHooks(hooks types.ObservabilityHooks) {
	a.observabilityHooks = hooks
}

// GetCurrentUserID returns the authenticated user ID stored in the request
// context, or an empty string if no user ID is attached.
func (a authImplementation) GetCurrentUserID(r *http.Request) string {
	authenticatedUserID := r.Context().Value(types.AuthenticatedUserID{})
	if authenticatedUserID == nil {
		return ""
	}
	return authenticatedUserID.(string)
}

// GetUseCookies returns whether the Auth instance uses cookies for authentication.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetUseCookies() bool {
	return a.useCookies
}

// GetFuncUserFindByAuthToken returns the function used to find a user by authentication token.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetFuncUserFindByAuthToken() func(ctx context.Context, token string, options types.UserAuthOptions) (string, error) {
	return a.funcUserFindByAuthToken
}

// SetUseCookies allows overriding whether the Auth instance uses cookies for authentication.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetUseCookies(useCookies bool) {
	a.useCookies = useCookies
}

// GetCookieName returns the cookie name used by the Auth instance.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetCookieName() string {
	if a.cookieName != "" {
		return a.cookieName
	}
	return types.CookieName
}

// SetCookieName allows overriding the cookie name used by the Auth instance.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetCookieName(name string) {
	a.cookieName = name
}

// SetFuncUserFindByAuthToken allows overriding the function used to find a user by authentication token.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetFuncUserFindByAuthToken(fn func(ctx context.Context, token string, options types.UserAuthOptions) (string, error)) {
	a.funcUserFindByAuthToken = fn
}

// GetDisableRateLimit returns whether the Auth instance disables rate limiting.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetDisableRateLimit() bool {
	return a.disableRateLimit
}

// SetDisableRateLimit allows overriding whether the Auth instance disables rate limiting.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetDisableRateLimit(disable bool) {
	a.disableRateLimit = disable
}

// GetPasswordStrength returns the password strength configuration for the Auth instance.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetPasswordStrength() *types.PasswordStrengthConfig {
	return a.passwordStrength
}

// SetPasswordStrength allows overriding the password strength configuration for the Auth instance.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetPasswordStrength(cfg *types.PasswordStrengthConfig) {
	a.passwordStrength = cfg
}

// GetFuncUserLogin returns the function used to authenticate a user by username and password.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetFuncUserLogin() func(ctx context.Context, username, password string, options types.UserAuthOptions) (string, error) {
	return a.funcUserLogin
}

// SetFuncUserLogin allows overriding the function used to authenticate a user by username and password.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetFuncUserLogin(fn func(ctx context.Context, username, password string, options types.UserAuthOptions) (string, error)) {
	a.funcUserLogin = fn
}

// GetPasswordlessUserRegister returns the function used to register a user without a password.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetPasswordlessUserRegister() func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) error {
	return a.passwordlessFuncUserRegister
}

// SetPasswordlessUserRegister allows overriding the function used to register a user without a password.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetPasswordlessUserRegister(fn func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) error) {
	a.passwordlessFuncUserRegister = fn
}

// GetFuncUserRegister returns the function used to register a user.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetFuncUserRegister() func(ctx context.Context, username, password, firstName, lastName string, options types.UserAuthOptions) error {
	return a.funcUserRegister
}

// SetFuncUserRegister allows overriding the function used to register a user.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetFuncUserRegister(fn func(ctx context.Context, username, password, firstName, lastName string, options types.UserAuthOptions) error) {
	a.funcUserRegister = fn
}

// GetFuncUserPasswordChange returns the function used to change a user's password.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetFuncUserPasswordChange() func(ctx context.Context, userID, password string, options types.UserAuthOptions) error {
	return a.funcUserPasswordChange
}

// SetFuncUserPasswordChange allows overriding the function used to change a user's password.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetFuncUserPasswordChange(fn func(ctx context.Context, userID, password string, options types.UserAuthOptions) error) {
	a.funcUserPasswordChange = fn
}

// GetFuncUserLogout returns the function used to log out a user.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetFuncUserLogout() func(ctx context.Context, userID string, options types.UserAuthOptions) error {
	return a.funcUserLogout
}

// SetFuncUserLogout allows overriding the function used to log out a user.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetFuncUserLogout(fn func(ctx context.Context, userID string, options types.UserAuthOptions) error) {
	a.funcUserLogout = fn
}

// GetPasswordlessUserFindByEmail returns the function used to find a user by email for passwordless authentication.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetPasswordlessUserFindByEmail() func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
	return a.passwordlessFuncUserFindByEmail
}

// SetPasswordlessUserFindByEmail allows overriding the function used to find a user by email for passwordless authentication.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetPasswordlessUserFindByEmail(fn func(ctx context.Context, email string, options types.UserAuthOptions) (string, error)) {
	a.passwordlessFuncUserFindByEmail = fn
}

// GetPasswordlessFuncEmailTemplateLoginCode returns the function used to generate the email template for login code.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetPasswordlessFuncEmailTemplateLoginCode() func(ctx context.Context, email string, loginLink string, options types.UserAuthOptions) string {
	return a.passwordlessFuncEmailTemplateLoginCode
}

// SetPasswordlessFuncEmailTemplateLoginCode allows overriding the function used to generate the email template for login code.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetPasswordlessFuncEmailTemplateLoginCode(fn func(ctx context.Context, email string, loginLink string, options types.UserAuthOptions) string) {
	a.passwordlessFuncEmailTemplateLoginCode = fn
}

// GetPasswordlessFuncEmailTemplateRegisterCode returns the function used to generate the email template for register code.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetPasswordlessFuncEmailTemplateRegisterCode() func(ctx context.Context, email string, registerLink string, options types.UserAuthOptions) string {
	return a.passwordlessFuncEmailTemplateRegisterCode
}

// SetPasswordlessFuncEmailTemplateRegisterCode allows overriding the function used to generate the email template for register code.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetPasswordlessFuncEmailTemplateRegisterCode(fn func(ctx context.Context, email string, registerLink string, options types.UserAuthOptions) string) {
	a.passwordlessFuncEmailTemplateRegisterCode = fn
}

// GetPasswordlessFuncEmailSend returns the function used to send an email.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetPasswordlessFuncEmailSend() func(ctx context.Context, email string, emailSubject string, emailBody string) error {
	return a.passwordlessFuncEmailSend
}

// SetPasswordlessFuncEmailSend allows overriding the function used to send an email.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetPasswordlessFuncEmailSend(fn func(ctx context.Context, email string, emailSubject string, emailBody string) error) {
	a.passwordlessFuncEmailSend = fn
}

// GetFuncUserFindByUsername returns the function used to find a user by username.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetFuncUserFindByUsername() func(ctx context.Context, username, firstName, lastName string, options types.UserAuthOptions) (string, error) {
	return a.funcUserFindByUsername
}

// SetFuncUserFindByUsername allows overriding the function used to find a user by username.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetFuncUserFindByUsername(fn func(ctx context.Context, username, firstName, lastName string, options types.UserAuthOptions) (string, error)) {
	a.funcUserFindByUsername = fn
}

// GetFuncEmailTemplatePasswordRestore returns the function used to generate the email template for password restore.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetFuncEmailTemplatePasswordRestore() func(ctx context.Context, userID string, passwordRestoreLink string, options types.UserAuthOptions) string {
	return a.funcEmailTemplatePasswordRestore
}

// SetFuncEmailTemplatePasswordRestore allows overriding the function used to generate the email template for password restore.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetFuncEmailTemplatePasswordRestore(fn func(ctx context.Context, userID string, passwordRestoreLink string, options types.UserAuthOptions) string) {
	a.funcEmailTemplatePasswordRestore = fn
}

// GetFuncEmailTemplateRegisterCode returns the function used to generate the email template for register code.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetFuncEmailTemplateRegisterCode() func(ctx context.Context, email string, registerLink string, options types.UserAuthOptions) string {
	return a.funcEmailTemplateRegisterCode
}

// SetFuncEmailTemplateRegisterCode allows overriding the function used to generate the email template for register code.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetFuncEmailTemplateRegisterCode(fn func(ctx context.Context, email string, registerLink string, options types.UserAuthOptions) string) {
	a.funcEmailTemplateRegisterCode = fn
}

// GetFuncEmailSend returns the function used to send an email.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetFuncEmailSend() func(ctx context.Context, userID, emailSubject, emailBody string) error {
	return a.funcEmailSend
}

// SetFuncEmailSend allows overriding the function used to send an email.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetFuncEmailSend(fn func(ctx context.Context, userID, emailSubject, emailBody string) error) {
	a.funcEmailSend = fn
}

// GetFuncUserStoreAuthToken returns the function used to store an authentication token for a user.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetFuncUserStoreAuthToken() func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
	return a.funcUserStoreAuthToken
}

// SetFuncUserStoreAuthToken allows overriding the function used to store an authentication token for a user.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetFuncUserStoreAuthToken(fn func(ctx context.Context, token, userID string, options types.UserAuthOptions) error) {
	a.funcUserStoreAuthToken = fn
}

// SetAuthCookie sets the authentication cookie.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) SetAuthCookie(w http.ResponseWriter, r *http.Request, token string) {
	a.setAuthCookie(w, r, token)
}

// RemoveAuthCookie removes the authentication cookie.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) RemoveAuthCookie(w http.ResponseWriter, r *http.Request) {
	a.removeAuthCookie(w, r)
}

// AuthenticateViaUsername authenticates a user via username.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) AuthenticateViaUsername(w http.ResponseWriter, r *http.Request, email, firstName, lastName string) {
	a.authenticateViaUsername(w, r, email, firstName, lastName)
}

// GetFuncTemporaryKeyGet returns the function used to get a temporary key.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetFuncTemporaryKeyGet() func(key string) (string, error) {
	return a.funcTemporaryKeyGet
}

// SetFuncTemporaryKeyGet allows overriding the function used to get a temporary key.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetFuncTemporaryKeyGet(fn func(key string) (string, error)) {
	a.funcTemporaryKeyGet = fn
}

// GetFuncTemporaryKeySet returns the function used to set a temporary key.
// This is useful for integrating with application authentication systems or for testing.
func (a authImplementation) GetFuncTemporaryKeySet() func(key string, value string, expiresSeconds int) error {
	return a.funcTemporaryKeySet
}

// SetFuncTemporaryKeySet allows overriding the function used to set a temporary key.
// This is useful for integrating with application authentication systems or for testing.
func (a *authImplementation) SetFuncTemporaryKeySet(fn func(key string, value string, expiresSeconds int) error) {
	a.funcTemporaryKeySet = fn
}

// LinkApiLogin returns the API login endpoint URL.
func (a authImplementation) LinkApiLogin() string {
	return links.ApiLogin(a.endpoint)
}

// LinkApiLoginCodeVerify returns the API login code verify endpoint URL.
func (a authImplementation) LinkApiLoginCodeVerify() string {
	return links.ApiLoginCodeVerify(a.endpoint)
}

// LinkApiLogout returns the API logout endpoint URL.
func (a authImplementation) LinkApiLogout() string {
	return links.ApiLogout(a.endpoint)
}

// LinkApiRegister returns the API register endpoint URL.
func (a authImplementation) LinkApiRegister() string {
	return links.ApiRegister(a.endpoint)
}

// LinkApiRegisterCodeVerify returns the API register code verify endpoint URL.
func (a authImplementation) LinkApiRegisterCodeVerify() string {
	return links.ApiRegisterCodeVerify(a.endpoint)
}

// LinkApiPasswordRestore returns the API password restore endpoint URL.
func (a authImplementation) LinkApiPasswordRestore() string {
	return links.ApiPasswordRestore(a.endpoint)
}

// LinkApiPasswordReset returns the API password reset endpoint URL.
func (a authImplementation) LinkApiPasswordReset() string {
	return links.ApiPasswordReset(a.endpoint)
}

// LinkLogin returns the login page URL.
func (a authImplementation) LinkLogin() string {
	return links.Login(a.endpoint)
}

// LinkLoginCodeVerify returns the login code verify page URL.
func (a authImplementation) LinkLoginCodeVerify() string {
	return links.LoginCodeVerify(a.endpoint)
}

// LinkLogout returns the logout page URL.
func (a authImplementation) LinkLogout() string {
	return links.Logout(a.endpoint)
}

// LinkPasswordRestore returns the password restore page URL.
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

// WebAuthOrRedirectMiddleware returns a middleware that checks if the user is authenticated and redirects to the login page if not.
func (a *authImplementation) WebAuthOrRedirectMiddleware(next http.Handler) http.Handler {
	return middlewares.WebAuthOrRedirectMiddleware(next, a)
}

// ApiAuthOrErrorMiddleware returns a middleware that checks if the user is authenticated and returns an error if not.
func (a *authImplementation) ApiAuthOrErrorMiddleware(next http.Handler) http.Handler {
	return middlewares.ApiAuthOrErrorMiddleware(next, a)
}

// WebAppendUserIdIfExistsMiddleware returns a middleware that appends the user ID to the context if it exists.
func (a *authImplementation) WebAppendUserIdIfExistsMiddleware(next http.Handler) http.Handler {
	return middlewares.WebAppendUserIdIfExistsMiddleware(next, a)
}

// ======================================================================
// Impersonation
// ======================================================================

// IsImpersonationEnabled returns whether impersonation is enabled.
func (a authImplementation) IsImpersonationEnabled() bool {
	return a.enableImpersonation
}

// GetFuncCanImpersonate returns the function that checks if a user can impersonate another user.
func (a authImplementation) GetFuncCanImpersonate() func(ctx context.Context, adminUserID string, targetUserID string) (bool, error) {
	return a.funcCanImpersonate
}

// SetFuncCanImpersonate sets the function that checks if a user can impersonate another user.
func (a *authImplementation) SetFuncCanImpersonate(fn func(ctx context.Context, adminUserID string, targetUserID string) (bool, error)) {
	a.funcCanImpersonate = fn
}

// GetFuncImpersonationStart returns the function that starts an impersonation session.
func (a authImplementation) GetFuncImpersonationStart() func(ctx context.Context, adminUserID string, targetUserID string) error {
	return a.funcImpersonationStart
}

// SetFuncImpersonationStart sets the function that starts an impersonation session.
func (a *authImplementation) SetFuncImpersonationStart(fn func(ctx context.Context, adminUserID string, targetUserID string) error) {
	a.funcImpersonationStart = fn
}

// GetFuncImpersonationStop returns the function that stops an impersonation session.
func (a authImplementation) GetFuncImpersonationStop() func(ctx context.Context, adminUserID string, targetUserID string) error {
	return a.funcImpersonationStop
}

// SetFuncImpersonationStop sets the function that stops an impersonation session.
func (a *authImplementation) SetFuncImpersonationStop(fn func(ctx context.Context, adminUserID string, targetUserID string) error) {
	a.funcImpersonationStop = fn
}

// LinkApiImpersonateStart returns the API impersonate start endpoint URL.
func (a authImplementation) LinkApiImpersonateStart() string {
	return links.ApiImpersonateStart(a.endpoint)
}

// LinkApiImpersonateStop returns the API impersonate stop endpoint URL.
func (a authImplementation) LinkApiImpersonateStop() string {
	return links.ApiImpersonateStop(a.endpoint)
}

// ======================================================================
// AuthKnight
// ======================================================================

// IsAuthKnight returns whether AuthKnight is enabled.
func (a authImplementation) IsAuthKnight() bool {
	return a.authKnight
}

// GetAuthKnightUserFindByEmail returns the function that finds a user by email.
func (a authImplementation) GetAuthKnightUserFindByEmail() func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
	return a.authKnightFuncUserFindByEmail
}

// SetAuthKnightUserFindByEmail sets the function that finds a user by email.
func (a *authImplementation) SetAuthKnightUserFindByEmail(fn func(ctx context.Context, email string, options types.UserAuthOptions) (string, error)) {
	a.authKnightFuncUserFindByEmail = fn
}

// GetAuthKnightUserRegister returns the function that registers a user.
func (a authImplementation) GetAuthKnightUserRegister() func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (string, error) {
	return a.authKnightFuncUserRegister
}

// SetAuthKnightUserRegister sets the function that registers a user.
func (a *authImplementation) SetAuthKnightUserRegister(fn func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (string, error)) {
	a.authKnightFuncUserRegister = fn
}

// GetAuthKnightRedirectURL returns the function that generates the redirect URL after authentication.
func (a authImplementation) GetAuthKnightRedirectURL() func(ctx context.Context, userID string) string {
	return a.authKnightFuncRedirectURL
}

// SetAuthKnightRedirectURL sets the function that generates the redirect URL after authentication.
func (a *authImplementation) SetAuthKnightRedirectURL(fn func(ctx context.Context, userID string) string) {
	a.authKnightFuncRedirectURL = fn
}

// GetAuthKnightHTTPTimeout returns the HTTP timeout for AuthKnight requests.
func (a authImplementation) GetAuthKnightHTTPTimeout() time.Duration {
	if a.authKnightHTTPTimeout > 0 {
		return a.authKnightHTTPTimeout
	}
	return 10 * time.Second
}

// SetAuthKnightHTTPTimeout sets the HTTP timeout for AuthKnight requests.
func (a *authImplementation) SetAuthKnightHTTPTimeout(d time.Duration) {
	a.authKnightHTTPTimeout = d
}

// LinkAuthKnightCallback returns the callback URL on this server
// that AuthKnight will redirect to after successful authentication.
func (a authImplementation) LinkAuthKnightCallback() string {
	return links.ApiAuthKnightCallback(a.endpoint)
}

// LinkAuthKnightLogin returns the AuthKnight-hosted login URL
// with back_url and next_url parameters set.
// Both parameters are URL-encoded.
func (a authImplementation) LinkAuthKnightLogin(backURL, nextURL string) string {
	return AuthKnightBaseURL + "/app/login?back_url=" + url.QueryEscape(backURL) + "&next_url=" + url.QueryEscape(nextURL)
}

// LinkAuthKnightRedirect builds the full AuthKnight login URL from the request.
// back_url is the cancel URL (scheme+host+endpoint), next_url is the callback URL
// (scheme+host+callback) where AuthKnight sends the once token.
func (a authImplementation) LinkAuthKnightRedirect(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	baseURL := scheme + "://" + r.Host
	cancelURL := baseURL + a.endpoint
	callbackURL := baseURL + a.LinkAuthKnightCallback()
	return a.LinkAuthKnightLogin(cancelURL, callbackURL)
}
