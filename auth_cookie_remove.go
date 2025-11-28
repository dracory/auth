package auth

import "net/http"

func AuthCookieRemove(w http.ResponseWriter, r *http.Request) {
	removeCookieWithConfig(w, r, defaultCookieConfig())
}
