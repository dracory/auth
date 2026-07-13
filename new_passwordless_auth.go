package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/dracory/auth/internal/emails"
	"github.com/dracory/auth/internal/helpers"
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/csrf"
	"github.com/dracory/req"
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
	auth.cookieConfig = defaultCookieConfig()
	if config.CookieConfig != nil {
		if config.CookieConfig.Name != "" {
			auth.cookieConfig.Name = config.CookieConfig.Name
		}
		if config.CookieConfig.HttpOnly != "" {
			auth.cookieConfig.HttpOnly = config.CookieConfig.HttpOnly
		}
		if config.CookieConfig.Secure != "" {
			auth.cookieConfig.Secure = config.CookieConfig.Secure
		}
		if config.CookieConfig.SameSite != 0 {
			auth.cookieConfig.SameSite = config.CookieConfig.SameSite
		}
		if config.CookieConfig.MaxAge > 0 {
			auth.cookieConfig.MaxAge = config.CookieConfig.MaxAge
		}
		if config.CookieConfig.Domain != "" {
			auth.cookieConfig.Domain = config.CookieConfig.Domain
		}
		if config.CookieConfig.Path != "" {
			auth.cookieConfig.Path = config.CookieConfig.Path
		}
	}
	if auth.cookieConfig.Name != "" {
		auth.cookieName = auth.cookieConfig.Name
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
	auth.observabilityHooks = config.ObservabilityHooks

	auth.enableImpersonation = config.EnableImpersonation
	auth.funcCanImpersonate = config.FuncCanImpersonate
	auth.funcImpersonationStart = config.FuncImpersonationStart
	auth.funcImpersonationStop = config.FuncImpersonationStop

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

	if config.EnableImpersonation && config.FuncCanImpersonate == nil {
		return errors.New("auth: FuncCanImpersonate function is required when EnableImpersonation is true")
	}

	return nil
}
