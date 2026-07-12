package auth

import (
	"net/http"

	"github.com/dracory/auth/utils"
)

// Public backward-compatible shortcuts for the v0.29.0 auth package API.
// The underlying implementations live in the utils package, but these
// top-level functions keep existing callers working without modification.

// AuthCookieSet sets the authentication cookie.
func AuthCookieSet(w http.ResponseWriter, r *http.Request, token string) {
	utils.AuthCookieSet(w, r, token)
}

// AuthCookieGet returns the authentication cookie value, or an empty string if missing.
func AuthCookieGet(r *http.Request) string {
	return utils.AuthCookieGet(r)
}

// AuthCookieRemove removes the authentication cookie.
func AuthCookieRemove(w http.ResponseWriter, r *http.Request) {
	utils.AuthCookieRemove(w, r)
}

// AuthTokenRetrieve retrieves the auth token from the request.
func AuthTokenRetrieve(r *http.Request, useCookies bool) string {
	return utils.AuthTokenRetrieve(r, useCookies)
}
