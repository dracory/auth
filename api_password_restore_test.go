package auth

import (
	"context"
	"errors"
	"net/url"
	"testing"
)

func TestPasswordRestoreEndpointRequiresEmail(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() error = %v", err)
	}
	if authInstance == nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() returned nil auth instance")
	}

	expectedError := `"status":"error"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordRestore(), url.Values{}, expectedError, "%")

	expectedErrorMessage := `"message":"Email is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordRestore(), url.Values{}, expectedErrorMessage, "%")
}

func TestPasswordRestoreEndpointRequiresFirstName(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() error = %v", err)
	}
	if authInstance == nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() returned nil auth instance")
	}

	expectedErrorMessage := `"message":"First name is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordRestore(), url.Values{
		"email": {"test@test.com"},
	}, expectedErrorMessage, "%")
}

func TestPasswordRestoreEndpointRequiresLastName(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() error = %v", err)
	}
	if authInstance == nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() returned nil auth instance")
	}

	expectedErrorMessage := `"message":"Last name is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiPasswordRestore(), url.Values{
		"email":      {"test@test.com"},
		"first_name": {"John"},
	}, expectedErrorMessage, "%")
}

func TestPasswordRestoreEndpointUserNotFound(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() error = %v", err)
	}
	if authInstance == nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() returned nil auth instance")
	}

	// Mock user not found
	authInstance.funcUserFindByUsername = func(ctx context.Context, username string, firstName string, lastName string, options UserAuthOptions) (userID string, err error) {
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
	if err != nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() error = %v", err)
	}
	if authInstance == nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() returned nil auth instance")
	}

	// Mock internal error
	authInstance.funcUserFindByUsername = func(ctx context.Context, username string, firstName string, lastName string, options UserAuthOptions) (userID string, err error) {
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
	if err != nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() error = %v", err)
	}
	if authInstance == nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() returned nil auth instance")
	}

	// Mock success
	authInstance.funcUserFindByUsername = func(ctx context.Context, username string, firstName string, lastName string, options UserAuthOptions) (userID string, err error) {
		return "user123", nil
	}
	authInstance.funcTemporaryKeySet = func(key string, value string, expiresSeconds int) (err error) {
		return nil
	}
	authInstance.funcEmailSend = func(ctx context.Context, userID string, emailSubject string, emailBody string) (err error) {
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
