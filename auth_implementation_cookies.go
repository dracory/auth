package auth

import (
	"net/http"
	"time"

	"github.com/dracory/auth/types"
)

func (a authImplementation) setAuthCookie(w http.ResponseWriter, r *http.Request, token string) {
	cfg := a.cookieConfig
	if cfg == (CookieConfig{}) {
		cfg = defaultCookieConfig()
	}
	if cfg.Name == "" && a.cookieName != "" {
		cfg.Name = a.cookieName
	}
	setCookieWithConfig(w, r, token, cfg)
}

func (a authImplementation) removeAuthCookie(w http.ResponseWriter, r *http.Request) {
	cfg := a.cookieConfig
	if cfg == (CookieConfig{}) {
		cfg = defaultCookieConfig()
	}
	if cfg.Name == "" && a.cookieName != "" {
		cfg.Name = a.cookieName
	}
	removeCookieWithConfig(w, r, cfg)
}

func defaultCookieConfig() CookieConfig {
	return CookieConfig{
		HttpOnly: types.CookieHttpOnly,
		Secure:   types.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   2 * 60 * 60,
		Path:     "/",
	}
}

func setCookieWithConfig(w http.ResponseWriter, r *http.Request, token string, cfg CookieConfig) {
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

	secure := types.ResolveSecure(cfg.Secure)

	expires := time.Now().Add(time.Duration(maxAge) * time.Second)

	name := cfg.Name
	if name == "" {
		name = CookieName
	}

	cookie := http.Cookie{
		Name:     name,
		Value:    token,
		HttpOnly: types.ResolveHttpOnly(cfg.HttpOnly),
		Secure:   secure,
		SameSite: sameSite,
		Path:     path,
		Domain:   cfg.Domain,
		Expires:  expires,
		MaxAge:   maxAge,
	}

	http.SetCookie(w, &cookie)
}

func removeCookieWithConfig(w http.ResponseWriter, r *http.Request, cfg CookieConfig) {
	sameSite := cfg.SameSite
	if sameSite == 0 {
		sameSite = http.SameSiteLaxMode
	}

	path := cfg.Path
	if path == "" {
		path = "/"
	}

	name := cfg.Name
	if name == "" {
		name = CookieName
	}

	httpOnly := types.ResolveHttpOnly(cfg.HttpOnly)
	baseCookie := http.Cookie{
		Name:     name,
		Value:    "none",
		HttpOnly: httpOnly,
		SameSite: sameSite,
		Path:     path,
		Domain:   cfg.Domain,
		Expires:  time.Now().Add(-time.Hour),
		MaxAge:   -1,
	}

	// Emit both Secure and Insecure removal cookies. A removal cookie only
	// overwrites a cookie whose Secure attribute matches the original, so this
	// ensures the auth token is cleared regardless of how it was set.
	for _, secure := range []bool{types.ResolveSecure(cfg.Secure), !types.ResolveSecure(cfg.Secure)} {
		cookie := baseCookie
		cookie.Secure = secure
		http.SetCookie(w, &cookie)
	}
}
