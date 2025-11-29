package emails

import (
	"strings"
	"testing"
)

func TestEmailRegisterCodeTemplate_IncludesEmailAndCode(t *testing.T) {
	email := "user@example.com"
	code := "REG12345"

	result := EmailRegisterCodeTemplate(email, code)

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
