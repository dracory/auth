package auth

import (
	"strings"
	"testing"
)

func TestEmailTemplatePasswordChange_IncludesURL(t *testing.T) {
	name := "User"
	url := "https://example.com/reset?token=abc123"

	result := emailTemplatePasswordChange(name, url, UserAuthOptions{})

	if result == "" {
		t.Fatalf("expected non-empty template output")
	}

	if !strings.Contains(result, url) {
		t.Fatalf("expected template to contain URL %q, got %q", url, result)
	}
}
