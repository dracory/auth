package auth

import (
	"errors"
	"net/url"
	"testing"

	"github.com/dracory/auth/tests/testassert"
)

func TestApiLoginCodeVerifyRequiresVerificationCode(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	expectedStatus := `"status":"error"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), url.Values{}, expectedStatus, "%")

	expectedMessage := `"message":"Verification code is required field"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), url.Values{}, expectedMessage, "%")
}

func TestApiLoginCodeVerifyInvalidLength(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	values := url.Values{
		"verification_code": {"123456"},
	}

	expectedMessage := `"message":"Verification code is invalid length"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), values, expectedMessage, "%")
}

func TestApiLoginCodeVerifyInvalidCharacters(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	values := url.Values{
		"verification_code": {"12345678"},
	}

	expectedMessage := `"message":"Verification code contains invalid characters"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), values, expectedMessage, "%")
}

func TestApiLoginCodeVerifyExpiredCode(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return "", errors.New("expired")
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	expectedMessage := `"message":"Verification code has expired"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), values, expectedMessage, "%")
}

func TestApiLoginCodeVerifyAuthenticationError(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return "user@example.com", nil
	}

	authInstance.passwordlessFuncUserFindByEmail = func(email string, options UserAuthOptions) (userID string, err error) {
		return "", errors.New("db error")
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	expectedMessage := `"message":"authentication failed. db error"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), values, expectedMessage, "%")
}

func TestApiLoginCodeVerifyUserNotFound(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return "user@example.com", nil
	}

	authInstance.passwordlessFuncUserFindByEmail = func(email string, options UserAuthOptions) (userID string, err error) {
		return "", nil
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	expectedMessage := `"message":"authentication failed. user not found"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), values, expectedMessage, "%")
}

func TestApiLoginCodeVerifyTokenStoreError(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return "user@example.com", nil
	}

	authInstance.passwordlessFuncUserFindByEmail = func(email string, options UserAuthOptions) (userID string, err error) {
		return "user123", nil
	}

	authInstance.funcUserStoreAuthToken = func(sessionID string, userID string, options UserAuthOptions) (err error) {
		return errors.New("db error")
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	expectedMessage := `"message":"token store failed. db error"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), values, expectedMessage, "%")
}

func TestApiLoginCodeVerifySuccess(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return "user@example.com", nil
	}

	authInstance.passwordlessFuncUserFindByEmail = func(email string, options UserAuthOptions) (userID string, err error) {
		return "user123", nil
	}

	authInstance.funcUserStoreAuthToken = func(sessionID string, userID string, options UserAuthOptions) (err error) {
		return nil
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	expectedStatus := `"status":"success"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), values, expectedStatus, "%")

	expectedMessage := `"message":"login success"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), values, expectedMessage, "%")

	expectedToken := `"token":"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), values, expectedToken, "%")
}
