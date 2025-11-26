package tests

import (
	"net/url"
	"testing"

	"github.com/dracory/auth"
	assert "github.com/dracory/auth/tests/testassert"
)

func TestLoginEndpointRequiresEmail(t *testing.T) {
	authentication, err := newAuthWithRegistrationDisabled()

	assert.Nil(t, err)
	assert.NotNil(t, authentication)

	expectedError := `"status":"error"`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkApiLogin(), url.Values{}, expectedError, "%")

	expectedErrorMessage := `"message":"Email is required field"`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkApiLogin(), url.Values{}, expectedErrorMessage, "%")
}

func TestLoginEndpointRequiresPassword(t *testing.T) {
	authInstance, err := newAuthWithRegistrationDisabled()

	assert.Nil(t, err)
	assert.NotNil(t, authInstance)

	expectedError := `"status":"error"`
	assert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), url.Values{
		"email": {"test@test.com"},
	}, expectedError, "%")

	expectedErrorMessage := `"message":"Password is required field"`
	assert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), url.Values{
		"email": {"test@test.com"},
	}, expectedErrorMessage, "%")
}

func TestLoginEndpointRequiresPasswords(t *testing.T) {
	authInstance, err := newAuthWithRegistrationDisabled()

	assert.Nil(t, err)
	assert.NotNil(t, authInstance)

	expectedSuccess := `"status":"success"`
	assert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}, expectedSuccess, "%")

	expectedSuccessMessage := `"message":"login success"`
	assert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}, expectedSuccessMessage, "%")

	expectedToken := `"token":"`
	assert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}, expectedToken, "%")
}

func TestRegisterEndpointRequiresFirstName(t *testing.T) {
	authInstance, err := newAuthWithRegistrationDisabled()

	assert.Nil(t, err)
	assert.NotNil(t, authInstance)

	expectedError := `"status":"error"`
	assert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), url.Values{}, expectedError, "%")

	expectedErrorMessage := `"message":"First name is required field"`
	assert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), url.Values{}, expectedErrorMessage, "%")
}

func TestRegisterEndpointRequiresLastName(t *testing.T) {
	authInstance, err := newAuthWithRegistrationDisabled()

	assert.Nil(t, err)
	assert.NotNil(t, authInstance)

	expectedError := `"status":"error"`
	assert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), url.Values{
		"first_name": {"John"},
	}, expectedError, "%")

	expectedErrorMessage := `"message":"Last name is required field"`
	assert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), url.Values{
		"first_name": {"John"},
	}, expectedErrorMessage, "%")
}

func TestRegisterEndpointRequiresEmail(t *testing.T) {
	authentication, err := newAuthWithRegistrationDisabled()

	assert.Nil(t, err)
	assert.NotNil(t, authentication)

	expectedError := `"status":"error"`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkApiRegister(), url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}, expectedError, "%")

	expectedErrorMessage := `"message":"Email is required field"`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkApiRegister(), url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}, expectedErrorMessage, "%")
}

func TestRegisterEndpointRequiresPassword(t *testing.T) {
	authentication, err := newAuthWithRegistrationDisabled()

	assert.Nil(t, err)
	assert.NotNil(t, authentication)

	expectedError := `"status":"error"`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkApiRegister(), url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}, expectedError, "%")

	expectedErrorMessage := `"message":"Password is required field"`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkApiRegister(), url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}, expectedErrorMessage, "%")
}

func TestRegisterEndpointRequiresPasswords(t *testing.T) {
	authInstance, err := newAuthWithRegistrationEnabled()

	assert.Nil(t, err)
	assert.NotNil(t, authInstance)

	expectedSuccess := `"status":"success"`
	assert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
		"password":   {"1234"},
	}, expectedSuccess, "%")

	expectedMessage := `"message":"registration success"`
	assert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
		"password":   {"1234"},
	}, expectedMessage, "%")
}

func TestPasswordlessLoginEndpointRequiresEmail(t *testing.T) {
	authentication, err := newAuthPasswordless()

	assert.Nil(t, err)
	assert.NotNil(t, authentication)

	expectedError := `"status":"error"`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkApiLogin(), url.Values{}, expectedError, "%")

	expectedErrorMessage := `"message":"Email is required field"`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkApiLogin(), url.Values{}, expectedErrorMessage, "%")
}

func TestPasswordlessLoginEndpointSendsLoginCodeEmail(t *testing.T) {
	authentication, err := newAuthPasswordless()

	assert.Nil(t, err)
	assert.NotNil(t, authentication)

	expectedSuccess := `"status":"success"`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkApiLogin(), url.Values{
		"email": {"test@test.com"},
	}, expectedSuccess, "%")

	expectedMessage := `"message":"Login code was sent successfully"`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkApiLogin(), url.Values{
		"email": {"test@test.com"},
	}, expectedMessage, "%")
}

func TestPasswordlessLoginCodeVerifyEndpointRequiresVerificationCode(t *testing.T) {
	authentication, err := newAuthPasswordless()

	assert.Nil(t, err)
	assert.NotNil(t, authentication)

	expectedError := `"status":"error"`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkApiLoginCodeVerify(), url.Values{}, expectedError, "%")

	expectedErrorMessage := `"message":"Verification code is required field"`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkApiLoginCodeVerify(), url.Values{}, expectedErrorMessage, "%")
}

func TestPasswordlessLoginCodeVerifyEndpointVerifiesEmail(t *testing.T) {
	authentication, err := newAuthPasswordless()

	assert.Nil(t, err)
	assert.NotNil(t, authentication)

	expectedErrorMessage := `"message":"Verification code is invalid length"`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkApiLoginCodeVerify(), url.Values{
		"verification_code": {"123456"},
	}, expectedErrorMessage, "%")

	expectedErrorMessage2 := `"message":"Verification code contains invalid characters"`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkApiLoginCodeVerify(), url.Values{
		"verification_code": {"12345678"},
	}, expectedErrorMessage2, "%")

	expectedSuccess := `"status":"success"`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkApiLoginCodeVerify(), url.Values{
		"verification_code": {"BCDFGHJK"},
	}, expectedSuccess, "%")

	expectedMessage := `"message":"login success"`
	assert.HTTPBodyContainsf(t, authentication.Router().ServeHTTP, "POST", authentication.LinkApiLoginCodeVerify(), url.Values{
		"verification_code": {"BCDFGHJK"},
	}, expectedMessage, "%")
}

func newAuthPasswordless() (*auth.Auth, error) {
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

func newAuthWithRegistrationDisabled() (*auth.Auth, error) {
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

func newAuthWithRegistrationEnabled() (*auth.Auth, error) {
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
