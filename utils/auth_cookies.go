package utils

import (
	"log/slog"
	"net/http"

	"github.com/dracory/auth/types"
)

func AuthCookieSet(w http.ResponseWriter, r *http.Request, token string) {
	setCookieWithConfig(w, r, token, defaultCookieConfig())
}

func AuthCookieRemove(w http.ResponseWriter, r *http.Request) {
	removeCookieWithConfig(w, r, defaultCookieConfig())
}

func AuthCookieGet(r *http.Request) string {
	cookie, err := r.Cookie(types.CookieName)

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
