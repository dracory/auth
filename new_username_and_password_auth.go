package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	authtypes "github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/csrf"
	"github.com/dracory/req"
)

func NewUsernameAndPasswordAuth(config ConfigUsernameAndPassword) (*Auth, error) {
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

	if config.FuncUserFindByUsername == nil {
		return nil, errors.New("auth: FuncUserFindByUsername function is required")
	}

	if config.FuncUserLogin == nil {
		return nil, errors.New("auth: FuncUserLogin function is required")
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

	logger := config.Logger
	if logger == nil {
		logger = slog.Default()
	}

	auth := &Auth{}
	auth.enableRegistration = config.EnableRegistration
	auth.enableVerification = config.EnableVerification
	auth.endpoint = config.Endpoint
	auth.passwordless = false
	auth.urlRedirectOnSuccess = config.UrlRedirectOnSuccess
	auth.useCookies = config.UseCookies
	auth.useLocalStorage = config.UseLocalStorage
	if config.CookieConfig != nil {
		auth.cookieConfig = *config.CookieConfig
	} else {
		auth.cookieConfig = defaultCookieConfig()
	}
	auth.funcEmailSend = config.FuncEmailSend
	auth.funcEmailTemplatePasswordRestore = config.FuncEmailTemplatePasswordRestore
	auth.funcLayout = config.FuncLayout
	auth.funcTemporaryKeyGet = config.FuncTemporaryKeyGet
	auth.funcTemporaryKeySet = config.FuncTemporaryKeySet
	auth.funcUserLogin = config.FuncUserLogin
	auth.funcUserLogout = config.FuncUserLogout
	auth.funcUserPasswordChange = config.FuncUserPasswordChange
	auth.funcUserRegister = config.FuncUserRegister
	auth.funcUserFindByAuthToken = config.FuncUserFindByAuthToken
	auth.funcUserFindByUsername = config.FuncUserFindByUsername
	auth.funcUserStoreAuthToken = config.FuncUserStoreAuthToken
	auth.passwordStrength = config.PasswordStrength
	if auth.passwordStrength == nil {
		auth.passwordStrength = &authtypes.PasswordStrengthConfig{
			MinLength:         8,
			RequireUppercase:  true,
			RequireLowercase:  true,
			RequireDigit:      true,
			RequireSpecial:    true,
			ForbidCommonWords: true,
		}
	}

	auth.logger = logger

	// If no user defined layout is set, use default
	if auth.funcLayout == nil {
		auth.funcLayout = auth.layout
	}

	// If no user defined email template is set, use default
	if auth.funcEmailTemplatePasswordRestore == nil {
		auth.funcEmailTemplatePasswordRestore = func(ctx context.Context, userID string, passwordRestoreLink string, options UserAuthOptions) string {
			// userID here is effectively the name/email for the template
			return emailTemplatePasswordChange(userID, passwordRestoreLink, options)
		}
	}

	// If no user defined email template is set, use default
	if auth.funcEmailTemplateRegisterCode == nil {
		auth.funcEmailTemplateRegisterCode = func(ctx context.Context, email string, code string, options UserAuthOptions) string {
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

	// Initialize CSRF protection
	auth.enableCSRFProtection = config.EnableCSRFProtection
	auth.csrfSecret = config.CSRFSecret
	if auth.enableCSRFProtection {
		if auth.csrfSecret == "" {
			return nil, errors.New("auth: CSRFSecret is required when EnableCSRFProtection is true")
		}
		auth.funcCSRFTokenGenerate = func(r *http.Request) string {
			return csrf.TokenGenerate(auth.csrfSecret, &csrf.Options{
				Request:       r,
				BindIP:        true,
				BindUserAgent: true,
				BindPath:      true,
			})
		}
		auth.funcCSRFTokenValidate = func(r *http.Request) bool {
			token := req.GetStringTrimmed(r, "csrf_token")
			if token == "" {
				token = r.Header.Get("X-CSRF-Token")
			}
			return csrf.TokenValidate(token, auth.csrfSecret, &csrf.Options{
				Request:       r,
				BindIP:        true,
				BindUserAgent: true,
				BindPath:      true,
			})
		}
	}

	return auth, nil
}
