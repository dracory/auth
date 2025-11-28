package auth

import (
	"context"
	"errors"
	"net/url"
	"testing"
)

func TestApiRegisterCodeVerifyRequiresVerificationCode(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	expectedStatus := `"status":"error"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), url.Values{}, expectedStatus, "%")

	expectedMessage := `"message":"Verification code is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), url.Values{}, expectedMessage, "%")
}

func TestApiRegisterCodeVerifyInvalidLength(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	values := url.Values{
		"verification_code": {"123456"},
	}

	expectedMessage := `"message":"Verification code is invalid length"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedMessage, "%")
}

func TestApiRegisterCodeVerifyInvalidCharacters(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	values := url.Values{
		"verification_code": {"12345678"},
	}

	expectedMessage := `"message":"Verification code contains invalid characters"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedMessage, "%")
}

func TestApiRegisterCodeVerifyExpiredCode(t *testing.T) {
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
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedMessage, "%")
}

func TestApiRegisterCodeVerifyMalformedJSON(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return "not-json", nil
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	expectedMessage := `"message":"Serialized format is malformed"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedMessage, "%")
}

func TestApiRegisterCodeVerifyRegistrationFailed(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	// Valid JSON payload
	jsonPayload := `{"email":"test@test.com","first_name":"John","last_name":"Doe","password":"1234"}`
	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return jsonPayload, nil
	}

	authInstance.passwordlessFuncUserRegister = func(ctx context.Context, email string, firstName string, lastName string, options UserAuthOptions) (err error) {
		return errors.New("db error")
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	expectedMessage := `"message":"registration failed."`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedMessage, "%")
}

func TestApiRegisterCodeVerifyAuthenticationError(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	jsonPayload := `{"email":"test@test.com","first_name":"John","last_name":"Doe","password":"1234"}`
	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return jsonPayload, nil
	}

	authInstance.passwordlessFuncUserRegister = func(ctx context.Context, email string, firstName string, lastName string, options UserAuthOptions) (err error) {
		return nil
	}

	authInstance.passwordlessFuncUserFindByEmail = func(ctx context.Context, email string, options UserAuthOptions) (userID string, err error) {
		return "", errors.New("db error")
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	expectedMessage := `"message":"Invalid credentials"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedMessage, "%")
}

func TestApiRegisterCodeVerifyUserNotFound(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	jsonPayload := `{"email":"test@test.com","first_name":"John","last_name":"Doe","password":"1234"}`
	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return jsonPayload, nil
	}

	authInstance.passwordlessFuncUserRegister = func(ctx context.Context, email string, firstName string, lastName string, options UserAuthOptions) (err error) {
		return nil
	}

	authInstance.passwordlessFuncUserFindByEmail = func(ctx context.Context, email string, options UserAuthOptions) (userID string, err error) {
		return "", nil
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	expectedMessage := `"message":"Invalid credentials"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedMessage, "%")
}

func TestApiRegisterCodeVerifyTokenStoreError(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	jsonPayload := `{"email":"test@test.com","first_name":"John","last_name":"Doe","password":"1234"}`
	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return jsonPayload, nil
	}

	authInstance.passwordlessFuncUserRegister = func(ctx context.Context, email string, firstName string, lastName string, options UserAuthOptions) (err error) {
		return nil
	}

	authInstance.passwordlessFuncUserFindByEmail = func(ctx context.Context, email string, options UserAuthOptions) (userID string, err error) {
		return "user123", nil
	}

	authInstance.funcUserStoreAuthToken = func(ctx context.Context, token string, userID string, options UserAuthOptions) error {
		return errors.New("db error")
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	expectedMessage := `"message":"token store failed."`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedMessage, "%")
}

func TestApiRegisterCodeVerifySuccess(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	jsonPayload := `{"email":"test@test.com","first_name":"John","last_name":"Doe","password":"1234"}`
	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return jsonPayload, nil
	}

	authInstance.passwordlessFuncUserRegister = func(ctx context.Context, email string, firstName string, lastName string, options UserAuthOptions) (err error) {
		return nil
	}

	authInstance.passwordlessFuncUserFindByEmail = func(ctx context.Context, email string, options UserAuthOptions) (userID string, err error) {
		return "user123", nil
	}

	authInstance.funcUserStoreAuthToken = func(ctx context.Context, token string, userID string, options UserAuthOptions) error {
		return nil
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	expectedStatus := `"status":"success"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedStatus, "%")

	expectedMessage := `"message":"login success"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedMessage, "%")

	expectedToken := `"token":"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiRegisterCodeVerify(), values, expectedToken, "%")
}
