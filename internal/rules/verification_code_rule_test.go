package rules

import (
	"testing"

	"github.com/dracory/auth/types"
)

func TestVerificationCodeRule_EmptyFails(t *testing.T) {
	r := NewVerificationCodeRule("", false)
	if r.Passes() {
		t.Fatal("expected empty code to fail")
	}
	if r.FailMessageFirst() != types.MsgVerificationCodeRequired {
		t.Fatalf("expected required message, got %q", r.FailMessageFirst())
	}
}

func TestVerificationCodeRule_InvalidLengthFails(t *testing.T) {
	r := NewVerificationCodeRule("ABC", false)
	if r.Passes() {
		t.Fatal("expected short code to fail")
	}
	if r.FailMessageFirst() != types.MsgVerificationCodeInvalidLength {
		t.Fatalf("expected length message, got %q", r.FailMessageFirst())
	}
}

func TestVerificationCodeRule_InvalidCharactersFails(t *testing.T) {
	r := NewVerificationCodeRule("12345678", false)
	if r.Passes() {
		t.Fatal("expected invalid characters to fail")
	}
	if r.FailMessageFirst() != types.MsgVerificationCodeInvalidCharacters {
		t.Fatalf("expected characters message, got %q", r.FailMessageFirst())
	}
}

func TestVerificationCodeRule_ValidPasses(t *testing.T) {
	r := NewVerificationCodeRule("BCDFGHJK", false)
	if r.Fails() {
		t.Fatal("expected valid code to pass")
	}
}

func TestVerificationCodeRule_HardenedGamma(t *testing.T) {
	// Hardened mode (rate limit disabled) uses 12-char codes with a wider
	// alphabet — a normal 8-char code must fail length validation.
	r := NewVerificationCodeRule("BCDFGHJK", true)
	if r.Passes() {
		t.Fatal("expected normal-length code to fail hardened validation")
	}
	if r.FailMessageFirst() != types.MsgVerificationCodeInvalidLength {
		t.Fatalf("expected length message, got %q", r.FailMessageFirst())
	}
}
