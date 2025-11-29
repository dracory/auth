package api_authenticate_via_username

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestAuthenticateViaUsername_Passwordless_MissingDependency tests that when
// passwordless mode is enabled but the required dependency is nil, an error is returned.
func TestAuthenticateViaUsername_Passwordless_MissingDependency(t *testing.T) {
	deps := Dependencies{
		Passwordless:                true,
		PasswordlessUserFindByEmail: nil, // Missing required dependency
	}

	result, aerr := AuthenticateViaUsername(context.Background(), "test@example.com", "", "", deps)

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
	if aerr == nil {
		t.Fatal("expected error, got nil")
	}
	if aerr.Code != AuthenticateErrorCodeUserLookup {
		t.Errorf("expected error code %q, got %q", AuthenticateErrorCodeUserLookup, aerr.Code)
	}
	if aerr.Message != "Invalid credentials" {
		t.Errorf("expected message %q, got %q", "Invalid credentials", aerr.Message)
	}
}

// TestAuthenticateViaUsername_Passwordless_UserLookupError tests that when
// the user lookup function returns an error, it's properly handled.
func TestAuthenticateViaUsername_Passwordless_UserLookupError(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		PasswordlessUserFindByEmail: func(ctx context.Context, email string) (string, error) {
			return "", errors.New("database error")
		},
	}

	result, aerr := AuthenticateViaUsername(context.Background(), "test@example.com", "", "", deps)

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
	if aerr == nil {
		t.Fatal("expected error, got nil")
	}
	if aerr.Code != AuthenticateErrorCodeUserLookup {
		t.Errorf("expected error code %q, got %q", AuthenticateErrorCodeUserLookup, aerr.Code)
	}
	if aerr.Message != "Invalid credentials" {
		t.Errorf("expected message %q, got %q", "Invalid credentials", aerr.Message)
	}
	if aerr.Err == nil {
		t.Error("expected underlying error to be set")
	}
}

// TestAuthenticateViaUsername_Passwordless_UserNotFound tests that when
// the user lookup returns an empty userID, an error is returned.
func TestAuthenticateViaUsername_Passwordless_UserNotFound(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		PasswordlessUserFindByEmail: func(ctx context.Context, email string) (string, error) {
			return "", nil // Empty userID indicates user not found
		},
	}

	result, aerr := AuthenticateViaUsername(context.Background(), "test@example.com", "", "", deps)

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
	if aerr == nil {
		t.Fatal("expected error, got nil")
	}
	if aerr.Code != AuthenticateErrorCodeUserLookup {
		t.Errorf("expected error code %q, got %q", AuthenticateErrorCodeUserLookup, aerr.Code)
	}
	if aerr.Message != "Invalid credentials" {
		t.Errorf("expected message %q, got %q", "Invalid credentials", aerr.Message)
	}
}

// TestAuthenticateViaUsername_Passwordless_TokenStoreMissing tests that when
// the token store dependency is nil, an error is returned.
func TestAuthenticateViaUsername_Passwordless_TokenStoreMissing(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		PasswordlessUserFindByEmail: func(ctx context.Context, email string) (string, error) {
			return "user-123", nil
		},
		UserStoreAuthToken: nil, // Missing required dependency
	}

	result, aerr := AuthenticateViaUsername(context.Background(), "test@example.com", "", "", deps)

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
	if aerr == nil {
		t.Fatal("expected error, got nil")
	}
	if aerr.Code != AuthenticateErrorCodeTokenStore {
		t.Errorf("expected error code %q, got %q", AuthenticateErrorCodeTokenStore, aerr.Code)
	}
	if aerr.Message != "Failed to process request. Please try again later" {
		t.Errorf("expected message %q, got %q", "Failed to process request. Please try again later", aerr.Message)
	}
}

// TestAuthenticateViaUsername_Passwordless_TokenStoreError tests that when
// storing the auth token fails, an error is returned.
func TestAuthenticateViaUsername_Passwordless_TokenStoreError(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		PasswordlessUserFindByEmail: func(ctx context.Context, email string) (string, error) {
			return "user-123", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string) error {
			return errors.New("database error")
		},
	}

	result, aerr := AuthenticateViaUsername(context.Background(), "test@example.com", "", "", deps)

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
	if aerr == nil {
		t.Fatal("expected error, got nil")
	}
	if aerr.Code != AuthenticateErrorCodeTokenStore {
		t.Errorf("expected error code %q, got %q", AuthenticateErrorCodeTokenStore, aerr.Code)
	}
	if aerr.Message != "Failed to process request. Please try again later" {
		t.Errorf("expected message %q, got %q", "Failed to process request. Please try again later", aerr.Message)
	}
	if aerr.Err == nil {
		t.Error("expected underlying error to be set")
	}
}

// TestAuthenticateViaUsername_Passwordless_Success tests a successful
// passwordless authentication flow.
func TestAuthenticateViaUsername_Passwordless_Success(t *testing.T) {
	var storedToken string
	var storedUserID string

	deps := Dependencies{
		Passwordless: true,
		PasswordlessUserFindByEmail: func(ctx context.Context, email string) (string, error) {
			if email == "test@example.com" {
				return "user-123", nil
			}
			return "", errors.New("user not found")
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string) error {
			storedToken = token
			storedUserID = userID
			return nil
		},
	}

	result, aerr := AuthenticateViaUsername(context.Background(), "test@example.com", "", "", deps)

	if aerr != nil {
		t.Fatalf("expected no error, got %+v", aerr)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Token == "" {
		t.Error("expected non-empty token")
	}
	if storedToken != result.Token {
		t.Errorf("expected stored token %q to match result token %q", storedToken, result.Token)
	}
	if storedUserID != "user-123" {
		t.Errorf("expected stored userID %q, got %q", "user-123", storedUserID)
	}
}

// TestAuthenticateViaUsername_UsernamePassword_MissingDependency tests that when
// username/password mode is enabled but the required dependency is nil, an error is returned.
func TestAuthenticateViaUsername_UsernamePassword_MissingDependency(t *testing.T) {
	deps := Dependencies{
		Passwordless:       false,
		UserFindByUsername: nil, // Missing required dependency
	}

	result, aerr := AuthenticateViaUsername(context.Background(), "testuser", "John", "Doe", deps)

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
	if aerr == nil {
		t.Fatal("expected error, got nil")
	}
	if aerr.Code != AuthenticateErrorCodeUserLookup {
		t.Errorf("expected error code %q, got %q", AuthenticateErrorCodeUserLookup, aerr.Code)
	}
	if aerr.Message != "Invalid credentials" {
		t.Errorf("expected message %q, got %q", "Invalid credentials", aerr.Message)
	}
}

// TestAuthenticateViaUsername_UsernamePassword_UserLookupError tests that when
// the user lookup function returns an error, it's properly handled.
func TestAuthenticateViaUsername_UsernamePassword_UserLookupError(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		UserFindByUsername: func(ctx context.Context, username, firstName, lastName string) (string, error) {
			return "", errors.New("database error")
		},
	}

	result, aerr := AuthenticateViaUsername(context.Background(), "testuser", "John", "Doe", deps)

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
	if aerr == nil {
		t.Fatal("expected error, got nil")
	}
	if aerr.Code != AuthenticateErrorCodeUserLookup {
		t.Errorf("expected error code %q, got %q", AuthenticateErrorCodeUserLookup, aerr.Code)
	}
	if aerr.Message != "Invalid credentials" {
		t.Errorf("expected message %q, got %q", "Invalid credentials", aerr.Message)
	}
	if aerr.Err == nil {
		t.Error("expected underlying error to be set")
	}
}

// TestAuthenticateViaUsername_UsernamePassword_UserNotFound tests that when
// the user lookup returns an empty userID, an error is returned.
func TestAuthenticateViaUsername_UsernamePassword_UserNotFound(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		UserFindByUsername: func(ctx context.Context, username, firstName, lastName string) (string, error) {
			return "", nil // Empty userID indicates user not found
		},
	}

	result, aerr := AuthenticateViaUsername(context.Background(), "testuser", "John", "Doe", deps)

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
	if aerr == nil {
		t.Fatal("expected error, got nil")
	}
	if aerr.Code != AuthenticateErrorCodeUserLookup {
		t.Errorf("expected error code %q, got %q", AuthenticateErrorCodeUserLookup, aerr.Code)
	}
	if aerr.Message != "Invalid credentials" {
		t.Errorf("expected message %q, got %q", "Invalid credentials", aerr.Message)
	}
}

// TestAuthenticateViaUsername_UsernamePassword_TokenStoreMissing tests that when
// the token store dependency is nil, an error is returned.
func TestAuthenticateViaUsername_UsernamePassword_TokenStoreMissing(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		UserFindByUsername: func(ctx context.Context, username, firstName, lastName string) (string, error) {
			return "user-456", nil
		},
		UserStoreAuthToken: nil, // Missing required dependency
	}

	result, aerr := AuthenticateViaUsername(context.Background(), "testuser", "John", "Doe", deps)

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
	if aerr == nil {
		t.Fatal("expected error, got nil")
	}
	if aerr.Code != AuthenticateErrorCodeTokenStore {
		t.Errorf("expected error code %q, got %q", AuthenticateErrorCodeTokenStore, aerr.Code)
	}
	if aerr.Message != "Failed to process request. Please try again later" {
		t.Errorf("expected message %q, got %q", "Failed to process request. Please try again later", aerr.Message)
	}
}

// TestAuthenticateViaUsername_UsernamePassword_TokenStoreError tests that when
// storing the auth token fails, an error is returned.
func TestAuthenticateViaUsername_UsernamePassword_TokenStoreError(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		UserFindByUsername: func(ctx context.Context, username, firstName, lastName string) (string, error) {
			return "user-456", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string) error {
			return errors.New("database error")
		},
	}

	result, aerr := AuthenticateViaUsername(context.Background(), "testuser", "John", "Doe", deps)

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
	if aerr == nil {
		t.Fatal("expected error, got nil")
	}
	if aerr.Code != AuthenticateErrorCodeTokenStore {
		t.Errorf("expected error code %q, got %q", AuthenticateErrorCodeTokenStore, aerr.Code)
	}
	if aerr.Message != "Failed to process request. Please try again later" {
		t.Errorf("expected message %q, got %q", "Failed to process request. Please try again later", aerr.Message)
	}
	if aerr.Err == nil {
		t.Error("expected underlying error to be set")
	}
}

// TestAuthenticateViaUsername_UsernamePassword_Success tests a successful
// username/password authentication flow.
func TestAuthenticateViaUsername_UsernamePassword_Success(t *testing.T) {
	var storedToken string
	var storedUserID string

	deps := Dependencies{
		Passwordless: false,
		UserFindByUsername: func(ctx context.Context, username, firstName, lastName string) (string, error) {
			if username == "testuser" && firstName == "John" && lastName == "Doe" {
				return "user-456", nil
			}
			return "", errors.New("user not found")
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string) error {
			storedToken = token
			storedUserID = userID
			return nil
		},
	}

	result, aerr := AuthenticateViaUsername(context.Background(), "testuser", "John", "Doe", deps)

	if aerr != nil {
		t.Fatalf("expected no error, got %+v", aerr)
	}
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if result.Token == "" {
		t.Error("expected non-empty token")
	}
	if storedToken != result.Token {
		t.Errorf("expected stored token %q to match result token %q", storedToken, result.Token)
	}
	if storedUserID != "user-456" {
		t.Errorf("expected stored userID %q, got %q", "user-456", storedUserID)
	}
}

// TestApiAuthenticateViaUsername_Passwordless_Success tests the HTTP handler
// for a successful passwordless authentication.
func TestApiAuthenticateViaUsername_Passwordless_Success(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		PasswordlessUserFindByEmail: func(ctx context.Context, email string) (string, error) {
			return "user-123", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string) error {
			return nil
		},
		UseCookies: false,
	}

	req := httptest.NewRequest("POST", "/api/authenticate", nil)
	recorder := httptest.NewRecorder()

	ApiAuthenticateViaUsername(recorder, req, "test@example.com", "", "", deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"success"`) {
		t.Errorf("expected success status, got %q", body)
	}
	if !strings.Contains(body, `"message":"login success"`) {
		t.Errorf("expected success message, got %q", body)
	}
	if !strings.Contains(body, `"token":`) {
		t.Errorf("expected token in response, got %q", body)
	}
}

// TestApiAuthenticateViaUsername_UsernamePassword_Success tests the HTTP handler
// for a successful username/password authentication.
func TestApiAuthenticateViaUsername_UsernamePassword_Success(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		UserFindByUsername: func(ctx context.Context, username, firstName, lastName string) (string, error) {
			return "user-456", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string) error {
			return nil
		},
		UseCookies: false,
	}

	req := httptest.NewRequest("POST", "/api/authenticate", nil)
	recorder := httptest.NewRecorder()

	ApiAuthenticateViaUsername(recorder, req, "testuser", "John", "Doe", deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"success"`) {
		t.Errorf("expected success status, got %q", body)
	}
	if !strings.Contains(body, `"message":"login success"`) {
		t.Errorf("expected success message, got %q", body)
	}
	if !strings.Contains(body, `"token":`) {
		t.Errorf("expected token in response, got %q", body)
	}
}

// TestApiAuthenticateViaUsername_WithCookies tests that when cookies are enabled,
// the SetAuthCookie function is called.
func TestApiAuthenticateViaUsername_WithCookies(t *testing.T) {
	var cookieSet bool
	var cookieToken string

	deps := Dependencies{
		Passwordless: false,
		UserFindByUsername: func(ctx context.Context, username, firstName, lastName string) (string, error) {
			return "user-456", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string) error {
			return nil
		},
		UseCookies: true,
		SetAuthCookie: func(w http.ResponseWriter, r *http.Request, token string) {
			cookieSet = true
			cookieToken = token
		},
	}

	req := httptest.NewRequest("POST", "/api/authenticate", nil)
	recorder := httptest.NewRecorder()

	ApiAuthenticateViaUsername(recorder, req, "testuser", "John", "Doe", deps)

	if !cookieSet {
		t.Error("expected SetAuthCookie to be called")
	}
	if cookieToken == "" {
		t.Error("expected cookie token to be set")
	}

	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"success"`) {
		t.Errorf("expected success status, got %q", body)
	}
}

// TestApiAuthenticateViaUsername_Error tests that errors are properly
// returned in the HTTP response.
func TestApiAuthenticateViaUsername_Error(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		UserFindByUsername: func(ctx context.Context, username, firstName, lastName string) (string, error) {
			return "", errors.New("database error")
		},
	}

	req := httptest.NewRequest("POST", "/api/authenticate", nil)
	recorder := httptest.NewRecorder()

	ApiAuthenticateViaUsername(recorder, req, "testuser", "John", "Doe", deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"error"`) {
		t.Errorf("expected error status, got %q", body)
	}
	if !strings.Contains(body, `"message":"Invalid credentials"`) {
		t.Errorf("expected error message, got %q", body)
	}
}

// TestAuthenticateError_Error tests the Error() method of AuthenticateError.
func TestAuthenticateError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AuthenticateError
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name: "error with message",
			err: &AuthenticateError{
				Code:    AuthenticateErrorCodeUserLookup,
				Message: "Invalid credentials",
			},
			expected: "Invalid credentials",
		},
		{
			name: "error with underlying error",
			err: &AuthenticateError{
				Code: AuthenticateErrorCodeTokenStore,
				Err:  errors.New("database error"),
			},
			expected: "database error",
		},
		{
			name: "error with code only",
			err: &AuthenticateError{
				Code: AuthenticateErrorCodeCodeGen,
			},
			expected: "code_generation",
		},
		{
			name: "error with message takes precedence",
			err: &AuthenticateError{
				Code:    AuthenticateErrorCodeUserLookup,
				Message: "Custom message",
				Err:     errors.New("underlying error"),
			},
			expected: "Custom message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
