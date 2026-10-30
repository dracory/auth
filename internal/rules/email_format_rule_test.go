package rules

import (
	"strings"
	"testing"
)

func TestEmailFormatRule_ValidEmailPasses(t *testing.T) {
	r := NewEmailFormatRule("user@example.com")
	if r.Fails() {
		t.Fatal("expected valid email to pass")
	}
}

func TestEmailFormatRule_EmptyEmailPasses(t *testing.T) {
	r := NewEmailFormatRule("")
	if r.Fails() {
		t.Fatal("expected empty email to pass (format rule defers presence check to required rule)")
	}
}

func TestEmailFormatRule_InvalidEmailFails(t *testing.T) {
	r := NewEmailFormatRule("not-an-email")
	if r.Passes() {
		t.Fatal("expected invalid email to fail")
	}
	if !strings.Contains(r.FailMessageFirst(), "not a valid email") {
		t.Fatalf("expected fail message containing 'not a valid email', got %q", r.FailMessageFirst())
	}
}
