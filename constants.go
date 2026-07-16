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

var CookieName = "authtoken"

const (
	keyEndpoint = contextKey("endpoint")

	// Path constants re-exported from types for backward compatibility
	PathApiLogin              = authtypes.PathApiLogin
	PathApiLoginCodeVerify    = authtypes.PathApiLoginCodeVerify
	PathApiLogout             = authtypes.PathApiLogout
	PathApiRegister           = authtypes.PathApiRegister
	PathApiRegisterCodeVerify = authtypes.PathApiRegisterCodeVerify
	PathApiRestorePassword    = authtypes.PathApiRestorePassword
	PathApiResetPassword      = authtypes.PathApiResetPassword
	PathApiImpersonateStart   = authtypes.PathApiImpersonateStart
	PathApiImpersonateStop    = authtypes.PathApiImpersonateStop
	PathApiAuthKnightCallback = authtypes.PathApiAuthKnightCallback

	AuthKnightBaseURL = authtypes.AuthKnightBaseURL

	PathLogin              = authtypes.PathLogin
	PathLoginCodeVerify    = authtypes.PathLoginCodeVerify
	PathLogout             = authtypes.PathLogout
	PathRegister           = authtypes.PathRegister
	PathRegisterCodeVerify = authtypes.PathRegisterCodeVerify
	PathPasswordRestore    = authtypes.PathPasswordRestore
	PathPasswordReset      = authtypes.PathPasswordReset

	EndpointLogin              = authtypes.EndpointLogin
	EndpointLoginCodeVerify    = authtypes.EndpointLoginCodeVerify
	EndpointRegister           = authtypes.EndpointRegister
	EndpointRegisterCodeVerify = authtypes.EndpointRegisterCodeVerify
	EndpointPasswordReset      = authtypes.EndpointPasswordReset
	EndpointPasswordRestore    = authtypes.EndpointPasswordRestore
	EndpointImpersonateStart   = authtypes.EndpointImpersonateStart
	EndpointImpersonateStop    = authtypes.EndpointImpersonateStop
	EndpointAuthKnightCallback = authtypes.EndpointAuthKnightCallback

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
