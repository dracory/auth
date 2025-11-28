package auth

import (
	"net/http"
	"time"
)

func AuthCookieRemove(w http.ResponseWriter, r *http.Request) {
	secureCookie := true

	if r.TLS == nil {
		secureCookie = false // the scheme is HTTP
	}

	expiration := time.Now().Add(-365 * 24 * time.Hour)

	cookie := http.Cookie{
		Name:     CookieName,
		Value:    "none",
		Expires:  expiration,
		HttpOnly: true,
		Secure:   secureCookie,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}

	http.SetCookie(w, &cookie)
}
