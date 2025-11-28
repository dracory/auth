package auth

import (
	"context"
	"errors"
	"testing"
)

func newAuthForLoginTests() *Auth {
	return &Auth{}
}

func TestLoginWithUsernameAndPassword_RequiresEmail(t *testing.T) {
	authInstance := newAuthForLoginTests()

	resp := authInstance.LoginWithUsernameAndPassword(context.Background(), "", "password", UserAuthOptions{})

	if resp.ErrorMessage != "Email is required field" {
		t.Fatalf("expected error %q, got %q", "Email is required field", resp.ErrorMessage)
	}
}

func TestLoginWithUsernameAndPassword_RequiresPassword(t *testing.T) {
	authInstance := newAuthForLoginTests()

	resp := authInstance.LoginWithUsernameAndPassword(context.Background(), "test@test.com", "", UserAuthOptions{})

	if resp.ErrorMessage != "Password is required field" {
		t.Fatalf("expected error %q, got %q", "Password is required field", resp.ErrorMessage)
	}
}

func TestLoginWithUsernameAndPassword_InvalidEmail(t *testing.T) {
	authInstance := newAuthForLoginTests()

	resp := authInstance.LoginWithUsernameAndPassword(context.Background(), "invalid-email", "password", UserAuthOptions{})

	expected := "This is not a valid email: invalid-email"
	if resp.ErrorMessage != expected {
		t.Fatalf("expected error %q, got %q", expected, resp.ErrorMessage)
	}
}

func TestLoginWithUsernameAndPassword_LoginError(t *testing.T) {
	authInstance := newAuthForLoginTests()

	authInstance.funcUserLogin = func(ctx context.Context, username string, password string, options UserAuthOptions) (userID string, err error) {
		return "", errors.New("db error")
	}

	resp := authInstance.LoginWithUsernameAndPassword(context.Background(), "test@test.com", "password", UserAuthOptions{})

	expected := "Invalid credentials"
	if resp.ErrorMessage != expected {
		t.Fatalf("expected error %q, got %q", expected, resp.ErrorMessage)
	}
}

func TestLoginWithUsernameAndPassword_UserNotFound(t *testing.T) {
	authInstance := newAuthForLoginTests()

	authInstance.funcUserLogin = func(ctx context.Context, username string, password string, options UserAuthOptions) (userID string, err error) {
		return "", nil
	}

	resp := authInstance.LoginWithUsernameAndPassword(context.Background(), "test@test.com", "password", UserAuthOptions{})

	expected := "Invalid credentials"
	if resp.ErrorMessage != expected {
		t.Fatalf("expected error %q, got %q", expected, resp.ErrorMessage)
	}
}

func TestLoginWithUsernameAndPassword_TokenStoreError(t *testing.T) {
	authInstance := newAuthForLoginTests()

	authInstance.funcUserLogin = func(ctx context.Context, username string, password string, options UserAuthOptions) (userID string, err error) {
		return "user123", nil
	}

	authInstance.funcUserStoreAuthToken = func(ctx context.Context, token string, userID string, options UserAuthOptions) error {
		return errors.New("db error")
	}

	resp := authInstance.LoginWithUsernameAndPassword(context.Background(), "test@test.com", "password", UserAuthOptions{})

	expected := "token store failed."
	if resp.ErrorMessage != expected {
		t.Fatalf("expected error %q, got %q", expected, resp.ErrorMessage)
	}
}

func TestLoginWithUsernameAndPassword_Success(t *testing.T) {
	authInstance := newAuthForLoginTests()

	authInstance.funcUserLogin = func(ctx context.Context, username string, password string, options UserAuthOptions) (userID string, err error) {
		return "user123", nil
	}

	// Ensure token store succeeds
	authInstance.funcUserStoreAuthToken = func(ctx context.Context, token string, userID string, options UserAuthOptions) error {
		if token == "" {
			t.Fatalf("expected non-empty token in FuncUserStoreAuthToken")
		}
		return nil
	}

	resp := authInstance.LoginWithUsernameAndPassword(context.Background(), "test@test.com", "password", UserAuthOptions{})

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
