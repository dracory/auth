package auth

import (
	"context"
	"testing"

	"github.com/dracory/auth/types"
)

func TestNewUsernameAndPasswordAuth_EndpointRequired(t *testing.T) {
	_, err := NewUsernameAndPasswordAuth(ConfigUsernameAndPassword{})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: endpoint is required" {
		t.Fatal("Error SHOULD BE '', but found '", err.Error(), "'")
	}
}

func TestNewUsernameAndPasswordAuth_UrlToRedirectOnSuccessIsRequired(t *testing.T) {
	_, err := NewUsernameAndPasswordAuth(ConfigUsernameAndPassword{
		Endpoint: "/auth",
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: url to redirect to on success is required" {
		t.Fatal("Error SHOULD BE '', but found '", err.Error(), "'")
	}
}

func TestNewUsernameAndPasswordAuth_FuncTemporaryKeyGetIsRequired(t *testing.T) {
	_, err := NewUsernameAndPasswordAuth(ConfigUsernameAndPassword{
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

func TestNewUsernameAndPasswordAuth_FuncTemporaryKeySetIsRequired(t *testing.T) {
	_, err := NewUsernameAndPasswordAuth(ConfigUsernameAndPassword{
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

func TestNewUsernameAndPasswordAuth_FuncUserFindByAuthTokenIsRequired(t *testing.T) {
	_, err := NewUsernameAndPasswordAuth(ConfigUsernameAndPassword{
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

func TestNewUsernameAndPasswordAuth_FuncUserFindByUsernameIsRequired(t *testing.T) {
	_, err := NewUsernameAndPasswordAuth(ConfigUsernameAndPassword{
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
	if err.Error() != "auth: FuncUserFindByUsername function is required" {
		t.Fatal("Error SHOULD BE '', but found ", "'"+err.Error()+"'")
	}
}

func TestNewUsernameAndPasswordAuth_FuncUserLoginIsRequired(t *testing.T) {
	_, err := NewUsernameAndPasswordAuth(ConfigUsernameAndPassword{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserFindByUsername: func(ctx context.Context, username, firstName, lastName string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: FuncUserLogin function is required" {
		t.Fatal("Error SHOULD BE '', but found ", "'"+err.Error()+"'")
	}
}

func TestNewUsernameAndPasswordAuth_FuncUserLogoutIsRequired(t *testing.T) {
	_, err := NewUsernameAndPasswordAuth(ConfigUsernameAndPassword{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserFindByUsername: func(ctx context.Context, username, firstName, lastName string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogin: func(ctx context.Context, username, password string, options types.UserAuthOptions) (userID string, err error) {
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

func TestNewUsernameAndPasswordAuth_FuncUserStoreTokenFuncUserStoreTokenIsRequired(t *testing.T) {
	_, err := NewUsernameAndPasswordAuth(ConfigUsernameAndPassword{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserFindByUsername: func(ctx context.Context, username, firstName, lastName string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogin: func(ctx context.Context, username, password string, options types.UserAuthOptions) (userID string, err error) {
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

func TestNewUsernameAndPasswordAuth_FuncEmailSendIsRequired(t *testing.T) {
	_, err := NewUsernameAndPasswordAuth(ConfigUsernameAndPassword{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error { return nil },
		FuncUserFindByUsername: func(ctx context.Context, username, firstName, lastName string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogin: func(ctx context.Context, username, password string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogout: func(ctx context.Context, userID string, options types.UserAuthOptions) (err error) { return nil },
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: FuncEmailSend function is required" {
		t.Fatal("Error SHOULD BE '', but found ", "'"+err.Error()+"'")
	}
}

func TestNewUsernameAndPasswordAuth_UseCookiesAndLocalStorageCannotBeBothFalse(t *testing.T) {
	_, err := NewUsernameAndPasswordAuth(ConfigUsernameAndPassword{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserFindByUsername: func(ctx context.Context, username, firstName, lastName string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogin: func(ctx context.Context, username, password string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogout:         func(ctx context.Context, userID string, options types.UserAuthOptions) (err error) { return nil },
		FuncUserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error { return nil },
		FuncEmailSend:          func(ctx context.Context, email, emailSubject, emailBody string) (err error) { return nil },
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: UseCookies and UseLocalStorage cannot be both false" {
		t.Fatal("Error SHOULD BE '', but found ", "'"+err.Error()+"'")
	}
}

func TestNewUsernameAndPasswordAuth_UseCookiesAndLocalStorageCannotBeBothTrue(t *testing.T) {
	_, err := NewUsernameAndPasswordAuth(ConfigUsernameAndPassword{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogout:         func(ctx context.Context, userID string, options types.UserAuthOptions) (err error) { return nil },
		FuncUserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error { return nil },
		FuncEmailSend:          func(ctx context.Context, email, emailSubject, emailBody string) (err error) { return nil },
		FuncUserFindByUsername: func(ctx context.Context, username, firstName, lastName string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogin: func(ctx context.Context, username, password string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		UseCookies:      true,
		UseLocalStorage: true,
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: UseCookies and UseLocalStorage cannot be both true" {
		t.Fatal("Error SHOULD BE '', but found ", "'"+err.Error()+"'")
	}
}

func TestNewUsernameAndPasswordAuth_UseCookiesAndLocalStorageCannotBeBothTruee(t *testing.T) {
	auth, err := NewUsernameAndPasswordAuth(ConfigUsernameAndPassword{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogout:         func(ctx context.Context, userID string, options types.UserAuthOptions) (err error) { return nil },
		FuncUserStoreAuthToken: func(ctx context.Context, sessionID, userID string, options types.UserAuthOptions) error { return nil },
		FuncEmailSend:          func(ctx context.Context, email, emailSubject, emailBody string) (err error) { return nil },
		FuncUserFindByUsername: func(ctx context.Context, username, firstName, lastName string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogin: func(ctx context.Context, username, password string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		UseCookies:      true,
		UseLocalStorage: false,
	})

	if err != nil {
		t.Fatal("Error SHOULD BE NULL, but found ", "'"+err.Error()+"'")
	}

	if auth == nil {
		t.Fatal("Auth SHOULD NOT be NULL, but found NULL")
	}
}

func TestNewUsernameAndPasswordAuth_CSRFSecretRequiredWhenEnabled(t *testing.T) {
	_, err := NewUsernameAndPasswordAuth(ConfigUsernameAndPassword{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserFindByUsername: func(ctx context.Context, username, firstName, lastName string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogin: func(ctx context.Context, username, password string, options types.UserAuthOptions) (userID string, err error) {
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

func TestNewUsernameAndPasswordAuth_CSRFEnabledWithSecretSucceeds(t *testing.T) {
	auth, err := NewUsernameAndPasswordAuth(ConfigUsernameAndPassword{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserFindByUsername: func(ctx context.Context, username, firstName, lastName string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogin: func(ctx context.Context, username, password string, options types.UserAuthOptions) (userID string, err error) {
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
		t.Fatal("expected *authImplementation concrete type from NewUsernameAndPasswordAuth")
	}
	if !concrete.enableCSRFProtection {
		t.Fatal("enableCSRFProtection SHOULD be true")
	}
	if concrete.csrfSecret != "super-secret" {
		t.Fatal("csrfSecret SHOULD be 'super-secret', but found ", "'"+concrete.csrfSecret+"'")
	}
}
