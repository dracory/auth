package auth

import (
	"errors"
	"net/url"
	"testing"
)

func TestPasswordRestoreEndpointRequiresEmail(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	expectedError := `"status":"error"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordRestore(), url.Values{}, expectedError, "%")

	expectedErrorMessage := `"message":"Email is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordRestore(), url.Values{}, expectedErrorMessage, "%")
}

func TestPasswordRestoreEndpointRequiresFirstName(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	expectedErrorMessage := `"message":"First name is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordRestore(), url.Values{
		"email": {"test@test.com"},
	}, expectedErrorMessage, "%")
}

func TestPasswordRestoreEndpointRequiresLastName(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	expectedErrorMessage := `"message":"Last name is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordRestore(), url.Values{
		"email":      {"test@test.com"},
		"first_name": {"John"},
	}, expectedErrorMessage, "%")
}

func TestPasswordRestoreEndpointUserNotFound(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	// Mock user not found
	authInstance.funcUserFindByUsername = func(username string, firstName string, lastName string, options UserAuthOptions) (userID string, err error) {
		return "", nil
	}

	expectedErrorMessage := `"message":"User not found"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordRestore(), url.Values{
		"email":      {"test@test.com"},
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}, expectedErrorMessage, "%")
}

func TestPasswordRestoreEndpointInternalError(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	// Mock internal error
	authInstance.funcUserFindByUsername = func(username string, firstName string, lastName string, options UserAuthOptions) (userID string, err error) {
		return "", errors.New("db error")
	}

	expectedErrorMessage := `"message":"Internal server error"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordRestore(), url.Values{
		"email":      {"test@test.com"},
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}, expectedErrorMessage, "%")
}

func TestPasswordRestoreEndpointSuccess(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	// Mock success
	authInstance.funcUserFindByUsername = func(username string, firstName string, lastName string, options UserAuthOptions) (userID string, err error) {
		return "user123", nil
	}
	authInstance.funcTemporaryKeySet = func(key string, value string, expiresSeconds int) (err error) {
		return nil
	}
	authInstance.funcEmailSend = func(userID string, emailSubject string, emailBody string) (err error) {
		return nil
	}

	expectedSuccess := `"status":"success"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordRestore(), url.Values{
		"email":      {"test@test.com"},
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}, expectedSuccess, "%")

	expectedMessage := `"message":"Password reset link was sent to your e-mail"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordRestore(), url.Values{
		"email":      {"test@test.com"},
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}, expectedMessage, "%")
}
