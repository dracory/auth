package links

import "strings"

// Join combines the base endpoint and a relative URI in a consistent way.
// If the endpoint already has a trailing slash, the URI is appended directly;
// otherwise a slash is inserted between them.
func Join(endpoint, uri string) string {
	if strings.HasSuffix(endpoint, "/") {
		return endpoint + uri
	}
	return endpoint + "/" + uri
}

func ApiLogin(endpoint string) string              { return Join(endpoint, "api/login") }
func ApiLoginCodeVerify(endpoint string) string    { return Join(endpoint, "api/login-code-verify") }
func ApiLogout(endpoint string) string             { return Join(endpoint, "api/logout") }
func ApiRegister(endpoint string) string           { return Join(endpoint, "api/register") }
func ApiRegisterCodeVerify(endpoint string) string { return Join(endpoint, "api/register-code-verify") }
func ApiPasswordRestore(endpoint string) string    { return Join(endpoint, "api/restore-password") }
func ApiPasswordReset(endpoint string) string      { return Join(endpoint, "api/reset-password") }

func Login(endpoint string) string              { return Join(endpoint, "login") }
func LoginCodeVerify(endpoint string) string    { return Join(endpoint, "login-code-verify") }
func Logout(endpoint string) string             { return Join(endpoint, "logout") }
func PasswordRestore(endpoint string) string    { return Join(endpoint, "password-restore") }
func PasswordReset(endpoint string) string      { return Join(endpoint, "password-reset") }
func Register(endpoint string) string           { return Join(endpoint, "register") }
func RegisterCodeVerify(endpoint string) string { return Join(endpoint, "register-code-verify") }
