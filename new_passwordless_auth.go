package auth

import (
	"context"
	"errors"
	"time"

	"github.com/dracory/auth/internal/emails"
	"github.com/dracory/auth/internal/helpers"
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
)

func NewPasswordlessAuth(config types.ConfigPasswordless) (types.AuthPasswordlessInterface, error) {
	if err := validatePasswordlessConfig(config); err != nil {
		return nil, err
	}

	auth := &authImplementation{}
	auth.enableRegistration = config.EnableRegistration
	auth.endpoint = config.Endpoint
	auth.passwordless = true
	auth.urlRedirectOnSuccess = config.UrlRedirectOnSuccess
	auth.useCookies = config.UseCookies
	auth.useLocalStorage = config.UseLocalStorage
	if config.CookieConfig != nil {
		auth.cookieConfig = *config.CookieConfig
	} else {
		auth.cookieConfig = defaultCookieConfig()
	}
	auth.funcLayout = config.FuncLayout
	if auth.funcLayout == nil {
		auth.funcLayout = helpers.Layout
	}
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
		auth.passwordlessFuncEmailTemplateLoginCode = func(ctx context.Context, email string, code string, options types.UserAuthOptions) string {
			return emails.EmailLoginCodeTemplate(email, code)
		}
	}

	// If no user defined email template is set, use default
	if auth.passwordlessFuncEmailTemplateRegisterCode == nil {
		auth.passwordlessFuncEmailTemplateRegisterCode = func(ctx context.Context, email string, code string, options types.UserAuthOptions) string {
			return emails.EmailRegisterCodeTemplate(email, code)
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

	auth.logger = config.Logger

	return auth, nil
}

// validatePasswordlessConfig performs validation of the ConfigPasswordless
// values and returns a descriptive error if any required field is missing
// or invalid.
func validatePasswordlessConfig(config types.ConfigPasswordless) error {
	if config.Endpoint == "" {
		return errors.New("auth: endpoint is required")
	}

	if config.UrlRedirectOnSuccess == "" {
		return errors.New("auth: url to redirect to on success is required")
	}

	if config.FuncTemporaryKeyGet == nil {
		return errors.New("auth: FuncTemporaryKeyGet function is required")
	}

	if config.FuncTemporaryKeySet == nil {
		return errors.New("auth: FuncTemporaryKeySet function is required")
	}

	if config.FuncUserFindByAuthToken == nil {
		return errors.New("auth: FuncUserFindByAuthToken function is required")
	}

	if config.FuncUserFindByEmail == nil {
		return errors.New("auth: FuncUserFindByEmail function is required")
	}

	if config.FuncUserLogout == nil {
		return errors.New("auth: FuncUserLogout function is required")
	}

	if config.EnableRegistration && config.FuncUserRegister == nil {
		return errors.New("auth: FuncUserRegister function is required")
	}

	if config.FuncUserStoreAuthToken == nil {
		return errors.New("auth: FuncUserStoreToken function is required")
	}

	if config.FuncEmailSend == nil {
		return errors.New("auth: FuncEmailSend function is required")
	}

	if config.UseCookies && config.UseLocalStorage {
		return errors.New("auth: UseCookies and UseLocalStorage cannot be both true")
	}

	if !config.UseCookies && !config.UseLocalStorage {
		return errors.New("auth: UseCookies and UseLocalStorage cannot be both false")
	}

	return nil
}
