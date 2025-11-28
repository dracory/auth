package utils

import "testing"

func TestValidateEmailFormat_EmptyEmail(t *testing.T) {
	if msg := ValidateEmailFormat(""); msg != "" {
		t.Fatalf("expected empty message for empty email, got %q", msg)
	}
}

func TestValidateEmailFormat_ValidEmail(t *testing.T) {
	if msg := ValidateEmailFormat("test@test.com"); msg != "" {
		t.Fatalf("expected empty message for valid email, got %q", msg)
	}
}

func TestValidateEmailFormat_InvalidEmail(t *testing.T) {
	msg := ValidateEmailFormat("invalid-email")

	if msg != "This is not a valid email: invalid-email" {
		t.Fatalf("unexpected validation message: %q", msg)
	}
}
