package types

type AuthenticatedUserID struct{}

type ImpersonatorUserID struct{}

type IsImpersonating struct{}

// CookieName is the default cookie name used by auth token helpers.
// It is a variable for backward compatibility and global override scenarios,
// but it should be set at init-time before any auth instances are used.
// For per-instance customization, use CookieConfig.Name or SetCookieName.
var CookieName = "authtoken"

// ImpersonationKeyPrefix is the prefix used for temporary keys that store
// the original admin auth token during an impersonation session.
const ImpersonationKeyPrefix = "imp:"

// Path constants for API endpoints
const (
	PathApiLogin              string = "api/login"
	PathApiLoginCodeVerify    string = "api/login-code-verify"
	PathApiLogout             string = "api/logout"
	PathApiRegister           string = "api/register"
	PathApiRegisterCodeVerify string = "api/register-code-verify"
	PathApiRestorePassword    string = "api/restore-password"
	PathApiResetPassword      string = "api/reset-password"
	PathApiImpersonateStart   string = "api/impersonate/start"
	PathApiImpersonateStop    string = "api/impersonate/stop"
	PathApiAuthKnightCallback string = "api/authknight/callback"
)

// Path constants for UI pages
const (
	PathLogin              string = "login"
	PathLoginCodeVerify    string = "login-code-verify"
	PathLogout             string = "logout"
	PathRegister           string = "register"
	PathRegisterCodeVerify string = "register-code-verify"
	PathPasswordRestore    string = "password-restore"
	PathPasswordReset      string = "password-reset"
)

// AuthKnightBaseURL is the base URL of the AuthKnight service.
const AuthKnightBaseURL string = "https://authknight.com"

// Endpoint constants for rate limiting
const (
	EndpointLogin              string = "login"
	EndpointLoginCodeVerify    string = "login_code_verify"
	EndpointRegister           string = "register"
	EndpointRegisterCodeVerify string = "register_code_verify"
	EndpointPasswordReset      string = "password_reset"
	EndpointPasswordRestore    string = "password_restore"
	EndpointImpersonateStart   string = "impersonate_start"
	EndpointImpersonateStop    string = "impersonate_stop"
	EndpointAuthKnightCallback string = "authknight_callback"
)
