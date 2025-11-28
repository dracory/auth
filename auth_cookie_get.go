package auth

import (
	"log/slog"
	"net/http"
)

func AuthCookieGet(r *http.Request) string {
	cookie, err := r.Cookie(CookieName)

	if err != nil {

		if err != http.ErrNoCookie {
			slog.Error("auth cookie retrieval failed",
				"error", err,
			)
		}

		return ""
	}

	return cookie.Value
}
