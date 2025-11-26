package tests

import (
	"testing"

	"github.com/dracory/auth"
	assert "github.com/dracory/auth/tests/testassert"
)

func TestEndpointIsRequired(t *testing.T) {
	// expected := `<title>Home | Rem.land</title>`
	// assert.HTTPBodyContainsf(t, routes.Router().ServeHTTP, "POST", links.Home(), url.Values{}, expected, "%")
	_, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{})
	assert.NotNil(t, err)
	assert.Equal(t, "auth: endpoint is required", err.Error())
}

func TestUrlRedirectOnSuccessIsRequired(t *testing.T) {
	// expected := `<title>Home | Rem.land</title>`
	// assert.HTTPBodyContainsf(t, routes.Router().ServeHTTP, "POST", links.Home(), url.Values{}, expected, "%")
	_, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
		Endpoint: "http://localhost/auth",
	})
	assert.NotNil(t, err)
	assert.Equal(t, "auth: url to redirect to on success is required", err.Error())
}

func TestFuncTemporaryKeyGetIsRequired(t *testing.T) {
	// expected := `<title>Home | Rem.land</title>`
	// assert.HTTPBodyContainsf(t, routes.Router().ServeHTTP, "POST", links.Home(), url.Values{}, expected, "%")
	_, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
		Endpoint:             "http://localhost/auth",
		UrlRedirectOnSuccess: "http://localhost/dashboard",
	})
	assert.NotNil(t, err)
	assert.Equal(t, "auth: FuncTemporaryKeyGet function is required", err.Error())
}

func TestFuncTemporaryKeySetIsRequired(t *testing.T) {
	// expected := `<title>Home | Rem.land</title>`
	// assert.HTTPBodyContainsf(t, routes.Router().ServeHTTP, "POST", links.Home(), url.Values{}, expected, "%")
	_, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
		Endpoint:             "http://localhost/auth",
		UrlRedirectOnSuccess: "http://localhost/dashboard",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
	})
	assert.NotNil(t, err)
	assert.Equal(t, "auth: FuncTemporaryKeySet function is required", err.Error())
}

func TestFuncUserFindByTokenIsRequired(t *testing.T) {
	_, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
		Endpoint:             "http://localhost/auth",
		UrlRedirectOnSuccess: "http://localhost/dashboard",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key string, value string, expiresSeconds int) (err error) { return nil },
	})
	assert.NotNil(t, err)
	assert.Equal(t, "auth: FuncUserFindByAuthToken function is required", err.Error())
}

func TestFuncUserFindByUsernameIsRequired(t *testing.T) {
	_, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
		Endpoint:                "http://localhost/auth",
		UrlRedirectOnSuccess:    "http://localhost/dashboard",
		FuncTemporaryKeyGet:     func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:     func(key string, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(token string, options auth.UserAuthOptions) (userID string, err error) { return "", nil },
	})
	assert.NotNil(t, err)
	assert.Equal(t, "auth: FuncUserFindByUsername function is required", err.Error())
}

func TestInitializationSuccess(t *testing.T) {
	authInstance, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
		Endpoint:                "http://localhost/auth",
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
		UseLocalStorage:        false,
	})
	assert.Nil(t, err)
	assert.NotNil(t, authInstance)
}

func TestLinks(t *testing.T) {
	endpoint := "http://localhost/auth"
	authentication, err := newAuthForTests()

	assert.Nil(t, err)
	assert.NotNil(t, authentication)
	assert.Equal(t, endpoint+"/"+auth.PathApiLogin, authentication.LinkApiLogin())
	assert.Equal(t, endpoint+"/"+auth.PathApiLogout, authentication.LinkApiLogout())
	assert.Equal(t, endpoint+"/"+auth.PathApiRegister, authentication.LinkApiRegister())
	assert.Equal(t, endpoint+"/"+auth.PathApiResetPassword, authentication.LinkApiPasswordReset())
	assert.Equal(t, endpoint+"/"+auth.PathApiRestorePassword, authentication.LinkApiPasswordRestore())

	assert.Equal(t, endpoint+"/"+auth.PathLogin, authentication.LinkLogin())
	assert.Equal(t, endpoint+"/"+auth.PathLogout, authentication.LinkLogout())
	assert.Equal(t, endpoint+"/"+auth.PathPasswordReset+"?t=mytoken", authentication.LinkPasswordReset("mytoken"))
	assert.Equal(t, endpoint+"/"+auth.PathPasswordRestore, authentication.LinkPasswordRestore())
	assert.Equal(t, endpoint+"/"+auth.PathRegister, authentication.LinkRegister())
}

func newAuthForTests() (*auth.Auth, error) {
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
