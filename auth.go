package auth

import (
	"net/http"
	"strings"
	"time"
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
	funcUserFindByAuthToken func(token string, options UserAuthOptions) (userID string, err error)
	funcUserLogout          func(userID string, options UserAuthOptions) (err error)
	funcUserStoreAuthToken  func(token string, userID string, options UserAuthOptions) error
	// ===== END: shared by all implementations

	// ===== START: username(email) and password options
	enableVerification               bool
	funcEmailTemplatePasswordRestore func(userID string, passwordRestoreLink string, options UserAuthOptions) string // optional
	funcEmailTemplateRegisterCode    func(email string, passwordRestoreLink string, options UserAuthOptions) string  // optional
	funcEmailSend                    func(userID string, emailSubject string, emailBody string) (err error)
	funcUserLogin                    func(username string, password string, options UserAuthOptions) (userID string, err error)
	funcUserPasswordChange           func(username string, newPassword string, options UserAuthOptions) (err error)
	funcUserRegister                 func(username string, password string, first_name string, last_name string, options UserAuthOptions) (err error)
	funcUserFindByUsername           func(username string, first_name string, last_name string, options UserAuthOptions) (userID string, err error)
	// ===== END: username(email) and password options

	// ===== START: passwordless options
	passwordless                              bool
	passwordlessFuncUserFindByEmail           func(email string, options UserAuthOptions) (userID string, err error)
	passwordlessFuncEmailTemplateLoginCode    func(email string, passwordRestoreLink string, options UserAuthOptions) string // optional
	passwordlessFuncEmailTemplateRegisterCode func(email string, passwordRestoreLink string, options UserAuthOptions) string // optional
	passwordlessFuncEmailSend                 func(email string, emailSubject string, emailBody string) (err error)
	passwordlessFuncUserRegister              func(email string, firstName string, lastName string, options UserAuthOptions) (err error)
	// ===== END: passwordless options

	// ===== START: rate limiting
	disableRateLimit   bool
	funcCheckRateLimit func(ip string, endpoint string) (allowed bool, retryAfter time.Duration, err error)
	rateLimiter        *InMemoryRateLimiter
	// ===== END: rate limiting

	// ===== START: CSRF Protection
	enableCSRFProtection  bool
	csrfSecret            string
	funcCSRFTokenGenerate func(r *http.Request) string
	funcCSRFTokenValidate func(r *http.Request) bool
	// ===== END: CSRF Protection

	// labelUsername   string
	useCookies      bool
	useLocalStorage bool
}

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
