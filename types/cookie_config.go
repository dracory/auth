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

// CookieOption customizes a CookieConfig by mutating it.
// Options are applied on top of the default config.
type CookieOption func(*CookieConfig)

// WithSecure sets the Secure flag.
func WithSecure(secure bool) CookieOption {
	return func(c *CookieConfig) { c.Secure = secure }
}

// WithHttpOnly sets the HttpOnly flag.
func WithHttpOnly(httpOnly bool) CookieOption {
	return func(c *CookieConfig) { c.HttpOnly = httpOnly }
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
