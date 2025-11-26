package auth

import (
	"errors"
	"net/url"
	"testing"

	"github.com/dracory/auth/tests/testassert"
)

// Username and password registration tests

func TestApiRegisterUsernameAndPasswordRequiresFirstName(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	expectedStatus := `"status":"error"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), url.Values{}, expectedStatus, "%")

	expectedMessage := `"message":"First name is required field"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), url.Values{}, expectedMessage, "%")
}

func TestApiRegisterUsernameAndPasswordRequiresLastName(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	values := url.Values{
		"first_name": {"John"},
	}

	expectedMessage := `"message":"Last name is required field"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterUsernameAndPasswordRequiresEmail(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}

	expectedMessage := `"message":"Email is required field"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterUsernameAndPasswordRequiresPassword(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}

	expectedMessage := `"message":"Password is required field"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterUsernameAndPasswordInvalidEmail(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"invalid-email"},
		"password":   {"1234"},
	}

	expectedMessage := `"message":"This is not a valid email: invalid-email"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterUsernameAndPasswordFuncUserRegisterNotDefined(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	// funcUserRegister is nil by default in testSetupUsernameAndPasswordAuth
	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
		"password":   {"1234"},
	}

	expectedMessage := `"message":"registration failed. FuncUserRegister function not defined"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterUsernameAndPasswordRegistrationFailed(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

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
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterUsernameAndPasswordSuccess(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

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
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedStatus, "%")

	expectedMessage := `"message":"registration success"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

// Passwordless registration tests

func TestApiRegisterPasswordlessRequiresFirstName(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	expectedStatus := `"status":"error"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), url.Values{}, expectedStatus, "%")

	expectedMessage := `"message":"First name is required field"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), url.Values{}, expectedMessage, "%")
}

func TestApiRegisterPasswordlessRequiresLastName(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	values := url.Values{
		"first_name": {"John"},
	}

	expectedMessage := `"message":"Last name is required field"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterPasswordlessRequiresEmail(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}

	expectedMessage := `"message":"Email is required field"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterPasswordlessTokenStoreError(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	authInstance.funcTemporaryKeySet = func(key string, value string, expiresSeconds int) (err error) {
		return errors.New("db error")
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}

	expectedMessage := `"message":"token store failed. db error"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterPasswordlessEmailSendError(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	authInstance.passwordlessFuncEmailSend = func(email string, emailSubject string, emailBody string) (err error) {
		return errors.New("smtp error")
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}

	expectedMessage := `"message":"Registration code failed to be send. Please try again later"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}

func TestApiRegisterPasswordlessSuccess(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}

	expectedStatus := `"status":"success"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedStatus, "%")

	expectedMessage := `"message":"Registration code was sent successfully"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegister(), values, expectedMessage, "%")
}
