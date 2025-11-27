package auth

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
)

// Username and password registration tests

func TestApiRegisterUsernameAndPasswordRequiresFirstName(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	expectedStatus := `"status":"error"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), url.Values{}, expectedStatus, "%")

	expectedMessage := `"message":"First name is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), url.Values{}, expectedMessage, "%")
}

func TestApiRegisterUsernameAndPasswordRequiresLastName(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	values := url.Values{
		"first_name": {"John"},
	}

	expectedMessage := `"message":"Last name is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterUsernameAndPasswordRequiresEmail(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}

	expectedMessage := `"message":"Email is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterUsernameAndPasswordRequiresPassword(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}

	expectedMessage := `"message":"Password is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterUsernameAndPasswordInvalidEmail(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"invalid-email"},
		"password":   {"1234"},
	}

	expectedMessage := `"message":"This is not a valid email: invalid-email"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterUsernameAndPasswordFuncUserRegisterNotDefined(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	// funcUserRegister is nil by default in testSetupUsernameAndPasswordAuth
	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
		"password":   {"1234"},
	}

	expectedMessage := `"message":"registration failed. FuncUserRegister function not defined"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterUsernameAndPasswordRegistrationFailed(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	authInstance.funcUserRegister = func(username string, password string, firstName string, lastName string, options UserAuthOptions) (err error) {
		return errors.New("db error")
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
		"password":   {"1234"},
	}

	expectedMessage := `"message":"registration failed. db error"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterUsernameAndPasswordSuccess(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	authInstance.funcUserRegister = func(username string, password string, firstName string, lastName string, options UserAuthOptions) (err error) {
		return nil
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
		"password":   {"1234"},
	}

	expectedStatus := `"status":"success"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedStatus, "%")

	expectedMessage := `"message":"registration success"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterUsernameAndPasswordInvalidCSRFToken(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	authInstance.enableCSRFProtection = true
	authInstance.funcCSRFTokenValidate = func(r *http.Request) bool {
		return false
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
		"password":   {"1234"},
	}

	expectedMessage := `"message":"Invalid CSRF token"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

// Passwordless registration tests

func TestApiRegisterPasswordlessRequiresFirstName(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	expectedStatus := `"status":"error"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), url.Values{}, expectedStatus, "%")

	expectedMessage := `"message":"First name is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), url.Values{}, expectedMessage, "%")
}

func TestApiRegisterPasswordlessRequiresLastName(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	values := url.Values{
		"first_name": {"John"},
	}

	expectedMessage := `"message":"Last name is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterPasswordlessRequiresEmail(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}

	expectedMessage := `"message":"Email is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterPasswordlessTokenStoreError(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	authInstance.funcTemporaryKeySet = func(key string, value string, expiresSeconds int) (err error) {
		return errors.New("db error")
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}

	expectedMessage := `"message":"token store failed. db error"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterPasswordlessEmailSendError(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	authInstance.passwordlessFuncEmailSend = func(email string, emailSubject string, emailBody string) (err error) {
		return errors.New("smtp error")
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}

	expectedMessage := `"message":"Registration code failed to be send. Please try again later"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterPasswordlessSuccess(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}

	expectedStatus := `"status":"success"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedStatus, "%")

	expectedMessage := `"message":"Registration code was sent successfully"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}
