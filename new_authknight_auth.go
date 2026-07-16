package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/dracory/auth/internal/helpers"
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/csrf"
	"github.com/dracory/req"
)

// NewAuthKnightAuth creates a new AuthKnight-based authentication instance.
// It validates the provided ConfigAuthKnight and returns an error if required
// fields are missing.
func NewAuthKnightAuth(config types.ConfigAuthKnight) (types.AuthAuthKnightInterface, error) {
	if err := validateAuthKnightConfig(config); err != nil {
		return nil, err
	}

	a := &authImplementation{}

	// Shared config
	a.enableRegistration = config.EnableRegistration
	a.endpoint = config.Endpoint
	a.urlRedirectOnSuccess = config.UrlRedirectOnSuccess
	a.useCookies = config.UseCookies
	a.useLocalStorage = config.UseLocalStorage
	a.cookieConfig = defaultCookieConfig()
	if config.CookieConfig != nil {
		if config.CookieConfig.Name != "" {
			a.cookieConfig.Name = config.CookieConfig.Name
		}
		if config.CookieConfig.HttpOnly != "" {
			a.cookieConfig.HttpOnly = config.CookieConfig.HttpOnly
		}
		if config.CookieConfig.Secure != "" {
			a.cookieConfig.Secure = config.CookieConfig.Secure
		}
		if config.CookieConfig.SameSite != 0 {
			a.cookieConfig.SameSite = config.CookieConfig.SameSite
		}
		if config.CookieConfig.MaxAge > 0 {
			a.cookieConfig.MaxAge = config.CookieConfig.MaxAge
		}
		if config.CookieConfig.Domain != "" {
			a.cookieConfig.Domain = config.CookieConfig.Domain
		}
		if config.CookieConfig.Path != "" {
			a.cookieConfig.Path = config.CookieConfig.Path
		}
	}
	if a.cookieConfig.Name != "" {
		a.cookieName = a.cookieConfig.Name
	}

	a.funcLayout = config.FuncLayout
	if a.funcLayout == nil {
		a.funcLayout = helpers.Layout
	}

	// Shared callbacks
	a.funcUserFindByAuthToken = config.FuncUserFindByAuthToken
	a.funcUserLogout = config.FuncUserLogout
	a.funcUserStoreAuthToken = config.FuncUserStoreAuthToken
	a.funcTemporaryKeyGet = config.FuncTemporaryKeyGet
	a.funcTemporaryKeySet = config.FuncTemporaryKeySet

	// Rate limiting
	a.disableRateLimit = config.DisableRateLimit
	a.funcCheckRateLimit = config.FuncCheckRateLimit
	if !a.disableRateLimit && a.funcCheckRateLimit == nil {
		maxAttempts := config.MaxLoginAttempts
		if maxAttempts == 0 {
			maxAttempts = 5
		}
		lockoutDuration := config.LockoutDuration
		if lockoutDuration == 0 {
			lockoutDuration = 15 * time.Minute
		}
		a.rateLimiter = utils.NewInMemoryRateLimiter(maxAttempts, lockoutDuration, lockoutDuration)
	}

	a.logger = config.Logger
	a.observabilityHooks = config.ObservabilityHooks

	// Impersonation
	a.enableImpersonation = config.EnableImpersonation
	a.funcCanImpersonate = config.FuncCanImpersonate
	a.funcImpersonationStart = config.FuncImpersonationStart
	a.funcImpersonationStop = config.FuncImpersonationStop

	// CSRF
	a.enableCSRFProtection = config.EnableCSRFProtection
	a.csrfSecret = config.CSRFSecret
	if a.enableCSRFProtection {
		if a.csrfSecret == "" {
			return nil, errors.New("auth: CSRFSecret is required when EnableCSRFProtection is true")
		}
		a.funcCSRFTokenGenerate = func(r *http.Request) string {
			return csrf.TokenGenerate(a.csrfSecret, &csrf.Options{
				Request:       r,
				BindIP:        true,
				BindUserAgent: true,
				BindPath:      true,
			})
		}
		a.funcCSRFTokenValidate = func(r *http.Request) bool {
			token := req.GetStringTrimmed(r, "csrf_token")
			if token == "" {
				token = r.Header.Get("X-CSRF-Token")
			}
			return csrf.TokenValidate(token, a.csrfSecret, &csrf.Options{
				Request:       r,
				BindIP:        true,
				BindUserAgent: true,
				BindPath:      true,
			})
		}
	}

	// AuthKnight-specific
	a.authKnight = true
	a.authKnightFuncUserFindByEmail = config.FuncUserFindByEmail
	a.authKnightFuncUserRegister = config.FuncUserRegister
	a.authKnightFuncRedirectURL = config.FuncRedirectURL
	a.authKnightHTTPTimeout = config.HTTPTimeout
	if a.authKnightHTTPTimeout <= 0 {
		a.authKnightHTTPTimeout = 10 * time.Second
	}

	return a, nil
}

// validateAuthKnightConfig validates the required fields of ConfigAuthKnight.
func validateAuthKnightConfig(config types.ConfigAuthKnight) error {
	if config.Endpoint == "" {
		return errors.New("Endpoint is required")
	}

	if config.UrlRedirectOnSuccess == "" {
		return errors.New("UrlRedirectOnSuccess is required")
	}

	if config.FuncUserFindByEmail == nil {
		return errors.New("FuncUserFindByEmail is required")
	}

	if config.FuncUserRegister == nil {
		return errors.New("FuncUserRegister is required")
	}

	if config.FuncUserStoreAuthToken == nil {
		return errors.New("FuncUserStoreAuthToken is required")
	}

	if config.FuncTemporaryKeyGet == nil {
		return errors.New("FuncTemporaryKeyGet is required")
	}

	if config.FuncTemporaryKeySet == nil {
		return errors.New("FuncTemporaryKeySet is required")
	}

	if config.FuncUserFindByAuthToken == nil {
		return errors.New("FuncUserFindByAuthToken is required")
	}

	if config.FuncUserLogout == nil {
		return errors.New("FuncUserLogout is required")
	}

	return nil
}
