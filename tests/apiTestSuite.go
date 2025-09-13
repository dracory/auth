package tests

import (
	"net/url"

	"github.com/dracory/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type apiTestSuite struct {
	suite.Suite
}

func (suite *apiTestSuite) SetupTest() {
	// config.SetupTests()
}

func (suite *apiTestSuite) TestLoginEndpointRequiresEmail() {
	authentication, err := suite.newAuthWithRegistrationDisabled()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), authentication)

	expectedError := `"status":"error"`
	assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkApiLogin(), url.Values{}, expectedError, "%")

	expectedErrorMessage := `"message":"Email is required field"`
	assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkApiLogin(), url.Values{}, expectedErrorMessage, "%")
}

func (suite *apiTestSuite) TestLoginEndpointRequiresPassword() {
	auth, err := suite.newAuthWithRegistrationDisabled()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), auth)

	expectedError := `"status":"error"`
	assert.HTTPBodyContainsf(suite.T(), auth.Router().ServeHTTP, "POST", auth.LinkApiLogin(), url.Values{
		"email": {"test@test.com"},
	}, expectedError, "%")

	expectedErrorMessage := `"message":"Password is required field"`
	assert.HTTPBodyContainsf(suite.T(), auth.Router().ServeHTTP, "POST", auth.LinkApiLogin(), url.Values{
		"email": {"test@test.com"},
	}, expectedErrorMessage, "%")
}

func (suite *apiTestSuite) TestLoginEndpointRequiresPasswords() {
	auth, err := suite.newAuthWithRegistrationDisabled()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), auth)

	expectedSuccess := `"status":"success"`
	assert.HTTPBodyContainsf(suite.T(), auth.Router().ServeHTTP, "POST", auth.LinkApiLogin(), url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}, expectedSuccess, "%")

	expectedSuccessMessage := `"message":"login success"`
	assert.HTTPBodyContainsf(suite.T(), auth.Router().ServeHTTP, "POST", auth.LinkApiLogin(), url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}, expectedSuccessMessage, "%")

	expectedToken := `"token":"`
	assert.HTTPBodyContainsf(suite.T(), auth.Router().ServeHTTP, "POST", auth.LinkApiLogin(), url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}, expectedToken, "%")
}

func (suite *apiTestSuite) TestRegisterEndpointRequiresFirstName() {
	auth, err := suite.newAuthWithRegistrationDisabled()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), auth)

	expectedError := `"status":"error"`
	assert.HTTPBodyContainsf(suite.T(), auth.Router().ServeHTTP, "POST", auth.LinkApiRegister(), url.Values{}, expectedError, "%")

	expectedErrorMessage := `"message":"First name is required field"`
	assert.HTTPBodyContainsf(suite.T(), auth.Router().ServeHTTP, "POST", auth.LinkApiRegister(), url.Values{}, expectedErrorMessage, "%")
}

func (suite *apiTestSuite) TestRegisterEndpointRequiresLastName() {
	auth, err := suite.newAuthWithRegistrationDisabled()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), auth)

	expectedError := `"status":"error"`
	assert.HTTPBodyContainsf(suite.T(), auth.Router().ServeHTTP, "POST", auth.LinkApiRegister(), url.Values{
		"first_name": {"John"},
	}, expectedError, "%")

	expectedErrorMessage := `"message":"Last name is required field"`
	assert.HTTPBodyContainsf(suite.T(), auth.Router().ServeHTTP, "POST", auth.LinkApiRegister(), url.Values{
		"first_name": {"John"},
	}, expectedErrorMessage, "%")
}

func (suite *apiTestSuite) TestRegisterEndpointRequiresEmail() {
	authentication, err := suite.newAuthWithRegistrationDisabled()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), authentication)

	expectedError := `"status":"error"`
	assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkApiRegister(), url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}, expectedError, "%")

	expectedErrorMessage := `"message":"Email is required field"`
	assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkApiRegister(), url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}, expectedErrorMessage, "%")
}

func (suite *apiTestSuite) TestRegisterEndpointRequiresPassword() {
	authentication, err := suite.newAuthWithRegistrationDisabled()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), authentication)

	expectedError := `"status":"error"`
	assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkApiRegister(), url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}, expectedError, "%")

	expectedErrorMessage := `"message":"Password is required field"`
	assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkApiRegister(), url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}, expectedErrorMessage, "%")
}

func (suite *apiTestSuite) TestRegisterEndpointRequiresPasswords() {
	auth, err := suite.newAuthWithRegistrationEnabled()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), auth)

	expectedSuccess := `"status":"success"`
	assert.HTTPBodyContainsf(suite.T(), auth.Router().ServeHTTP, "POST", auth.LinkApiRegister(), url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
		"password":   {"1234"},
	}, expectedSuccess, "%")

	expectedMessage := `"message":"registration success"`
	assert.HTTPBodyContainsf(suite.T(), auth.Router().ServeHTTP, "POST", auth.LinkApiRegister(), url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
		"password":   {"1234"},
	}, expectedMessage, "%")
}

func (suite *apiTestSuite) TestPasswordlessLoginEndpointRequiresEmail() {
	authentication, err := suite.newAuthPasswordless()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), authentication)

	expectedError := `"status":"error"`
	assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkApiLogin(), url.Values{}, expectedError, "%")

	expectedErrorMessage := `"message":"Email is required field"`
	assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkApiLogin(), url.Values{}, expectedErrorMessage, "%")
}

func (suite *apiTestSuite) TestPasswordlessLoginEndpointSendsLoginCodeEmail() {
	authentication, err := suite.newAuthPasswordless()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), authentication)

	expectedSuccess := `"status":"success"`
	assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkApiLogin(), url.Values{
		"email": {"test@test.com"},
	}, expectedSuccess, "%")

	expectedMessage := `"message":"Login code was sent successfully"`
	assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkApiLogin(), url.Values{
		"email": {"test@test.com"},
	}, expectedMessage, "%")
}

func (suite *apiTestSuite) TestPasswordlessLoginCodeVerifyEndpointRequiresVerificationCode() {
	authentication, err := suite.newAuthPasswordless()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), authentication)

	expectedError := `"status":"error"`
	assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkApiLoginCodeVerify(), url.Values{}, expectedError, "%")

	expectedErrorMessage := `"message":"Verification code is required field"`
	assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkApiLoginCodeVerify(), url.Values{}, expectedErrorMessage, "%")
}

// TODO
func (suite *apiTestSuite) TestPasswordlessLoginCodeVerifyEndpointVerifiesEmail() {
	authentication, err := suite.newAuthPasswordless()

	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), authentication)

	expectedErrorMessage := `"message":"Verification code is invalid length"`
	assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkApiLoginCodeVerify(), url.Values{
		"verification_code": {"123456"},
	}, expectedErrorMessage, "%")

	expectedErrorMessage2 := `"message":"Verification code contains invalid characters"`
	assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkApiLoginCodeVerify(), url.Values{
		"verification_code": {"12345678"},
	}, expectedErrorMessage2, "%")

	// expectedSuccess := `"status":"success"`
	// assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkApiLoginCodeVerify(), url.Values{
	// 	"verification_code": {"123456"},
	// }, expectedSuccess, "%")

	// expectedMessage := `"message":"Your code is correct"`
	// assert.HTTPBodyContainsf(suite.T(), authentication.Router().ServeHTTP, "POST", authentication.LinkApiLoginCodeVerify(), url.Values{
	// 	"email": {"123456"},
	// }, expectedMessage, "%")
}

func (suite *apiTestSuite) newAuthPasswordless() (*auth.Auth, error) {
	endpoint := "http://localhost"
	return auth.NewPasswordlessAuth(auth.ConfigPasswordless{
		Endpoint:                endpoint,
		UrlRedirectOnSuccess:    "http://localhost/dashboard",
		FuncTemporaryKeyGet:     func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:     func(key string, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(token string, options auth.UserAuthOptions) (userID string, err error) { return "111", nil },
		FuncUserFindByEmail:     func(email string, options auth.UserAuthOptions) (userID string, err error) { return "111", nil },
		FuncUserLogout:          func(userID string, options auth.UserAuthOptions) (err error) { return nil },
		FuncUserStoreAuthToken:  func(sessionID string, userID string, options auth.UserAuthOptions) (err error) { return nil },
		FuncEmailSend:           func(email string, emailSubject string, emailBody string) (err error) { return nil },
		UseCookies:              true,
	})
}

func (suite *apiTestSuite) newAuthWithRegistrationDisabled() (*auth.Auth, error) {
	endpoint := "http://localhost"
	return auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
		Endpoint:                endpoint,
		UrlRedirectOnSuccess:    "http://localhost/dashboard",
		FuncTemporaryKeyGet:     func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:     func(key string, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(token string, options auth.UserAuthOptions) (userID string, err error) { return "111", nil },
		FuncUserFindByUsername: func(username string, firstName string, lastName string, options auth.UserAuthOptions) (userID string, err error) {
			return "111", nil
		},
		FuncUserLogin: func(username string, password string, options auth.UserAuthOptions) (userID string, err error) {
			return "111", nil
		},
		FuncUserLogout:         func(userID string, options auth.UserAuthOptions) (err error) { return nil },
		FuncUserStoreAuthToken: func(sessionID string, userID string, options auth.UserAuthOptions) (err error) { return nil },
		FuncEmailSend:          func(userID string, emailSubject string, emailBody string) (err error) { return nil },
		UseCookies:             true,
	})
}

func (suite *apiTestSuite) newAuthWithRegistrationEnabled() (*auth.Auth, error) {
	endpoint := "http://localhost"
	return auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
		Endpoint:                endpoint,
		UrlRedirectOnSuccess:    "http://localhost/dashboard",
		FuncTemporaryKeyGet:     func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:     func(key string, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(token string, options auth.UserAuthOptions) (userID string, err error) { return "111", nil },
		FuncUserFindByUsername: func(username string, firstName string, lastName string, options auth.UserAuthOptions) (userID string, err error) {
			return "111", nil
		},
		FuncUserLogin: func(username string, password string, options auth.UserAuthOptions) (userID string, err error) {
			return "111", nil
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
