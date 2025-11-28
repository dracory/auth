package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/dracory/auth/utils"
)

func NewPasswordlessAuth(config ConfigPasswordless) (*Auth, error) {
	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	auth := &Auth{}

	if config.Endpoint == "" {
		return nil, errors.New("auth: endpoint is required")
	}

	if config.UrlRedirectOnSuccess == "" {
		return nil, errors.New("auth: url to redirect to on success is required")
	}

	if config.FuncTemporaryKeyGet == nil {
		return nil, errors.New("auth: FuncTemporaryKeyGet function is required")
	}

	if config.FuncTemporaryKeySet == nil {
		return nil, errors.New("auth: FuncTemporaryKeySet function is required")
	}

	if config.FuncUserFindByAuthToken == nil {
		return nil, errors.New("auth: FuncUserFindByAuthToken function is required")
	}

	if config.FuncUserFindByEmail == nil {
		return nil, errors.New("auth: FuncUserFindByEmail function is required")
	}

	if config.FuncUserLogout == nil {
		return nil, errors.New("auth: FuncUserLogout function is required")
	}

	if config.EnableRegistration && config.FuncUserRegister == nil {
		return nil, errors.New("auth: FuncUserRegister function is required")
	}

	if config.FuncUserStoreAuthToken == nil {
		return nil, errors.New("auth: FuncUserStoreToken function is required")
	}

	if config.FuncEmailSend == nil {
		return nil, errors.New("auth: FuncEmailSend function is required")
	}

	if config.UseCookies && config.UseLocalStorage {
		return nil, errors.New("auth: UseCookies and UseLocalStorage cannot be both true")
	}

	if !config.UseCookies && !config.UseLocalStorage {
		return nil, errors.New("auth: UseCookies and UseLocalStorage cannot be both false")
	}

	if config.FuncLayout == nil {
		config.FuncLayout = auth.layout
	}

	auth.enableRegistration = config.EnableRegistration
	auth.endpoint = config.Endpoint
	auth.passwordless = true
	auth.urlRedirectOnSuccess = config.UrlRedirectOnSuccess
	auth.useCookies = config.UseCookies
	auth.useLocalStorage = config.UseLocalStorage
	auth.funcLayout = config.FuncLayout
	auth.funcTemporaryKeyGet = config.FuncTemporaryKeyGet
	auth.funcTemporaryKeySet = config.FuncTemporaryKeySet
	auth.funcUserLogout = config.FuncUserLogout
	auth.funcUserFindByAuthToken = config.FuncUserFindByAuthToken
	auth.funcUserStoreAuthToken = config.FuncUserStoreAuthToken
	auth.passwordlessFuncEmailTemplateLoginCode = config.FuncEmailTemplateLoginCode
	// auth.passwordlessFuncEmailTemplateRegisterCode = config.FuncEmailTemplateRegisterCode
	auth.passwordlessFuncEmailSend = config.FuncEmailSend
	auth.passwordlessFuncUserFindByEmail = config.FuncUserFindByEmail
	auth.passwordlessFuncUserRegister = config.FuncUserRegister

	// If no user defined email template is set, use default
	if auth.passwordlessFuncEmailTemplateLoginCode == nil {
		auth.passwordlessFuncEmailTemplateLoginCode = func(ctx context.Context, email string, code string, options UserAuthOptions) string {
			return emailLoginCodeTemplate(email, code, options)
		}
	}

	// If no user defined email template is set, use default
	if auth.passwordlessFuncEmailTemplateRegisterCode == nil {
		auth.passwordlessFuncEmailTemplateRegisterCode = func(ctx context.Context, email string, code string, options UserAuthOptions) string {
			return emailRegisterCodeTemplate(email, code, options)
		}
	}

	// Initialize rate limiting
	auth.disableRateLimit = config.DisableRateLimit
	auth.funcCheckRateLimit = config.FuncCheckRateLimit

	// If rate limiting is not disabled and no custom function provided, use default in-memory rate limiter
	if !auth.disableRateLimit && auth.funcCheckRateLimit == nil {
		// Use config values or defaults
		maxAttempts := config.MaxLoginAttempts
		if maxAttempts == 0 {
			maxAttempts = 5 // Default: 5 attempts
		}

		lockoutDuration := config.LockoutDuration
		if lockoutDuration == 0 {
			lockoutDuration = 15 * time.Minute // Default: 15 minutes
		}

		auth.rateLimiter = utils.NewInMemoryRateLimiter(maxAttempts, lockoutDuration, lockoutDuration)
	}

	auth.logger = logger

	return auth, nil
}
