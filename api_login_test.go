package auth

import (
	"errors"
	"net/url"
	"testing"

	"github.com/dracory/auth/tests/testassert"
)

func TestApiLoginUsernameAndPasswordRequiresEmail(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	expectedStatus := `"status":"error"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), url.Values{}, expectedStatus, "%")

	expectedMessage := `"message":"Email is required field"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), url.Values{}, expectedMessage, "%")
}

func TestApiLoginUsernameAndPasswordRequiresPassword(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	values := url.Values{
		"email": {"test@test.com"},
	}

	expectedMessage := `"message":"Password is required field"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")
}

func TestApiLoginUsernameAndPasswordInvalidEmail(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	values := url.Values{
		"email":    {"invalid-email"},
		"password": {"1234"},
	}

	expectedMessage := `"message":"This is not a valid email: invalid-email"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")
}

func TestApiLoginUsernameAndPasswordUserLoginError(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	authInstance.funcUserLogin = func(username string, password string, options UserAuthOptions) (userID string, err error) {
		return "", errors.New("db error")
	}

	values := url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}

	expectedMessage := `"message":"authentication failed. db error"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")
}

func TestApiLoginUsernameAndPasswordUserNotFound(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	authInstance.funcUserLogin = func(username string, password string, options UserAuthOptions) (userID string, err error) {
		return "", nil
	}

	values := url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}

	expectedMessage := `"message":"User not found"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")
}

func TestApiLoginUsernameAndPasswordTokenStoreError(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	authInstance.funcUserLogin = func(username string, password string, options UserAuthOptions) (userID string, err error) {
		return "user123", nil
	}

	authInstance.funcUserStoreAuthToken = func(sessionID string, userID string, options UserAuthOptions) (err error) {
		return errors.New("db error")
	}

	values := url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}

	expectedMessage := `"message":"token store failed. db error"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")
}

func TestApiLoginUsernameAndPasswordSuccess(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	authInstance.funcUserLogin = func(username string, password string, options UserAuthOptions) (userID string, err error) {
		return "user123", nil
	}

	values := url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}

	expectedStatus := `"status":"success"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedStatus, "%")

	expectedMessage := `"message":"login success"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")

	expectedToken := `"token":"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedToken, "%")
}

func TestApiLoginPasswordlessRequiresEmail(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	expectedStatus := `"status":"error"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), url.Values{}, expectedStatus, "%")

	expectedMessage := `"message":"Email is required field"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), url.Values{}, expectedMessage, "%")
}

func TestApiLoginPasswordlessInvalidEmail(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	values := url.Values{
		"email":             {"invalid-email"},
		"verification_code": {"CODE1234"},
	}

	expectedMessage := `"message":"This is not a valid email: invalid-email"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")
}

func TestApiLoginPasswordlessTokenStoreError(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	authInstance.funcTemporaryKeySet = func(key string, value string, expiresSeconds int) (err error) {
		return errors.New("db error")
	}

	values := url.Values{
		"email":             {"test@test.com"},
		"verification_code": {"CODE1234"},
	}

	expectedMessage := `"message":"token store failed. db error"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")
}

func TestApiLoginPasswordlessEmailSendError(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	authInstance.passwordlessFuncEmailSend = func(email string, emailSubject string, emailBody string) (err error) {
		return errors.New("smtp error")
	}

	values := url.Values{
		"email":             {"test@test.com"},
		"verification_code": {"CODE1234"},
	}

	expectedMessage := `"message":"Login code failed to be send. Please try again later"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")
}

func TestApiLoginPasswordlessSuccess(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	values := url.Values{
		"email":             {"test@test.com"},
		"verification_code": {"CODE1234"},
	}

	expectedStatus := `"status":"success"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedStatus, "%")

	expectedMessage := `"message":"Login code was sent successfully"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")
}
