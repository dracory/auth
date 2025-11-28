package auth

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
)

func TestPasswordResetEndpointRequiresToken(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	expectedError := `"status":"error"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordReset(), url.Values{}, expectedError, "%")

	expectedErrorMessage := `"message":"Token is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordReset(), url.Values{}, expectedErrorMessage, "%")
}

func TestPasswordResetEndpointRequiresPassword(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	expectedErrorMessage := `"message":"Password is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordReset(), url.Values{
		"token": {"valid-token"},
	}, expectedErrorMessage, "%")
}

func TestPasswordResetEndpointRequiresMatchingPasswords(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	expectedErrorMessage := `"message":"Passwords do not match"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordReset(), url.Values{
		"token":            {"valid-token"},
		"password":         {"password123"},
		"password_confirm": {"password321"},
	}, expectedErrorMessage, "%")
}

func TestPasswordResetEndpointInvalidToken(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	// Mock invalid token (returns empty userID or error)
	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return "", nil
	}

	expectedErrorMessage := `"message":"Link not valid of expired"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordReset(), url.Values{
		"token":            {"invalid-token"},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}, expectedErrorMessage, "%")
}

func TestPasswordResetEndpointPasswordChangeError(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	// Mock valid token
	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return "user123", nil
	}

	// Mock password change error
	authInstance.funcUserPasswordChange = func(ctx context.Context, username string, newPassword string, options UserAuthOptions) (err error) {
		return errors.New("db error")
	}

	expectedErrorMessage := `"message":"authentication failed. db error"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordReset(), url.Values{
		"token":            {"valid-token"},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}, expectedErrorMessage, "%")
}

func TestPasswordResetEndpointSuccess(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	// Mock valid token
	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		return "user123", nil
	}

	// Mock success
	authInstance.funcUserPasswordChange = func(ctx context.Context, username string, newPassword string, options UserAuthOptions) (err error) {
		return nil
	}

	expectedSuccess := `"status":"success"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordReset(), url.Values{
		"token":            {"valid-token"},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}, expectedSuccess, "%")

	expectedMessage := `"message":"login success"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordReset(), url.Values{
		"token":            {"valid-token"},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}, expectedMessage, "%")
}

func TestPasswordResetEndpointInvalidCSRFToken(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	Nil(t, err)
	NotNil(t, authInstance)

	authInstance.enableCSRFProtection = true
	authInstance.funcCSRFTokenValidate = func(r *http.Request) bool {
		return false
	}

	values := url.Values{
		"token":            {"valid-token"},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}

	expectedMessage := `"message":"Invalid CSRF token"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordReset(), values, expectedMessage, "%")
}
