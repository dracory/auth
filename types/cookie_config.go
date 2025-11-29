package types

import "net/http"

type CookieConfig struct {
	HttpOnly bool
	Secure   bool
	SameSite http.SameSite
	MaxAge   int
	Domain   string
	Path     string
}
