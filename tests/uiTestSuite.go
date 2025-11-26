package tests

import (
	"net/url"
	"testing"

	"github.com/dracory/auth"
	assert "github.com/dracory/auth/tests/testassert"
)

func TestPageLogin(t *testing.T) {
	authentication, err := newUIAuthWithRegistrationDisabled()

	assert.Nil(t, err)
	assert.NotNil(t, authentication)
	expected := `<title>Login</title>`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkLogin(), url.Values{}, expected, "%")
}

func TestPageRegister(t *testing.T) {
	authentication, err := newUIAuthWithRegistrationEnabled()

	assert.Nil(t, err)
	assert.NotNil(t, authentication)
	expected := `<title>Register</title>`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkRegister(), url.Values{}, expected, "%")
}

func TestPageRegisterDisabled(t *testing.T) {
	authentication, err := newUIAuthWithRegistrationDisabled()

	assert.Nil(t, err)
	assert.NotNil(t, authentication)
	expectedTitle := `<title>Register</title>`
	assert.HTTPBodyNotContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkRegister(), url.Values{}, expectedTitle, "%")
	expectedEmpty := ""
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkRegister(), url.Values{}, expectedEmpty, "%")
}

func TestPagePasswordRestore(t *testing.T) {
	authentication, err := newUIAuthWithRegistrationEnabled()

	assert.Nil(t, err)
	assert.NotNil(t, authentication)
	expected := `<title>Restore Password</title>`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkPasswordRestore(), url.Values{}, expected, "%")
}

func TestPagePasswordReset(t *testing.T) {
	authentication, err := newUIAuthWithRegistrationEnabled()

	assert.Nil(t, err)
	assert.NotNil(t, authentication)
	expected := `<title>Reset Password</title>`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkPasswordReset("testtoken"), url.Values{}, expected, "%")
}

func newUIAuthWithRegistrationDisabled() (*auth.Auth, error) {
	endpoint := "http://localhost/auth"
	return auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
		Endpoint:                endpoint,
		UrlRedirectOnSuccess:    "http://localhost/dashboard",
		FuncTemporaryKeyGet:     func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:     func(key string, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(token string, options auth.UserAuthOptions) (userID string, err error) { return "", nil },
		FuncUserFindByUsername: func(username string, firstName string, lastName string, options auth.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogin: func(username string, password string, options auth.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogout:         func(userID string, options auth.UserAuthOptions) (err error) { return nil },
		FuncUserStoreAuthToken: func(sessionID string, userID string, options auth.UserAuthOptions) (err error) { return nil },
		FuncEmailSend:          func(userID string, emailSubject string, emailBody string) (err error) { return nil },
		UseCookies:             true,
	})
}

func newUIAuthWithRegistrationEnabled() (*auth.Auth, error) {
	endpoint := "http://localhost/auth"
	return auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
		Endpoint:                endpoint,
		UrlRedirectOnSuccess:    "http://localhost/dashboard",
		FuncTemporaryKeyGet:     func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:     func(key string, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(token string, options auth.UserAuthOptions) (userID string, err error) { return "", nil },
		FuncUserFindByUsername: func(username string, firstName string, lastName string, options auth.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogin: func(username string, password string, options auth.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserRegister: func(username string, password string, firstName string, lastName string, options auth.UserAuthOptions) (err error) {
			return nil
		},
		FuncUserLogout:         func(userID string, options auth.UserAuthOptions) (err error) { return nil },
		FuncUserStoreAuthToken: func(sessionID string, userID string, options auth.UserAuthOptions) (err error) { return nil },
		FuncEmailSend:          func(userID string, emailSubject string, emailBody string) (err error) { return nil },
		EnableRegistration:     true,
		UseCookies:             true,
	})
}
