package tests

import (
	"github.com/dracory/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type initTestSuite struct {
	suite.Suite
}

func (suite *initTestSuite) SetupTest() {
	// config.SetupTests()
}

func (suite *initTestSuite) TestEndpointIsRequired() {
	// expected := `<title>Home | Rem.land</title>`
	// assert.HTTPBodyContainsf(suite.T(), routes.Router().ServeHTTP, "POST", links.Home(), url.Values{}, expected, "%")
	_, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{})
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "auth: endpoint is required", err.Error())
}

func (suite *initTestSuite) TesUrlRedirectOnSuccessIsRequired() {
	// expected := `<title>Home | Rem.land</title>`
	// assert.HTTPBodyContainsf(suite.T(), routes.Router().ServeHTTP, "POST", links.Home(), url.Values{}, expected, "%")
	_, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
		Endpoint: "http://localhost/auth",
	})
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "auth: url to redirect to on success is required", err.Error())
}

func (suite *initTestSuite) TestFuncTemporaryKeyGetIsRequired() {
	// expected := `<title>Home | Rem.land</title>`
	// assert.HTTPBodyContainsf(suite.T(), routes.Router().ServeHTTP, "POST", links.Home(), url.Values{}, expected, "%")
	_, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
		Endpoint:             "http://localhost/auth",
		UrlRedirectOnSuccess: "http://localhost/dashboard",
	})
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "auth: FuncTemporaryKeyGet function is required", err.Error())
}

func (suite *initTestSuite) TestFuncTemporaryKeySetIsRequired() {
	// expected := `<title>Home | Rem.land</title>`
	// assert.HTTPBodyContainsf(suite.T(), routes.Router().ServeHTTP, "POST", links.Home(), url.Values{}, expected, "%")
	_, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
		Endpoint:             "http://localhost/auth",
		UrlRedirectOnSuccess: "http://localhost/dashboard",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
	})
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "auth: FuncTemporaryKeySet function is required", err.Error())
}

func (suite *initTestSuite) TestFuncUserFindByTokenIsRequired() {
	_, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
		Endpoint:             "http://localhost/auth",
		UrlRedirectOnSuccess: "http://localhost/dashboard",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key string, value string, expiresSeconds int) (err error) { return nil },
	})
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "auth: FuncUserFindByAuthToken function is required", err.Error())
}

func (suite *initTestSuite) TestFuncUserFindByUsernameIsRequired() {
	_, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
		Endpoint:                "http://localhost/auth",
		UrlRedirectOnSuccess:    "http://localhost/dashboard",
		FuncTemporaryKeyGet:     func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:     func(key string, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(token string, options auth.UserAuthOptions) (userID string, err error) { return "", nil },
	})
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), "auth: FuncUserFindByUsername function is required", err.Error())
}

func (suite *initTestSuite) TestInitializationSuccess() {
	auth, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
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
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), auth)
}

func (suite *initTestSuite) TestLinks() {
	endpoint := "http://localhost/auth"
	authentication, err := suite.newAuth()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), authentication)
	assert.Equal(suite.T(), endpoint+"/"+auth.PathApiLogin, authentication.LinkApiLogin())
	assert.Equal(suite.T(), endpoint+"/"+auth.PathApiLogout, authentication.LinkApiLogout())
	assert.Equal(suite.T(), endpoint+"/"+auth.PathApiRegister, authentication.LinkApiRegister())
	assert.Equal(suite.T(), endpoint+"/"+auth.PathApiResetPassword, authentication.LinkApiPasswordReset())
	assert.Equal(suite.T(), endpoint+"/"+auth.PathApiRestorePassword, authentication.LinkApiPasswordRestore())

	assert.Equal(suite.T(), endpoint+"/"+auth.PathLogin, authentication.LinkLogin())
	assert.Equal(suite.T(), endpoint+"/"+auth.PathLogout, authentication.LinkLogout())
	assert.Equal(suite.T(), endpoint+"/"+auth.PathPasswordReset+"?t=mytoken", authentication.LinkPasswordReset("mytoken"))
	assert.Equal(suite.T(), endpoint+"/"+auth.PathPasswordRestore, authentication.LinkPasswordRestore())
	assert.Equal(suite.T(), endpoint+"/"+auth.PathRegister, authentication.LinkRegister())
}

func (suite *initTestSuite) newAuth() (*auth.Auth, error) {
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
