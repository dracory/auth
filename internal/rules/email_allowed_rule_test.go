package rules

import (
	"testing"

	"github.com/dracory/auth/types"
)

func TestEmailAllowedRule_NilAllowlistPasses(t *testing.T) {
	r := NewEmailAllowedRule(nil, "any@example.com")
	if r.Fails() {
		t.Fatal("expected nil allowlist to pass")
	}
}

func TestEmailAllowedRule_AllowedPasses(t *testing.T) {
	r := NewEmailAllowedRule(func(email string) bool { return email == "ok@example.com" }, "ok@example.com")
	if r.Fails() {
		t.Fatal("expected allowlisted email to pass")
	}
}

func TestEmailAllowedRule_DeniedFails(t *testing.T) {
	r := NewEmailAllowedRule(func(email string) bool { return false }, "no@example.com")
	if !r.Fails() {
		t.Fatal("expected non-allowlisted email to fail")
	}
	if r.FailMessageFirst() != types.MsgInvalidCredentials {
		t.Fatalf("expected %q, got %q", types.MsgInvalidCredentials, r.FailMessageFirst())
	}
}
