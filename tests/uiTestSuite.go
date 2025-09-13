package tests

import (
	"net/url"

	"github.com/dracory/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type uiTestSuite struct {
	suite.Suite
}

func (suite *uiTestSuite) SetupTest() {
	// config.SetupTests()
}

func (suite *uiTestSuite) TestPageLogin() {
	authentication, err := suite.newAuthWithRegistrationDisabled()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), authentication)
	expected := `<title>Login</title>`
	assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkLogin(), url.Values{}, expected, "%")
}

func (suite *uiTestSuite) TestPageRegister() {
	authentication, err := suite.newAuthWithRegistrationEnabled()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), authentication)
	expected := `<title>Register</title>`
	assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkRegister(), url.Values{}, expected, "%")
}

func (suite *uiTestSuite) TestPageRegisterDisabled() {
	authentication, err := suite.newAuthWithRegistrationDisabled()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), authentication)
	expectedTitle := `<title>Register</title>`
	assert.HTTPBodyNotContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkRegister(), url.Values{}, expectedTitle, "%")
	expectedEmpty := ""
	assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkRegister(), url.Values{}, expectedEmpty, "%")
}

func (suite *uiTestSuite) TestPagePasswordRestore() {
	authentication, err := suite.newAuthWithRegistrationEnabled()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), authentication)
	expected := `<title>Restore Password</title>`
	assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkPasswordRestore(), url.Values{}, expected, "%")
}

func (suite *uiTestSuite) TestPagePasswordReset() {
	authentication, err := suite.newAuthWithRegistrationEnabled()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), authentication)
	expected := `<title>Reset Password</title>`
	assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkPasswordReset("testtoken"), url.Values{}, expected, "%")
}

func (suite *uiTestSuite) newAuthWithRegistrationDisabled() (*auth.Auth, error) {
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

func (suite *uiTestSuite) newAuthWithRegistrationEnabled() (*auth.Auth, error) {
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
