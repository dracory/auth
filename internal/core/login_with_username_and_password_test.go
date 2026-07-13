package core_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dracory/auth/internal/core"
	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/types"
)

func newPasswordAuthForLoginTest(t *testing.T) types.AuthPasswordInterface {
	t.Helper()

	shared := testutils.NewAuthSharedForTest()
	passwordAuth, ok := shared.(types.AuthPasswordInterface)
	if !ok {
		t.Fatalf("test auth does not implement AuthPasswordInterface")
	}
	return passwordAuth
}

func TestCoreLoginWithUsernameAndPassword_ValidationErrors(t *testing.T) {
	a := newPasswordAuthForLoginTest(t)

	resp := core.LoginWithUsernameAndPassword(context.Background(), a, "", "password", types.UserAuthOptions{})
	if resp.ErrorMessage != "Email is required field" {
		t.Fatalf("expected error %q, got %q", "Email is required field", resp.ErrorMessage)
	}

	resp = core.LoginWithUsernameAndPassword(context.Background(), a, "test@test.com", "", types.UserAuthOptions{})
	if resp.ErrorMessage != "Password is required field" {
		t.Fatalf("expected error %q, got %q", "Password is required field", resp.ErrorMessage)
	}

	resp = core.LoginWithUsernameAndPassword(context.Background(), a, "invalid-email", "password", types.UserAuthOptions{})
	if resp.ErrorMessage == "" {
		t.Fatalf("expected validation error for invalid email, got empty message")
	}
}

func TestCoreLoginWithUsernameAndPassword_Success(t *testing.T) {
	a := newPasswordAuthForLoginTest(t)

	a.SetFuncUserLogin(func(ctx context.Context, email, password string, options types.UserAuthOptions) (string, error) {
		return "user123", nil
	})

	var storedToken string
	var storedUserID string
	a.SetFuncUserStoreAuthToken(func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
		storedToken = token
		storedUserID = userID
		return nil
	})

	resp := core.LoginWithUsernameAndPassword(context.Background(), a, "test@test.com", "password", types.UserAuthOptions{})

	if resp.ErrorMessage != "" {
		t.Fatalf("expected no error, got %q", resp.ErrorMessage)
	}
	if resp.SuccessMessage != "login success" {
		t.Fatalf("expected success message %q, got %q", "login success", resp.SuccessMessage)
	}
	if resp.Token == "" {
		t.Fatalf("expected non-empty token in response")
	}
	if storedToken == "" || storedToken != resp.Token {
		t.Fatalf("expected stored token to match response token, got %q and %q", storedToken, resp.Token)
	}
	if storedUserID != "user123" {
		t.Fatalf("expected stored userID 'user123', got %q", storedUserID)
	}
}

func TestCoreLoginWithUsernameAndPassword_LoginError(t *testing.T) {
	a := newPasswordAuthForLoginTest(t)

	a.SetFuncUserLogin(func(ctx context.Context, email, password string, options types.UserAuthOptions) (string, error) {
		return "", errors.New("invalid credentials")
	})

	resp := core.LoginWithUsernameAndPassword(context.Background(), a, "test@test.com", "password", types.UserAuthOptions{})
	if resp.ErrorMessage != "Invalid credentials" {
		t.Fatalf("expected 'Invalid credentials', got %q", resp.ErrorMessage)
	}
}

func TestCoreLoginWithUsernameAndPassword_EmptyUserID(t *testing.T) {
	a := newPasswordAuthForLoginTest(t)

	a.SetFuncUserLogin(func(ctx context.Context, email, password string, options types.UserAuthOptions) (string, error) {
		return "", nil
	})

	resp := core.LoginWithUsernameAndPassword(context.Background(), a, "test@test.com", "password", types.UserAuthOptions{})
	if resp.ErrorMessage != "Invalid credentials" {
		t.Fatalf("expected 'Invalid credentials', got %q", resp.ErrorMessage)
	}
}

func TestCoreLoginWithUsernameAndPassword_TokenStoreError(t *testing.T) {
	a := newPasswordAuthForLoginTest(t)

	a.SetFuncUserLogin(func(ctx context.Context, email, password string, options types.UserAuthOptions) (string, error) {
		return "user123", nil
	})

	a.SetFuncUserStoreAuthToken(func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
		return errors.New("db error")
	})

	resp := core.LoginWithUsernameAndPassword(context.Background(), a, "test@test.com", "password", types.UserAuthOptions{})
	if resp.ErrorMessage != "Failed to process request. Please try again later" {
		t.Fatalf("expected token store error, got %q", resp.ErrorMessage)
	}
}
