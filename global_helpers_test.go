package auth

import (
	"context"
	"testing"

	"github.com/dracory/auth/types"
)

func newAuthForRegisterTests() *authImplementation {
	return &authImplementation{}
}

// These tests now act as high-level checks that the public wrapper
// RegisterWithUsernameAndPassword correctly delegates into the core
// logic for the two main flows: non-verification and verification.

func TestRegisterWithUsernameAndPassword_RegistrationSuccess_NoVerification(t *testing.T) {
	authInstance := newAuthForRegisterTests()

	// Configure minimal dependencies for non-verification flow.
	authInstance.funcUserRegister = func(ctx context.Context, username string, password string, firstName string, lastName string, options types.UserAuthOptions) error {
		return nil
	}

	resp := RegisterWithUsernameAndPassword(context.Background(), authInstance, "test@test.com", "password", "John", "Doe", types.UserAuthOptions{})

	if resp.ErrorMessage != "" {
		t.Fatalf("expected no error, got %q", resp.ErrorMessage)
	}

	expectedSuccess := "registration success"
	if resp.SuccessMessage != expectedSuccess {
		t.Fatalf("expected success message %q, got %q", expectedSuccess, resp.SuccessMessage)
	}
}

func newAuthForLoginTests() *authImplementation {
	return &authImplementation{}
}

// High-level wrapper test to ensure that LoginWithUsernameAndPassword
// correctly delegates into the core logic and returns a successful
// response when dependencies are configured.
func TestLoginWithUsernameAndPassword_Success(t *testing.T) {
	authInstance := newAuthForLoginTests()

	authInstance.funcUserLogin = func(ctx context.Context, username string, password string, options types.UserAuthOptions) (userID string, err error) {
		return "user123", nil
	}

	// Ensure token store succeeds
	authInstance.funcUserStoreAuthToken = func(ctx context.Context, token string, userID string, options types.UserAuthOptions) error {
		if token == "" {
			t.Fatalf("expected non-empty token in FuncUserStoreAuthToken")
		}
		return nil
	}

	resp := LoginWithUsernameAndPassword(context.Background(), authInstance, "test@test.com", "password", types.UserAuthOptions{})

	if resp.ErrorMessage != "" {
		t.Fatalf("expected no error, got %q", resp.ErrorMessage)
	}

	if resp.SuccessMessage != "login success" {
		t.Fatalf("expected success message %q, got %q", "login success", resp.SuccessMessage)
	}

	if resp.Token == "" {
		t.Fatalf("expected non-empty token on success")
	}
}

func TestRegisterWithUsernameAndPassword_VerificationEnabled_Success(t *testing.T) {
	authInstance := newAuthForRegisterTests()

	authInstance.enableVerification = true
	authInstance.funcUserRegister = func(ctx context.Context, username string, password string, firstName string, lastName string, options types.UserAuthOptions) error {
		return nil
	}

	authInstance.funcTemporaryKeySet = func(key string, value string, expiresSeconds int) error {
		return nil
	}

	authInstance.funcEmailTemplateRegisterCode = func(ctx context.Context, email string, code string, options types.UserAuthOptions) string {
		return "body"
	}

	authInstance.funcEmailSend = func(ctx context.Context, userID string, emailSubject string, emailBody string) error {
		return nil
	}

	resp := RegisterWithUsernameAndPassword(context.Background(), authInstance, "test@test.com", "password", "John", "Doe", types.UserAuthOptions{})

	if resp.ErrorMessage != "" {
		t.Fatalf("expected no error, got %q", resp.ErrorMessage)
	}

	expectedSuccess := "Registration code was sent successfully"
	if resp.SuccessMessage != expectedSuccess {
		t.Fatalf("expected success message %q, got %q", expectedSuccess, resp.SuccessMessage)
	}
}
