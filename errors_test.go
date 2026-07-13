package auth

import (
	"errors"
	"testing"
)

func TestAuthError_Error(t *testing.T) {
	e := AuthError{Message: "test error message"}
	if e.Error() != "test error message" {
		t.Fatalf("expected %q, got %q", "test error message", e.Error())
	}
}

func TestAuthError_EmptyMessage(t *testing.T) {
	e := AuthError{}
	if e.Error() != "" {
		t.Fatalf("expected empty string, got %q", e.Error())
	}
}

func TestNewEmailSendError(t *testing.T) {
	internalErr := errors.New("smtp connection failed")
	e := NewEmailSendError(internalErr)
	if e.Code != ErrCodeEmailSendFailed {
		t.Fatalf("expected code %q, got %q", ErrCodeEmailSendFailed, e.Code)
	}
	if e.Message == "" {
		t.Fatal("expected non-empty message")
	}
	if e.InternalErr != internalErr {
		t.Fatalf("expected InternalErr to match input error")
	}
	if e.Error() != e.Message {
		t.Fatalf("Error() should return Message, got %q vs %q", e.Error(), e.Message)
	}
}

func TestNewTokenStoreError(t *testing.T) {
	internalErr := errors.New("db write failed")
	e := NewTokenStoreError(internalErr)
	if e.Code != ErrCodeTokenStoreFailed {
		t.Fatalf("expected code %q, got %q", ErrCodeTokenStoreFailed, e.Code)
	}
	if e.Message == "" {
		t.Fatal("expected non-empty message")
	}
	if e.InternalErr != internalErr {
		t.Fatalf("expected InternalErr to match input error")
	}
}

func TestNewCodeGenerationError(t *testing.T) {
	internalErr := errors.New("random source exhausted")
	e := NewCodeGenerationError(internalErr)
	if e.Code != ErrCodeCodeGenerationFailed {
		t.Fatalf("expected code %q, got %q", ErrCodeCodeGenerationFailed, e.Code)
	}
	if e.Message == "" {
		t.Fatal("expected non-empty message")
	}
	if e.InternalErr != internalErr {
		t.Fatalf("expected InternalErr to match input error")
	}
}

func TestNewSerializationError(t *testing.T) {
	internalErr := errors.New("json marshal failed")
	e := NewSerializationError(internalErr)
	if e.Code != ErrCodeSerializationFailed {
		t.Fatalf("expected code %q, got %q", ErrCodeSerializationFailed, e.Code)
	}
	if e.Message == "" {
		t.Fatal("expected non-empty message")
	}
	if e.InternalErr != internalErr {
		t.Fatalf("expected InternalErr to match input error")
	}
}

func TestNewAuthenticationError(t *testing.T) {
	internalErr := errors.New("invalid credentials")
	e := NewAuthenticationError(internalErr)
	if e.Code != ErrCodeAuthenticationFailed {
		t.Fatalf("expected code %q, got %q", ErrCodeAuthenticationFailed, e.Code)
	}
	if e.Message == "" {
		t.Fatal("expected non-empty message")
	}
	if e.InternalErr != internalErr {
		t.Fatalf("expected InternalErr to match input error")
	}
}

func TestNewRegistrationError(t *testing.T) {
	internalErr := errors.New("user already exists")
	e := NewRegistrationError(internalErr)
	if e.Code != ErrCodeRegistrationFailed {
		t.Fatalf("expected code %q, got %q", ErrCodeRegistrationFailed, e.Code)
	}
	if e.Message == "" {
		t.Fatal("expected non-empty message")
	}
	if e.InternalErr != internalErr {
		t.Fatalf("expected InternalErr to match input error")
	}
}

func TestNewLogoutError(t *testing.T) {
	internalErr := errors.New("token deletion failed")
	e := NewLogoutError(internalErr)
	if e.Code != ErrCodeLogoutFailed {
		t.Fatalf("expected code %q, got %q", ErrCodeLogoutFailed, e.Code)
	}
	if e.Message == "" {
		t.Fatal("expected non-empty message")
	}
	if e.InternalErr != internalErr {
		t.Fatalf("expected InternalErr to match input error")
	}
}

func TestNewPasswordResetError(t *testing.T) {
	internalErr := errors.New("password hash failed")
	e := NewPasswordResetError(internalErr)
	if e.Code != ErrCodePasswordResetFailed {
		t.Fatalf("expected code %q, got %q", ErrCodePasswordResetFailed, e.Code)
	}
	if e.Message == "" {
		t.Fatal("expected non-empty message")
	}
	if e.InternalErr != internalErr {
		t.Fatalf("expected InternalErr to match input error")
	}
}

func TestNewInternalError(t *testing.T) {
	internalErr := errors.New("unexpected panic")
	e := NewInternalError(internalErr)
	if e.Code != ErrCodeInternalError {
		t.Fatalf("expected code %q, got %q", ErrCodeInternalError, e.Code)
	}
	if e.Message == "" {
		t.Fatal("expected non-empty message")
	}
	if e.InternalErr != internalErr {
		t.Fatalf("expected InternalErr to match input error")
	}
}

func TestNewErrors_WithNilInternalErr(t *testing.T) {
	tests := []struct {
		name string
		fn   func(error) AuthError
		code string
	}{
		{"EmailSend", NewEmailSendError, ErrCodeEmailSendFailed},
		{"TokenStore", NewTokenStoreError, ErrCodeTokenStoreFailed},
		{"CodeGeneration", NewCodeGenerationError, ErrCodeCodeGenerationFailed},
		{"Serialization", NewSerializationError, ErrCodeSerializationFailed},
		{"Authentication", NewAuthenticationError, ErrCodeAuthenticationFailed},
		{"Registration", NewRegistrationError, ErrCodeRegistrationFailed},
		{"Logout", NewLogoutError, ErrCodeLogoutFailed},
		{"PasswordReset", NewPasswordResetError, ErrCodePasswordResetFailed},
		{"Internal", NewInternalError, ErrCodeInternalError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.fn(nil)
			if e.Code != tt.code {
				t.Fatalf("expected code %q, got %q", tt.code, e.Code)
			}
			if e.Message == "" {
				t.Fatal("expected non-empty message")
			}
			if e.InternalErr != nil {
				t.Fatal("expected nil InternalErr")
			}
		})
	}
}
