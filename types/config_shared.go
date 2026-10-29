package types

import (
	"context"
	"log/slog"
	"time"
)

// ConfigShared contains fields common to all auth modes.
type ConfigShared struct {
	EnableRegistration bool
	Endpoint           string
	FuncLayout         func(content string) string
	// ThemeCSS is a stylesheet URL injected after Bootstrap on every
	// built-in page — point it at a Bootlegance theme (e.g.
	// https://cdn.jsdelivr.net/gh/lesichkovm/bootlegance@main/themes/dracula/theme.css)
	// or your own hosted stylesheet to rebrand the stock pages.
	ThemeCSS string
	// CustomCSS is raw CSS injected as an inline <style> block after
	// ThemeCSS — for small overrides without hosting a file.
	CustomCSS               string
	FuncTemporaryKeyGet     func(key string) (value string, err error)
	FuncTemporaryKeySet     func(key string, value string, expiresSeconds int) (err error)
	FuncUserStoreAuthToken  func(ctx context.Context, sessionID string, userID string, options UserAuthOptions) error
	FuncUserFindByAuthToken func(ctx context.Context, sessionID string, options UserAuthOptions) (userID string, err error)
	FuncUserLogout          func(ctx context.Context, userID string, options UserAuthOptions) (err error)
	UrlRedirectOnSuccess    string
	UseCookies              bool
	UseLocalStorage         bool
	CookieConfig            *CookieConfig
	Logger                  *slog.Logger
	ObservabilityHooks      ObservabilityHooks
}

// ConfigRateLimiting contains rate limiting options.
type ConfigRateLimiting struct {
	DisableRateLimit   bool
	FuncCheckRateLimit func(ip string, endpoint string) (allowed bool, retryAfter time.Duration, err error)
	MaxLoginAttempts   int
	LockoutDuration    time.Duration
}

// ConfigCSRF contains CSRF protection options.
type ConfigCSRF struct {
	EnableCSRFProtection bool
	CSRFSecret           string
}

// ConfigImpersonation contains impersonation options.
type ConfigImpersonation struct {
	EnableImpersonation    bool
	FuncCanImpersonate     func(ctx context.Context, adminUserID string, targetUserID string) (bool, error)
	FuncImpersonationStart func(ctx context.Context, adminUserID string, targetUserID string) error
	FuncImpersonationStop  func(ctx context.Context, adminUserID string, targetUserID string) error
}
