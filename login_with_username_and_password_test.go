package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/dracory/auth/types"
)

func newAuthForLoginTests() *authImplementation {
	return &authImplementation{}
}

func TestLoginWithUsernameAndPassword_RequiresEmail(t *testing.T) {
	authInstance := newAuthForLoginTests()

	resp := authInstance.LoginWithUsernameAndPassword(context.Background(), "", "password", types.UserAuthOptions{})

	if resp.ErrorMessage != "Email is required field" {
		t.Fatalf("expected error %q, got %q", "Email is required field", resp.ErrorMessage)
	}
}

func TestLoginWithUsernameAndPassword_RequiresPassword(t *testing.T) {
	authInstance := newAuthForLoginTests()

	resp := authInstance.LoginWithUsernameAndPassword(context.Background(), "test@test.com", "", types.UserAuthOptions{})

	if resp.ErrorMessage != "Password is required field" {
		t.Fatalf("expected error %q, got %q", "Password is required field", resp.ErrorMessage)
	}
}

func TestLoginWithUsernameAndPassword_InvalidEmail(t *testing.T) {
	authInstance := newAuthForLoginTests()

	resp := authInstance.LoginWithUsernameAndPassword(context.Background(), "invalid-email", "password", types.UserAuthOptions{})

	expected := "This is not a valid email: invalid-email"
	if resp.ErrorMessage != expected {
		t.Fatalf("expected error %q, got %q", expected, resp.ErrorMessage)
	}
}

func TestLoginWithUsernameAndPassword_LoginError(t *testing.T) {
	authInstance := newAuthForLoginTests()

	authInstance.funcUserLogin = func(ctx context.Context, username string, password string, options types.UserAuthOptions) (userID string, err error) {
		return "", errors.New("db error")
	}

	resp := authInstance.LoginWithUsernameAndPassword(context.Background(), "test@test.com", "password", types.UserAuthOptions{})

	expected := "Invalid credentials"
	if resp.ErrorMessage != expected {
		t.Fatalf("expected error %q, got %q", expected, resp.ErrorMessage)
	}
}

func TestLoginWithUsernameAndPassword_UserNotFound(t *testing.T) {
	authInstance := newAuthForLoginTests()

	authInstance.funcUserLogin = func(ctx context.Context, username string, password string, options types.UserAuthOptions) (userID string, err error) {
		return "", nil
	}

	resp := authInstance.LoginWithUsernameAndPassword(context.Background(), "test@test.com", "password", types.UserAuthOptions{})

	expected := "Invalid credentials"
	if resp.ErrorMessage != expected {
		t.Fatalf("expected error %q, got %q", expected, resp.ErrorMessage)
	}
}

func TestLoginWithUsernameAndPassword_TokenStoreError(t *testing.T) {
	authInstance := newAuthForLoginTests()

	authInstance.funcUserLogin = func(ctx context.Context, username string, password string, options types.UserAuthOptions) (userID string, err error) {
		return "user123", nil
	}

	authInstance.funcUserStoreAuthToken = func(ctx context.Context, token string, userID string, options types.UserAuthOptions) error {
		return errors.New("db error")
	}

	resp := authInstance.LoginWithUsernameAndPassword(context.Background(), "test@test.com", "password", types.UserAuthOptions{})

	expected := "Failed to process request. Please try again later"
	if resp.ErrorMessage != expected {
		t.Fatalf("expected error %q, got %q", expected, resp.ErrorMessage)
	}
}

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

	resp := authInstance.LoginWithUsernameAndPassword(context.Background(), "test@test.com", "password", types.UserAuthOptions{})

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
