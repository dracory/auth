package auth

import (
	"time"
)

// Config defines the available configuration options for authentication
type ConfigUsernameAndPassword struct {
	// ===== START: shared by all implementations
	EnableRegistration      bool
	Endpoint                string
	FuncLayout              func(content string) string
	FuncTemporaryKeyGet     func(key string) (value string, err error)
	FuncTemporaryKeySet     func(key string, value string, expiresSeconds int) (err error)
	FuncUserStoreAuthToken  func(sessionID string, userID string, options UserAuthOptions) error
	FuncUserFindByAuthToken func(sessionID string, options UserAuthOptions) (userID string, err error)
	UrlRedirectOnSuccess    string
	UseCookies              bool
	UseLocalStorage         bool
	// Rate limiting options
	DisableRateLimit   bool                                                                                 // Set to true to disable rate limiting (not recommended for production)
	FuncCheckRateLimit func(ip string, endpoint string) (allowed bool, retryAfter time.Duration, err error) // Optional: override default rate limiter
	MaxLoginAttempts   int                                                                                  // Maximum attempts before lockout (default: 5)
	LockoutDuration    time.Duration                                                                        // Duration to lock after max attempts (default: 15 minutes)
	// ===== END: shared by all implementations

	// ===== START: username(email) and password options
	EnableVerification               bool
	FuncEmailTemplatePasswordRestore func(userID string, passwordRestoreLink string, options UserAuthOptions) string // optional
	FuncEmailTemplateRegisterCode    func(userID string, passwordRestoreLink string, options UserAuthOptions) string // optional
	FuncEmailSend                    func(userID string, emailSubject string, emailBody string) (err error)
	FuncUserFindByUsername           func(username string, firstName string, lastName string, options UserAuthOptions) (userID string, err error)
	FuncUserLogin                    func(username string, password string, options UserAuthOptions) (userID string, err error)
	FuncUserLogout                   func(userID string, options UserAuthOptions) (err error)
	FuncUserPasswordChange           func(username string, newPassword string, options UserAuthOptions) (err error)
	FuncUserRegister                 func(username string, password string, first_name string, last_name string, options UserAuthOptions) (err error)
	LabelUsername                    string
	// ===== END: username(email) and password options
}
