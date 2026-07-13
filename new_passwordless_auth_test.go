package auth

import (
	"context"
	"testing"

	"github.com/dracory/auth/types"
)

func TestNewPasswordlessAuth_EndpointRequired(t *testing.T) {
	_, err := NewPasswordlessAuth(types.ConfigPasswordless{})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: endpoint is required" {
		t.Fatal("Error SHOULD BE '', but found '", err.Error(), "'")
	}
}

func TestNewPasswordlessAuth_UrlToRedirectOnSuccessIsRequired(t *testing.T) {
	_, err := NewPasswordlessAuth(types.ConfigPasswordless{
		Endpoint: "/auth",
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: url to redirect to on success is required" {
		t.Fatal("Error SHOULD BE '', but found '", err.Error(), "'")
	}
}

func TestNewPasswordlessAuth_FuncTemporaryKeyGetIsRequired(t *testing.T) {
	_, err := NewPasswordlessAuth(types.ConfigPasswordless{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: FuncTemporaryKeyGet function is required" {
		t.Fatal("Error SHOULD BE '', but found '", err.Error(), "'")
	}
}

func TestNewPasswordlessAuth_FuncTemporaryKeySetIsRequired(t *testing.T) {
	_, err := NewPasswordlessAuth(types.ConfigPasswordless{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: FuncTemporaryKeySet function is required" {
		t.Fatal("Error SHOULD BE '', but found '", err.Error(), "'")
	}
}

func TestNewPasswordlessAuth_FuncUserFindByAuthTokenIsRequired(t *testing.T) {
	_, err := NewPasswordlessAuth(types.ConfigPasswordless{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: FuncUserFindByAuthToken function is required" {
		t.Fatal("Error SHOULD BE '', but found '", err.Error(), "'")
	}
}

func TestNewPasswordlessAuth_FuncUserFindByEmailIsRequired(t *testing.T) {
	_, err := NewPasswordlessAuth(types.ConfigPasswordless{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: FuncUserFindByEmail function is required" {
		t.Fatal("Error SHOULD BE '', but found ", "'"+err.Error()+"'")
	}
}

func TestNewPasswordlessAuth_FuncUserLogoutIsRequired(t *testing.T) {
	_, err := NewPasswordlessAuth(types.ConfigPasswordless{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: FuncUserLogout function is required" {
		t.Fatal("Error SHOULD BE '', but found ", "'"+err.Error()+"'")
	}
}

func TestNewPasswordlessAuth_FuncUserStoreTokenFuncUserStoreTokenIsRequired(t *testing.T) {
	_, err := NewPasswordlessAuth(types.ConfigPasswordless{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogout: func(ctx context.Context, userID string, options types.UserAuthOptions) (err error) { return nil },
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: FuncUserStoreToken function is required" {
		t.Fatal("Error SHOULD BE '', but found ", "'"+err.Error()+"'")
	}
}

func TestNewPasswordlessAuth_FuncEmailSendIsRequired(t *testing.T) {
	_, err := NewPasswordlessAuth(types.ConfigPasswordless{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogout:         func(ctx context.Context, userID string, options types.UserAuthOptions) (err error) { return nil },
		FuncUserStoreAuthToken: func(ctx context.Context, sessionID, userID string, options types.UserAuthOptions) error { return nil },
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: FuncEmailSend function is required" {
		t.Fatal("Error SHOULD BE '', but found ", "'"+err.Error()+"'")
	}
}

func TestNewPasswordlessAuth_UseCookiesAndLocalStorageCannotBeBothFalse(t *testing.T) {
	_, err := NewPasswordlessAuth(types.ConfigPasswordless{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogout:         func(ctx context.Context, userID string, options types.UserAuthOptions) (err error) { return nil },
		FuncUserStoreAuthToken: func(ctx context.Context, sessionID, userID string, options types.UserAuthOptions) error { return nil },
		FuncEmailSend:          func(ctx context.Context, email, emailSubject, emailBody string) (err error) { return nil },
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: UseCookies and UseLocalStorage cannot be both false" {
		t.Fatal("Error SHOULD BE '', but found ", "'"+err.Error()+"'")
	}
}

func TestNewPasswordlessAuth_UseCookiesAndLocalStorageCannotBeBothTrue(t *testing.T) {
	_, err := NewPasswordlessAuth(types.ConfigPasswordless{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogout:         func(ctx context.Context, userID string, options types.UserAuthOptions) (err error) { return nil },
		FuncUserStoreAuthToken: func(ctx context.Context, sessionID, userID string, options types.UserAuthOptions) error { return nil },
		FuncEmailSend:          func(ctx context.Context, email, emailSubject, emailBody string) (err error) { return nil },
		UseCookies:             true,
		UseLocalStorage:        true,
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: UseCookies and UseLocalStorage cannot be both true" {
		t.Fatal("Error SHOULD BE '', but found ", "'"+err.Error()+"'")
	}
}

func TestNewPasswordlessAuth_UseCookiesAndLocalStorageCannotBeBothTruee(t *testing.T) {
	auth, err := NewPasswordlessAuth(types.ConfigPasswordless{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogout:         func(ctx context.Context, userID string, options types.UserAuthOptions) (err error) { return nil },
		FuncUserStoreAuthToken: func(ctx context.Context, sessionID, userID string, options types.UserAuthOptions) error { return nil },
		FuncEmailSend:          func(ctx context.Context, email, emailSubject, emailBody string) (err error) { return nil },
		UseCookies:             true,
		UseLocalStorage:        false,
	})

	if err != nil {
		t.Fatal("Error SHOULD BE NULL, but found ", "'"+err.Error()+"'")
	}

	if auth == nil {
		t.Fatal("Auth SHOULD NOT be NULL, but found NULL")
	}
}

func TestNewPasswordlessAuth_CookieConfigName(t *testing.T) {
	auth, err := NewPasswordlessAuth(types.ConfigPasswordless{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogout:         func(ctx context.Context, userID string, options types.UserAuthOptions) (err error) { return nil },
		FuncUserStoreAuthToken: func(ctx context.Context, sessionID, userID string, options types.UserAuthOptions) error { return nil },
		FuncEmailSend:          func(ctx context.Context, email, emailSubject, emailBody string) (err error) { return nil },
		UseCookies:             true,
		CookieConfig: &types.CookieConfig{
			Name: "custom-auth",
		},
	})
	if err != nil {
		t.Fatal("Error SHOULD BE NULL, but found ", "'"+err.Error()+"'")
	}
	if auth.GetCookieName() != "custom-auth" {
		t.Fatalf("expected cookie name %q, got %q", "custom-auth", auth.GetCookieName())
	}

	concrete, ok := auth.(*authImplementation)
	if !ok {
		t.Fatal("expected *authImplementation concrete type")
	}
	if !types.ResolveHttpOnly(concrete.cookieConfig.HttpOnly) {
		t.Fatal("expected HttpOnly to be true when only CookieConfig.Name is set")
	}
	if !types.ResolveSecure(concrete.cookieConfig.Secure) {
		t.Fatal("expected Secure to be true when only CookieConfig.Name is set")
	}
}

func TestNewPasswordlessAuth_CookieConfigInsecure(t *testing.T) {
	auth, err := NewPasswordlessAuth(types.ConfigPasswordless{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogout:         func(ctx context.Context, userID string, options types.UserAuthOptions) (err error) { return nil },
		FuncUserStoreAuthToken: func(ctx context.Context, sessionID, userID string, options types.UserAuthOptions) error { return nil },
		FuncEmailSend:          func(ctx context.Context, email, emailSubject, emailBody string) (err error) { return nil },
		UseCookies:             true,
		CookieConfig: &types.CookieConfig{
			Name:   "custom-auth",
			Secure: types.CookieInsecure,
		},
	})
	if err != nil {
		t.Fatal("Error SHOULD BE NULL, but found ", "'"+err.Error()+"'")
	}
	concrete, ok := auth.(*authImplementation)
	if !ok {
		t.Fatal("expected *authImplementation concrete type")
	}
	if types.ResolveSecure(concrete.cookieConfig.Secure) {
		t.Fatal("expected Secure to be false when CookieConfig.Secure is CookieInsecure")
	}
	if !types.ResolveHttpOnly(concrete.cookieConfig.HttpOnly) {
		t.Fatal("expected HttpOnly to remain true when only Secure is overridden")
	}
}

func TestNewPasswordlessAuth_CSRFSecretRequiredWhenEnabled(t *testing.T) {
	_, err := NewPasswordlessAuth(types.ConfigPasswordless{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogout:         func(ctx context.Context, userID string, options types.UserAuthOptions) (err error) { return nil },
		FuncUserStoreAuthToken: func(ctx context.Context, sessionID, userID string, options types.UserAuthOptions) error { return nil },
		FuncEmailSend:          func(ctx context.Context, email, emailSubject, emailBody string) (err error) { return nil },
		UseCookies:             true,
		UseLocalStorage:        false,
		EnableCSRFProtection:   true,
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: CSRFSecret is required when EnableCSRFProtection is true" {
		t.Fatal("Error SHOULD BE '', but found ", "'"+err.Error()+"'")
	}
}

func TestNewPasswordlessAuth_CSRFEnabledWithSecretSucceeds(t *testing.T) {
	auth, err := NewPasswordlessAuth(types.ConfigPasswordless{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogout:         func(ctx context.Context, userID string, options types.UserAuthOptions) (err error) { return nil },
		FuncUserStoreAuthToken: func(ctx context.Context, sessionID, userID string, options types.UserAuthOptions) error { return nil },
		FuncEmailSend:          func(ctx context.Context, email, emailSubject, emailBody string) (err error) { return nil },
		UseCookies:             true,
		UseLocalStorage:        false,
		EnableCSRFProtection:   true,
		CSRFSecret:             "super-secret",
	})
	if err != nil {
		t.Fatal("Error SHOULD BE NULL, but found ", "'"+err.Error()+"'")
	}
	if auth == nil {
		t.Fatal("Auth SHOULD NOT be NULL, but found NULL")
	}
	concrete, ok := auth.(*authImplementation)
	if !ok {
		t.Fatal("expected *authImplementation concrete type from NewPasswordlessAuth")
	}
	if !concrete.enableCSRFProtection {
		t.Fatal("enableCSRFProtection SHOULD be true")
	}
	if concrete.csrfSecret != "super-secret" {
		t.Fatal("csrfSecret SHOULD be 'super-secret', but found ", "'"+concrete.csrfSecret+"'")
	}
	if concrete.funcCSRFTokenGenerate == nil {
		t.Fatal("funcCSRFTokenGenerate SHOULD be set when CSRF protection is enabled")
	}
	if concrete.funcCSRFTokenValidate == nil {
		t.Fatal("funcCSRFTokenValidate SHOULD be set when CSRF protection is enabled")
	}
}
