package auth

import (
	"strings"
	"testing"
)

func TestEmailLoginCodeTemplate_IncludesEmailAndCode(t *testing.T) {
	email := "user@example.com"
	code := "ABC12345"

	result := emailLoginCodeTemplate(email, code, UserAuthOptions{})

	if result == "" {
		t.Fatalf("expected non-empty template output")
	}

	if !strings.Contains(result, email) {
		t.Fatalf("expected template to contain email %q, got %q", email, result)
	}

	if !strings.Contains(result, code) {
		t.Fatalf("expected template to contain code %q, got %q", code, result)
	}
}
