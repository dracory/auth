package utils

import "net/http"

// RememberSessionMarker is the session value marking remember-me sessions
// so a remember key cannot act as a normal auth credential and vice versa.
const RememberSessionMarker = "remember"

// DefaultRememberCookieName is the cookie holding the remember token.
const DefaultRememberCookieName = "remember_token"

// RememberMeCheckbox renders the opt-in checkbox fragment for login pages;
// returns "" when remember-me is disabled.
func RememberMeCheckbox(enabled bool) string {
	if !enabled {
		return ""
	}
	return `<label><input type="checkbox" name="remember" value="true" /> Remember me</label>`
}

// RememberCookieSet writes the remember cookie: HttpOnly, MaxAge=days,
// Secure in production only.
func RememberCookieSet(w http.ResponseWriter, name string, key string, days int, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    key,
		Path:     "/",
		MaxAge:   days * 24 * 3600,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// RememberCookieRemove expires the remember cookie (logout).
func RememberCookieRemove(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}
