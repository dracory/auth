package auth

import (
	"net/http"
	"time"
)

func AuthCookieSet(w http.ResponseWriter, r *http.Request, token string) {
	secureCookie := true

	if r.TLS == nil {
		secureCookie = false // the scheme is HTTP
	}

	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{
		Name:     CookieName,
		Value:    token,
		Expires:  expiration,
		HttpOnly: false,
		Secure:   secureCookie,
		Path:     "/",
	}

	http.SetCookie(w, &cookie)
}
