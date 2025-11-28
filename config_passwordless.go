package auth

import (
	"context"
	"log/slog"
	"time"
)

type ConfigPasswordless struct {

	// ===== START: shared by all implementations
	EnableRegistration      bool
	Endpoint                string
	FuncLayout              func(content string) string
	FuncTemporaryKeyGet     func(key string) (value string, err error)
	FuncTemporaryKeySet     func(key string, value string, expiresSeconds int) (err error)
	FuncUserFindByAuthToken func(ctx context.Context, sessionID string, options UserAuthOptions) (userID string, err error)
	FuncUserLogout          func(ctx context.Context, userID string, options UserAuthOptions) (err error)
	FuncUserStoreAuthToken  func(ctx context.Context, sessionID string, userID string, options UserAuthOptions) error
	UrlRedirectOnSuccess    string
	UseCookies              bool
	UseLocalStorage         bool
	// Rate limiting options
	DisableRateLimit   bool                                                                                 // Set to true to disable rate limiting (not recommended for production)
	FuncCheckRateLimit func(ip string, endpoint string) (allowed bool, retryAfter time.Duration, err error) // Optional: override default rate limiter
	MaxLoginAttempts   int                                                                                  // Maximum attempts before lockout (default: 5)
	LockoutDuration    time.Duration                                                                        // Duration to lock after max attempts (default: 15 minutes)
	// CSRF Protection
	EnableCSRFProtection bool
	CSRFSecret           string
	Logger               *slog.Logger

	// ===== END: shared by all implementations

	// ===== START: passwordless options
	FuncUserFindByEmail           func(ctx context.Context, email string, options UserAuthOptions) (userID string, err error)
	FuncEmailTemplateLoginCode    func(ctx context.Context, email string, logingLink string, options UserAuthOptions) string   // optional
	FuncEmailTemplateRegisterCode func(ctx context.Context, email string, registerLink string, options UserAuthOptions) string // optional
	FuncEmailSend                 func(ctx context.Context, email string, emailSubject string, emailBody string) (err error)
	FuncUserRegister              func(ctx context.Context, email string, firstName string, lastName string, options UserAuthOptions) (err error)
	// ===== END: passwordless options
}
