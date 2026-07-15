package types

import (
	"context"
	"log/slog"
	"time"
)

// ConfigShared contains fields common to all auth modes.
type ConfigShared struct {
	EnableRegistration      bool
	Endpoint                string
	FuncLayout              func(content string) string
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
