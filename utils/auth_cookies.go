package utils

import (
	"log/slog"
	"net/http"

	"github.com/dracory/auth/types"
)

func AuthCookieSet(w http.ResponseWriter, r *http.Request, token string, opts ...types.CookieOption) {
	c := defaultCookieConfig()
	for _, opt := range opts {
		opt(&c)
	}
	setCookieWithConfig(w, r, token, c)
}

func AuthCookieRemove(w http.ResponseWriter, r *http.Request, opts ...types.CookieOption) {
	c := defaultCookieConfig()
	for _, opt := range opts {
		opt(&c)
	}
	removeCookieWithConfig(w, r, c)
}

func AuthCookieGet(r *http.Request) string {
	return AuthCookieGetWithName(r, types.CookieName)
}

func AuthCookieGetWithName(r *http.Request, cookieName string) string {
	if cookieName == "" {
		cookieName = types.CookieName
	}
	cookie, err := r.Cookie(cookieName)

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
