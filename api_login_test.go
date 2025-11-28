package auth

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
)

func TestApiLoginUsernameAndPasswordRequiresEmail(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() error = %v", err)
	}
	if authInstance == nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() returned nil auth instance")
	}

	expectedStatus := `"status":"error"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), url.Values{}, expectedStatus, "%")

	expectedMessage := `"message":"Email is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), url.Values{}, expectedMessage, "%")
}

func TestApiLoginUsernameAndPasswordRequiresPassword(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() error = %v", err)
	}
	if authInstance == nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() returned nil auth instance")
	}

	values := url.Values{
		"email": {"test@test.com"},
	}

	expectedMessage := `"message":"Password is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")
}

func TestApiLoginUsernameAndPasswordUserLoginError(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() error = %v", err)
	}
	if authInstance == nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() returned nil auth instance")
	}

	authInstance.funcUserLogin = func(ctx context.Context, username string, password string, options UserAuthOptions) (userID string, err error) {
		return "", errors.New("db error")
	}

	values := url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}

	expectedMessage := `"message":"Invalid credentials"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")
}

func TestApiLoginUsernameAndPasswordUserNotFound(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() error = %v", err)
	}
	if authInstance == nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() returned nil auth instance")
	}

	authInstance.funcUserLogin = func(ctx context.Context, username string, password string, options UserAuthOptions) (userID string, err error) {
		return "", nil
	}

	values := url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}

	expectedMessage := `"message":"Invalid credentials"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")
}

func TestApiLoginUsernameAndPasswordTokenStoreError(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() error = %v", err)
	}
	if authInstance == nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() returned nil auth instance")
	}

	authInstance.funcUserLogin = func(ctx context.Context, username string, password string, options UserAuthOptions) (userID string, err error) {
		return "user123", nil
	}

	authInstance.funcUserStoreAuthToken = func(ctx context.Context, token string, userID string, options UserAuthOptions) error {
		return errors.New("db error")
	}

	values := url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}

	expectedMessage := `"message":"Failed to process request. Please try again later"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")
}

func TestApiLoginUsernameAndPasswordSuccess(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() error = %v", err)
	}
	if authInstance == nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() returned nil auth instance")
	}

	authInstance.funcUserLogin = func(ctx context.Context, username string, password string, options UserAuthOptions) (userID string, err error) {
		return "user123", nil
	}

	values := url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}

	expectedStatus := `"status":"success"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedStatus, "%")

	expectedMessage := `"message":"login success"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")

	expectedToken := `"token":"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedToken, "%")
}

func TestApiLoginUsernameAndPasswordInvalidCSRFToken(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() error = %v", err)
	}
	if authInstance == nil {
		t.Fatalf("testSetupUsernameAndPasswordAuth() returned nil auth instance")
	}

	authInstance.enableCSRFProtection = true
	authInstance.funcCSRFTokenValidate = func(r *http.Request) bool {
		return false
	}

	values := url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}

	expectedMessage := `"message":"Invalid CSRF token"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")
}

func TestApiLoginPasswordlessRequiresEmail(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	if err != nil {
		t.Fatalf("testSetupPasswordlessAuth() error = %v", err)
	}
	if authInstance == nil {
		t.Fatalf("testSetupPasswordlessAuth() returned nil auth instance")
	}

	expectedStatus := `"status":"error"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), url.Values{}, expectedStatus, "%")

	expectedMessage := `"message":"Email is required field"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), url.Values{}, expectedMessage, "%")
}

func TestApiLoginPasswordlessInvalidEmail(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	if err != nil {
		t.Fatalf("testSetupPasswordlessAuth() error = %v", err)
	}
	if authInstance == nil {
		t.Fatalf("testSetupPasswordlessAuth() returned nil auth instance")
	}

	values := url.Values{
		"email": {"invalid-email"},
	}

	expectedMessage := `"message":"This is not a valid email: invalid-email"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")
}

func TestApiLoginPasswordlessTokenStoreError(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	if err != nil {
		t.Fatalf("testSetupPasswordlessAuth() error = %v", err)
	}
	if authInstance == nil {
		t.Fatalf("testSetupPasswordlessAuth() returned nil auth instance")
	}

	authInstance.funcTemporaryKeySet = func(key string, value string, expiresSeconds int) (err error) {
		return errors.New("db error")
	}

	values := url.Values{
		"email": {"test@test.com"},
	}

	expectedMessage := `"message":"Failed to process request. Please try again later"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")
}

func TestApiLoginPasswordlessEmailSendError(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	if err != nil {
		t.Fatalf("testSetupPasswordlessAuth() error = %v", err)
	}
	if authInstance == nil {
		t.Fatalf("testSetupPasswordlessAuth() returned nil auth instance")
	}

	authInstance.passwordlessFuncEmailSend = func(ctx context.Context, email string, emailSubject string, emailBody string) (err error) {
		return errors.New("smtp error")
	}

	values := url.Values{
		"email": {"test@test.com"},
	}

	expectedMessage := `"message":"Failed to send email. Please try again later"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")
}

func TestApiLoginPasswordlessSuccess(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	if err != nil {
		t.Fatalf("testSetupPasswordlessAuth() error = %v", err)
	}
	if authInstance == nil {
		t.Fatalf("testSetupPasswordlessAuth() returned nil auth instance")
	}

	values := url.Values{
		"email": {"test@test.com"},
	}

	expectedStatus := `"status":"success"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedStatus, "%")

	expectedMessage := `"message":"Login code was sent successfully"`
	HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogin(), values, expectedMessage, "%")
}
