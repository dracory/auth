package utils

import (
	"net/http"
	"time"

	"github.com/dracory/auth/types"
)

func defaultCookieConfig() types.CookieConfig {
	return types.CookieConfig{
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   2 * 60 * 60,
		Path:     "/",
	}
}

func setCookieWithConfig(w http.ResponseWriter, r *http.Request, token string, cfg types.CookieConfig) {
	sameSite := cfg.SameSite
	if sameSite == 0 {
		sameSite = http.SameSiteLaxMode
	}

	path := cfg.Path
	if path == "" {
		path = "/"
	}

	maxAge := cfg.MaxAge
	if maxAge <= 0 {
		maxAge = 2 * 60 * 60
	}

	secure := false
	if cfg.Secure && r.TLS != nil {
		secure = true
	}

	expires := time.Now().Add(time.Duration(maxAge) * time.Second)

	cookie := http.Cookie{
		Name:     types.CookieName,
		Value:    token,
		HttpOnly: cfg.HttpOnly,
		Secure:   secure,
		SameSite: sameSite,
		Path:     path,
		Domain:   cfg.Domain,
		Expires:  expires,
		MaxAge:   maxAge,
	}

	http.SetCookie(w, &cookie)
}

func removeCookieWithConfig(w http.ResponseWriter, r *http.Request, cfg types.CookieConfig) {
	sameSite := cfg.SameSite
	if sameSite == 0 {
		sameSite = http.SameSiteLaxMode
	}

	path := cfg.Path
	if path == "" {
		path = "/"
	}

	secure := false
	if cfg.Secure && r.TLS != nil {
		secure = true
	}

	cookie := http.Cookie{
		Name:     types.CookieName,
		Value:    "none",
		HttpOnly: cfg.HttpOnly,
		Secure:   secure,
		SameSite: sameSite,
		Path:     path,
		Domain:   cfg.Domain,
		Expires:  time.Now().Add(-time.Hour),
		MaxAge:   -1,
	}

	http.SetCookie(w, &cookie)
}
