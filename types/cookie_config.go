package types

import "net/http"

type CookieConfig struct {
	Name     string
	HttpOnly string
	Secure   string
	SameSite http.SameSite
	MaxAge   int
	Domain   string
	Path     string
}

const (
	CookieUnset        = ""
	CookieSecure       = "secure"
	CookieInsecure     = "insecure"
	CookieHttpOnly     = "httponly"
	CookieHttpWritable = "httpwritable"
)

func ResolveSecure(s string) bool {
	return s != CookieInsecure
}

func ResolveHttpOnly(s string) bool {
	return s != CookieHttpWritable
}

// CookieOption customizes a CookieConfig by mutating it.
// Options are applied on top of the default config.
type CookieOption func(*CookieConfig)

// WithSecure sets the Secure flag. Pass true for CookieSecure, false for CookieInsecure.
func WithSecure(secure bool) CookieOption {
	return func(c *CookieConfig) {
		if secure {
			c.Secure = CookieSecure
		} else {
			c.Secure = CookieInsecure
		}
	}
}

// WithHttpOnly sets the HttpOnly flag. Pass true for CookieHttpOnly, false for CookieHttpWritable.
func WithHttpOnly(httpOnly bool) CookieOption {
	return func(c *CookieConfig) {
		if httpOnly {
			c.HttpOnly = CookieHttpOnly
		} else {
			c.HttpOnly = CookieHttpWritable
		}
	}
}

// WithSameSite sets the SameSite attribute.
func WithSameSite(sameSite http.SameSite) CookieOption {
	return func(c *CookieConfig) { c.SameSite = sameSite }
}

// WithMaxAge sets the MaxAge in seconds.
func WithMaxAge(maxAge int) CookieOption {
	return func(c *CookieConfig) { c.MaxAge = maxAge }
}

// WithDomain sets the Domain.
func WithDomain(domain string) CookieOption {
	return func(c *CookieConfig) { c.Domain = domain }
}

// WithPath sets the Path.
func WithPath(path string) CookieOption {
	return func(c *CookieConfig) { c.Path = path }
}

// WithCookieConfig replaces the entire cookie config with the provided one.
// Use this when you already have a complete CookieConfig.
func WithCookieConfig(cfg CookieConfig) CookieOption {
	return func(c *CookieConfig) { *c = cfg }
}

// WithCookieName sets the cookie name.
func WithCookieName(name string) CookieOption {
	return func(c *CookieConfig) { c.Name = name }
}
