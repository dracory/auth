package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/dracory/auth/types"
)

func newAuthForRegisterTests() *authImplementation {
	return &authImplementation{}
}

func TestRegisterWithUsernameAndPassword_RequiresFirstName(t *testing.T) {
	authInstance := newAuthForRegisterTests()

	resp := authInstance.RegisterWithUsernameAndPassword(context.Background(), "test@test.com", "password", "", "Doe", types.UserAuthOptions{})

	expected := "First name is required field"
	if resp.ErrorMessage != expected {
		t.Fatalf("expected error %q, got %q", expected, resp.ErrorMessage)
	}
}

func TestRegisterWithUsernameAndPassword_RequiresLastName(t *testing.T) {
	authInstance := newAuthForRegisterTests()

	resp := authInstance.RegisterWithUsernameAndPassword(context.Background(), "test@test.com", "password", "John", "", types.UserAuthOptions{})

	expected := "Last name is required field"
	if resp.ErrorMessage != expected {
		t.Fatalf("expected error %q, got %q", expected, resp.ErrorMessage)
	}
}

func TestRegisterWithUsernameAndPassword_RequiresEmail(t *testing.T) {
	authInstance := newAuthForRegisterTests()

	resp := authInstance.RegisterWithUsernameAndPassword(context.Background(), "", "password", "John", "Doe", types.UserAuthOptions{})

	expected := "Email is required field"
	if resp.ErrorMessage != expected {
		t.Fatalf("expected error %q, got %q", expected, resp.ErrorMessage)
	}
}

func TestRegisterWithUsernameAndPassword_RequiresPassword(t *testing.T) {
	authInstance := newAuthForRegisterTests()

	resp := authInstance.RegisterWithUsernameAndPassword(context.Background(), "test@test.com", "", "John", "Doe", types.UserAuthOptions{})

	expected := "Password is required field"
	if resp.ErrorMessage != expected {
		t.Fatalf("expected error %q, got %q", expected, resp.ErrorMessage)
	}
}

func TestRegisterWithUsernameAndPassword_InvalidEmail(t *testing.T) {
	authInstance := newAuthForRegisterTests()

	resp := authInstance.RegisterWithUsernameAndPassword(context.Background(), "invalid-email", "password", "John", "Doe", types.UserAuthOptions{})

	expected := "This is not a valid email: invalid-email"
	if resp.ErrorMessage != expected {
		t.Fatalf("expected error %q, got %q", expected, resp.ErrorMessage)
	}
}

func TestRegisterWithUsernameAndPassword_FuncUserRegisterNotDefined(t *testing.T) {
	authInstance := newAuthForRegisterTests()

	resp := authInstance.RegisterWithUsernameAndPassword(context.Background(), "test@test.com", "password", "John", "Doe", types.UserAuthOptions{})

	expected := "registration failed. FuncUserRegister function not defined"
	if resp.ErrorMessage != expected {
		t.Fatalf("expected error %q, got %q", expected, resp.ErrorMessage)
	}
}

func TestRegisterWithUsernameAndPassword_RegistrationFailed_NoVerification(t *testing.T) {
	authInstance := newAuthForRegisterTests()

	// enableVerification is false by default
	authInstance.funcUserRegister = func(ctx context.Context, username string, password string, firstName string, lastName string, options types.UserAuthOptions) error {
		return errors.New("db error")
	}

	resp := authInstance.RegisterWithUsernameAndPassword(context.Background(), "test@test.com", "password", "John", "Doe", types.UserAuthOptions{})

	expected := "registration failed."
	if resp.ErrorMessage != expected {
		t.Fatalf("expected error %q, got %q", expected, resp.ErrorMessage)
	}
}

func TestRegisterWithUsernameAndPassword_RegistrationSuccess_NoVerification(t *testing.T) {
	authInstance := newAuthForRegisterTests()

	authInstance.funcUserRegister = func(ctx context.Context, username string, password string, firstName string, lastName string, options types.UserAuthOptions) error {
		return nil
	}

	resp := authInstance.RegisterWithUsernameAndPassword(context.Background(), "test@test.com", "password", "John", "Doe", types.UserAuthOptions{})

	if resp.ErrorMessage != "" {
		t.Fatalf("expected no error, got %q", resp.ErrorMessage)
	}

	expectedSuccess := "registration success"
	if resp.SuccessMessage != expectedSuccess {
		t.Fatalf("expected success message %q, got %q", expectedSuccess, resp.SuccessMessage)
	}
}

func TestRegisterWithUsernameAndPassword_VerificationEnabled_TokenStoreError(t *testing.T) {
	authInstance := newAuthForRegisterTests()

	authInstance.enableVerification = true
	// funcUserRegister must be defined even if not used, otherwise earlier check fails
	authInstance.funcUserRegister = func(ctx context.Context, username string, password string, firstName string, lastName string, options types.UserAuthOptions) error {
		return nil
	}

	// Force token store error
	authInstance.funcTemporaryKeySet = func(key string, value string, expiresSeconds int) error {
		return errors.New("db error")
	}

	// Provide minimal email template and send functions
	authInstance.funcEmailTemplateRegisterCode = func(ctx context.Context, email string, code string, options types.UserAuthOptions) string {
		return "body"
	}
	// Email send not reached in this test
	authInstance.funcEmailSend = func(ctx context.Context, userID string, emailSubject string, emailBody string) error {
		return nil
	}

	resp := authInstance.RegisterWithUsernameAndPassword(context.Background(), "test@test.com", "password", "John", "Doe", types.UserAuthOptions{})

	expected := "Failed to process request. Please try again later"
	if resp.ErrorMessage != expected {
		t.Fatalf("expected error %q, got %q", expected, resp.ErrorMessage)
	}
}

func TestRegisterWithUsernameAndPassword_VerificationEnabled_EmailSendError(t *testing.T) {
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
		return errors.New("smtp error")
	}

	resp := authInstance.RegisterWithUsernameAndPassword(context.Background(), "test@test.com", "password", "John", "Doe", types.UserAuthOptions{})

	expected := "Failed to send email. Please try again later"
	if resp.ErrorMessage != expected {
		t.Fatalf("expected error %q, got %q", expected, resp.ErrorMessage)
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

	resp := authInstance.RegisterWithUsernameAndPassword(context.Background(), "test@test.com", "password", "John", "Doe", types.UserAuthOptions{})

	if resp.ErrorMessage != "" {
		t.Fatalf("expected no error, got %q", resp.ErrorMessage)
	}

	expectedSuccess := "Registration code was sent successfully"
	if resp.SuccessMessage != expectedSuccess {
		t.Fatalf("expected success message %q, got %q", expectedSuccess, resp.SuccessMessage)
	}
}
