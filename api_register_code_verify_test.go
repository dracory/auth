package auth

import (
	"errors"
	"net/url"
	"testing"

	"github.com/dracory/auth/tests/testassert"
)

func TestApiRegisterCodeVerifyRequiresVerificationCode(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	expectedStatus := `"status":"error"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), url.Values{}, expectedStatus, "%")

	expectedMessage := `"message":"Verification code is required field"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), url.Values{}, expectedMessage, "%")
}

func TestApiRegisterCodeVerifyInvalidLength(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	values := url.Values{
		"verification_code": {"123456"},
	}

	expectedMessage := `"message":"Verification code is invalid length"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedMessage, "%")
}

func TestApiRegisterCodeVerifyInvalidCharacters(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	values := url.Values{
		"verification_code": {"12345678"},
	}

	expectedMessage := `"message":"Verification code contains invalid characters"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedMessage, "%")
}

func TestApiRegisterCodeVerifyExpiredCode(t *testing.T) {
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
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedMessage, "%")
}

func TestApiRegisterCodeVerifyMalformedJSON(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return "not-json", nil
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	expectedMessage := `"message":"Serialized format is malformed"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedMessage, "%")
}

func TestApiRegisterCodeVerifyRegistrationFailed(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	// Valid JSON payload
	jsonPayload := `{"email":"test@test.com","first_name":"John","last_name":"Doe","password":"1234"}`
	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return jsonPayload, nil
	}

	authInstance.passwordlessFuncUserRegister = func(email string, firstName string, lastName string, options UserAuthOptions) (err error) {
		return errors.New("db error")
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	expectedMessage := `"message":"registration failed. db error"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedMessage, "%")
}

func TestApiRegisterCodeVerifyAuthenticationError(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	jsonPayload := `{"email":"test@test.com","first_name":"John","last_name":"Doe","password":"1234"}`
	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return jsonPayload, nil
	}

	authInstance.passwordlessFuncUserRegister = func(email string, firstName string, lastName string, options UserAuthOptions) (err error) {
		return nil
	}

	authInstance.passwordlessFuncUserFindByEmail = func(email string, options UserAuthOptions) (userID string, err error) {
		return "", errors.New("db error")
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	expectedMessage := `"message":"authentication failed. db error"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedMessage, "%")
}

func TestApiRegisterCodeVerifyUserNotFound(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	jsonPayload := `{"email":"test@test.com","first_name":"John","last_name":"Doe","password":"1234"}`
	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return jsonPayload, nil
	}

	authInstance.passwordlessFuncUserRegister = func(email string, firstName string, lastName string, options UserAuthOptions) (err error) {
		return nil
	}

	authInstance.passwordlessFuncUserFindByEmail = func(email string, options UserAuthOptions) (userID string, err error) {
		return "", nil
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	expectedMessage := `"message":"authentication failed. user not found"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedMessage, "%")
}

func TestApiRegisterCodeVerifyTokenStoreError(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	jsonPayload := `{"email":"test@test.com","first_name":"John","last_name":"Doe","password":"1234"}`
	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return jsonPayload, nil
	}

	authInstance.passwordlessFuncUserRegister = func(email string, firstName string, lastName string, options UserAuthOptions) (err error) {
		return nil
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
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedMessage, "%")
}

func TestApiRegisterCodeVerifySuccess(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	jsonPayload := `{"email":"test@test.com","first_name":"John","last_name":"Doe","password":"1234"}`
	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return jsonPayload, nil
	}

	authInstance.passwordlessFuncUserRegister = func(email string, firstName string, lastName string, options UserAuthOptions) (err error) {
		return nil
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
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedStatus, "%")

	expectedMessage := `"message":"login success"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedMessage, "%")

	expectedToken := `"token":"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedToken, "%")
}
