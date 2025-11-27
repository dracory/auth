package auth

import (
	"errors"
	"net/url"
	"testing"
)

func TestApiLoginCodeVerifyRequiresVerificationCode(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	expectedStatus := `"status":"error"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), url.Values{}, expectedStatus, "%")

	expectedMessage := `"message":"Verification code is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), url.Values{}, expectedMessage, "%")
}

func TestApiLoginCodeVerifyInvalidLength(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	values := url.Values{
		"verification_code": {"123456"},
	}

	expectedMessage := `"message":"Verification code is invalid length"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), values, expectedMessage, "%")
}

func TestApiLoginCodeVerifyInvalidCharacters(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	values := url.Values{
		"verification_code": {"12345678"},
	}

	expectedMessage := `"message":"Verification code contains invalid characters"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), values, expectedMessage, "%")
}

func TestApiLoginCodeVerifyExpiredCode(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return "", errors.New("expired")
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	expectedMessage := `"message":"Verification code has expired"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), values, expectedMessage, "%")
}

func TestApiLoginCodeVerifyAuthenticationError(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

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
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), values, expectedMessage, "%")
}

func TestApiLoginCodeVerifyUserNotFound(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

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
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), values, expectedMessage, "%")
}

func TestApiLoginCodeVerifyTokenStoreError(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

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
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), values, expectedMessage, "%")
}

func TestApiLoginCodeVerifySuccess(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

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
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), values, expectedStatus, "%")

	expectedMessage := `"message":"login success"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), values, expectedMessage, "%")

	expectedToken := `"token":"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLoginCodeVerify(), values, expectedToken, "%")
}
