package auth

import (
	"time"

	authtypes "github.com/dracory/auth/types"
)

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

type AuthenticatedUserID struct{}

type CookieConfig = authtypes.CookieConfig

const (
	keyEndpoint = contextKey("endpoint")

	CookieName = "authtoken"

	// PathApiLogin contains the path to api login endpoint
	PathApiLogin string = "api/login"

	// PathApiLoginCodeVerify contains the path to api login code verification endpoint
	PathApiLoginCodeVerify string = "api/login-code-verify"

	// PathApiLogout contains the path to api logout endpoint
	PathApiLogout string = "api/logout"

	// PathApiRegister contains the path to api register endpoint
	PathApiRegister string = "api/register"

	// PathApiRegisterCodeVerify contains the path to api register code verification endpoint
	PathApiRegisterCodeVerify string = "api/register-code-verify"

	// PathApiRestorePassword contains the path to api restore password endpoint
	PathApiRestorePassword string = "api/restore-password"

	// PathApiResetPassword contains the path to api reset password endpoint
	PathApiResetPassword string = "api/reset-password"

	// PathLogin contains the path to login page
	PathLogin string = "login"

	// PathLoginCodeVerify contains the path to login code verification page
	PathLoginCodeVerify string = "login-code-verify"

	// PathLogout contains the path to logout page
	PathLogout string = "logout"

	// PathRegister contains the path to logout page
	PathRegister string = "register"

	// PathRegisterCodeVerify contains the path to registration code verification page
	PathRegisterCodeVerify string = "register-code-verify"

	// PathRestore contains the path to password restore page
	PathPasswordRestore string = "password-restore"

	// PathReset contains the path to password reset page
	PathPasswordReset string = "password-reset"

	// LoginCodeLength specified the length of the login code
	LoginCodeLength int = 8

	// LoginCodeGamma specifies the characters to be used for building the login code
	LoginCodeGamma string = "BCDFGHJKLMNPQRSTVXYZ"

	DefaultVerificationCodeExpiration = 1 * time.Hour
	DefaultPasswordResetExpiration    = 1 * time.Hour
	DefaultAuthTokenExpiration        = 2 * time.Hour

	DefaultMaxLoginAttempts = 5
	DefaultLockoutDuration  = 15 * time.Minute
)
