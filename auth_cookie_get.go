package auth

import (
	"log"
	"net/http"
)

func AuthCookieGet(r *http.Request) string {
	cookie, err := r.Cookie(CookieName)

	if err != nil {

		if err != http.ErrNoCookie {
			log.Println(err.Error())
		}

		return ""
	}

	return cookie.Value
}
