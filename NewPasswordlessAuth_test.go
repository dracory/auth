package auth

import (
	"testing"
)

func TestNewPasswordlessAuth_EndpointRequired(t *testing.T) {
	_, err := NewPasswordlessAuth(ConfigPasswordless{})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: endpoint is required" {
		t.Fatal("Error SHOULD BE '', but found '", err.Error(), "'")
	}
}

func TestNewPasswordlessAuth_UrlToRedirectOnSuccessIsRequired(t *testing.T) {
	_, err := NewPasswordlessAuth(ConfigPasswordless{
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
	_, err := NewPasswordlessAuth(ConfigPasswordless{
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
	_, err := NewPasswordlessAuth(ConfigPasswordless{
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
	_, err := NewPasswordlessAuth(ConfigPasswordless{
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
	_, err := NewPasswordlessAuth(ConfigPasswordless{
		Endpoint:                "/auth",
		UrlRedirectOnSuccess:    "/user",
		FuncTemporaryKeyGet:     func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:     func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(sessionID string, userIP string, userAgent string) (userID string, err error) { return "", nil },
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: FuncUserFindByEmail function is required" {
		t.Fatal("Error SHOULD BE '', but found ", "'"+err.Error()+"'")
	}
}

func TestNewPasswordlessAuth_FuncUserLogoutIsRequired(t *testing.T) {
	_, err := NewPasswordlessAuth(ConfigPasswordless{
		Endpoint:                "/auth",
		UrlRedirectOnSuccess:    "/user",
		FuncTemporaryKeyGet:     func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:     func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(sessionID string, userIP string, userAgent string) (userID string, err error) { return "", nil },
		FuncUserFindByEmail:     func(email string) (userID string, err error) { return "", nil },
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: FuncUserLogout function is required" {
		t.Fatal("Error SHOULD BE '', but found ", "'"+err.Error()+"'")
	}
}

func TestNewPasswordlessAuth_FuncUserStoreTokenFuncUserStoreTokenIsRequired(t *testing.T) {
	_, err := NewPasswordlessAuth(ConfigPasswordless{
		Endpoint:                "/auth",
		UrlRedirectOnSuccess:    "/user",
		FuncTemporaryKeyGet:     func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:     func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(sessionID string, userIP string, userAgent string) (userID string, err error) { return "", nil },
		FuncUserFindByEmail:     func(email string) (userID string, err error) { return "", nil },
		FuncUserLogout:          func(userID string) (err error) { return nil },
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: FuncUserStoreToken function is required" {
		t.Fatal("Error SHOULD BE '', but found ", "'"+err.Error()+"'")
	}
}

func TestNewPasswordlessAuth_FuncEmailSendIsRequired(t *testing.T) {
	_, err := NewPasswordlessAuth(ConfigPasswordless{
		Endpoint:                "/auth",
		UrlRedirectOnSuccess:    "/user",
		FuncTemporaryKeyGet:     func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:     func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(sessionID string, userIP string, userAgent string) (userID string, err error) { return "", nil },
		FuncUserFindByEmail:     func(email string) (userID string, err error) { return "", nil },
		FuncUserLogout:          func(userID string) (err error) { return nil },
		FuncUserStoreAuthToken:  func(sessionID, userID string, userIP string, userAgent string) error { return nil },
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: FuncEmailSend function is required" {
		t.Fatal("Error SHOULD BE '', but found ", "'"+err.Error()+"'")
	}
}

func TestNewPasswordlessAuth_UseCookiesAndLocalStorageCannotBeBothFalse(t *testing.T) {
	_, err := NewPasswordlessAuth(ConfigPasswordless{
		Endpoint:                "/auth",
		UrlRedirectOnSuccess:    "/user",
		FuncTemporaryKeyGet:     func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:     func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(sessionID string, userIP string, userAgent string) (userID string, err error) { return "", nil },
		FuncUserFindByEmail:     func(email string) (userID string, err error) { return "", nil },
		FuncUserLogout:          func(userID string) (err error) { return nil },
		FuncUserStoreAuthToken:  func(sessionID, userID string, userIP string, userAgent string) error { return nil },
		FuncEmailSend:           func(email, emailSubject, emailBody string) (err error) { return nil },
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: UseCookies and UseLocalStorage cannot be both false" {
		t.Fatal("Error SHOULD BE '', but found ", "'"+err.Error()+"'")
	}
}

func TestNewPasswordlessAuth_UseCookiesAndLocalStorageCannotBeBothTrue(t *testing.T) {
	_, err := NewPasswordlessAuth(ConfigPasswordless{
		Endpoint:                "/auth",
		UrlRedirectOnSuccess:    "/user",
		FuncTemporaryKeyGet:     func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:     func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(sessionID string, userIP string, userAgent string) (userID string, err error) { return "", nil },
		FuncUserFindByEmail:     func(email string) (userID string, err error) { return "", nil },
		FuncUserLogout:          func(userID string) (err error) { return nil },
		FuncUserStoreAuthToken:  func(sessionID, userID string, userIP string, userAgent string) error { return nil },
		FuncEmailSend:           func(email, emailSubject, emailBody string) (err error) { return nil },
		UseCookies:              true,
		UseLocalStorage:         true,
	})
	if err == nil {
		t.Fatal("Error SHOULD NOT BE NULL")
	}
	if err.Error() != "auth: UseCookies and UseLocalStorage cannot be both true" {
		t.Fatal("Error SHOULD BE '', but found ", "'"+err.Error()+"'")
	}
}

func TestNewPasswordlessAuth_UseCookiesAndLocalStorageCannotBeBothTruee(t *testing.T) {
	auth, err := NewPasswordlessAuth(ConfigPasswordless{
		Endpoint:                "/auth",
		UrlRedirectOnSuccess:    "/user",
		FuncTemporaryKeyGet:     func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:     func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(sessionID string, userIP string, userAgent string) (userID string, err error) { return "", nil },
		FuncUserFindByEmail:     func(email string) (userID string, err error) { return "", nil },
		FuncUserLogout:          func(userID string) (err error) { return nil },
		FuncUserStoreAuthToken:  func(sessionID, userID string, userIP string, userAgent string) error { return nil },
		FuncEmailSend:           func(email, emailSubject, emailBody string) (err error) { return nil },
		UseCookies:              true,
		UseLocalStorage:         false,
	})

	if err != nil {
		t.Fatal("Error SHOULD BE NULL, but found ", "'"+err.Error()+"'")
	}

	if auth == nil {
		t.Fatal("Auth SHOULD NOT be NULL, but found NULL")
	}
}
