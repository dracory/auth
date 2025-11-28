package auth

import "net/http"

func AuthCookieSet(w http.ResponseWriter, r *http.Request, token string) {
	setCookieWithConfig(w, r, token, defaultCookieConfig())
}
