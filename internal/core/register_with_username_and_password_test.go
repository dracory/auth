package core_test

import (
	"context"
	"testing"
	"time"

	"github.com/dracory/auth/internal/core"
	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/types"
)

func newPasswordAuthForRegisterTest(t *testing.T) types.AuthPasswordInterface {
	t.Helper()

	shared := testutils.NewAuthSharedForTest()
	passwordAuth, ok := shared.(types.AuthPasswordInterface)
	if !ok {
		t.Fatalf("test auth does not implement AuthPasswordInterface")
	}
	return passwordAuth
}

func TestCoreRegisterWithUsernameAndPassword_NoVerification_Success(t *testing.T) {
	a := newPasswordAuthForRegisterTest(t)
	a.SetPasswordStrength(&types.PasswordStrengthConfig{MinLength: 4})

	called := false
	a.SetFuncUserRegister(func(ctx context.Context, email, password, firstName, lastName string, options types.UserAuthOptions) error {
		called = true
		return nil
	})

	resp := core.RegisterWithUsernameAndPassword(context.Background(), "test@test.com", "pass", "John", "Doe", types.UserAuthOptions{}, a, time.Hour)

	if resp.ErrorMessage != "" {
		t.Fatalf("expected no error, got %q", resp.ErrorMessage)
	}
	if resp.SuccessMessage != "registration success" {
		t.Fatalf("expected success message %q, got %q", "registration success", resp.SuccessMessage)
	}
	if !called {
		t.Fatalf("expected FuncUserRegister to be called")
	}
}

func TestCoreRegisterWithUsernameAndPassword_VerificationEnabled_Success(t *testing.T) {
	a := newPasswordAuthForRegisterTest(t)
	a.SetPasswordStrength(&types.PasswordStrengthConfig{MinLength: 4})

	// Enable verification
	SetVerificationForTest(a, true)

	a.SetFuncUserRegister(func(ctx context.Context, email, password, firstName, lastName string, options types.UserAuthOptions) error {
		return nil
	})

	var storedKey string
	var storedValue string
	var storedExpires int
	a.SetFuncTemporaryKeySet(func(key string, value string, expiresSeconds int) error {
		storedKey = key
		storedValue = value
		storedExpires = expiresSeconds
		return nil
	})

	calledEmail := false
	a.SetFuncEmailTemplateRegisterCode(func(ctx context.Context, email string, code string, options types.UserAuthOptions) string {
		calledEmail = true
		if code == "" {
			t.Fatalf("expected non-empty verification code in template")
		}
		return "body"
	})

	calledSend := false
	a.SetFuncEmailSend(func(ctx context.Context, userID string, subject string, body string) error {
		calledSend = true
		return nil
	})

	resp := core.RegisterWithUsernameAndPassword(context.Background(), "test@test.com", "pass", "John", "Doe", types.UserAuthOptions{}, a, time.Hour)

	if resp.ErrorMessage != "" {
		t.Fatalf("expected no error, got %q", resp.ErrorMessage)
	}
	if resp.SuccessMessage != "Registration code was sent successfully" {
		t.Fatalf("expected success message for verification, got %q", resp.SuccessMessage)
	}
	if storedKey == "" || storedValue == "" || storedExpires == 0 {
		t.Fatalf("expected temporary key to be stored with value and expiration")
	}
	if !calledEmail {
		t.Fatalf("expected FuncEmailTemplateRegisterCode to be called")
	}
	if !calledSend {
		t.Fatalf("expected FuncEmailSend to be called")
	}
}

// SetVerificationForTest toggles verification flag on the underlying authSharedTest
// used by the internal testutils helper. This relies on the concrete type used there.
func SetVerificationForTest(a types.AuthSharedInterface, verification bool) {
	testutils.SetVerificationForTest(a, verification)
}
