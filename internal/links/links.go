package links

import (
	"strings"

	"github.com/dracory/auth/types"
)

// Join combines the base endpoint and a relative URI in a consistent way.
// If the endpoint already has a trailing slash, the URI is appended directly;
// otherwise a slash is inserted between them.
func Join(endpoint, uri string) string {
	if strings.HasSuffix(endpoint, "/") {
		return endpoint + uri
	}
	return endpoint + "/" + uri
}

func ApiLogin(endpoint string) string           { return Join(endpoint, types.PathApiLogin) }
func ApiLoginCodeVerify(endpoint string) string { return Join(endpoint, types.PathApiLoginCodeVerify) }
func ApiLogout(endpoint string) string          { return Join(endpoint, types.PathApiLogout) }
func ApiRegister(endpoint string) string        { return Join(endpoint, types.PathApiRegister) }
func ApiRegisterCodeVerify(endpoint string) string {
	return Join(endpoint, types.PathApiRegisterCodeVerify)
}
func ApiPasswordRestore(endpoint string) string { return Join(endpoint, types.PathApiRestorePassword) }
func ApiPasswordReset(endpoint string) string   { return Join(endpoint, types.PathApiResetPassword) }
func ApiImpersonateStart(endpoint string) string {
	return Join(endpoint, types.PathApiImpersonateStart)
}
func ApiImpersonateStop(endpoint string) string { return Join(endpoint, types.PathApiImpersonateStop) }
func ApiAuthKnightCallback(endpoint string) string {
	return Join(endpoint, types.PathApiAuthKnightCallback)
}

func Login(endpoint string) string              { return Join(endpoint, types.PathLogin) }
func LoginCodeVerify(endpoint string) string    { return Join(endpoint, types.PathLoginCodeVerify) }
func Logout(endpoint string) string             { return Join(endpoint, types.PathLogout) }
func PasswordRestore(endpoint string) string    { return Join(endpoint, types.PathPasswordRestore) }
func PasswordReset(endpoint string) string      { return Join(endpoint, types.PathPasswordReset) }
func Register(endpoint string) string           { return Join(endpoint, types.PathRegister) }
func RegisterCodeVerify(endpoint string) string { return Join(endpoint, types.PathRegisterCodeVerify) }
