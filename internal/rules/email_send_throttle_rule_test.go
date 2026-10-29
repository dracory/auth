package rules

import (
	"errors"
	"testing"

	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
)

func TestEmailSendThrottleRule_NilGetPasses(t *testing.T) {
	r := NewEmailSendThrottleRule(nil, "user@example.com")
	if r.Fails() {
		t.Fatal("expected nil getter to pass (unconfigured store)")
	}
}

func TestEmailSendThrottleRule_UnderQuotaPasses(t *testing.T) {
	get := func(key string) (string, error) {
		return "", errors.New("not found")
	}
	r := NewEmailSendThrottleRule(get, "user@example.com")
	if r.Fails() {
		t.Fatal("expected under-quota email to pass")
	}
}

func TestEmailSendThrottleRule_AtQuotaFails(t *testing.T) {
	get := func(key string) (string, error) {
		return "3", nil
	}
	r := NewEmailSendThrottleRule(get, "user@example.com")
	if r.Passes() {
		t.Fatal("expected at-quota email to fail")
	}
	if r.FailMessageFirst() != types.MsgTooManySends {
		t.Fatalf("expected too-many-sends message, got %q", r.FailMessageFirst())
	}
}

func TestEmailSendThrottleRule_HashesEmail(t *testing.T) {
	var lookedUpKey string
	get := func(key string) (string, error) {
		lookedUpKey = key
		return "", errors.New("not found")
	}
	r := NewEmailSendThrottleRule(get, "User@Example.com")
	r.Evaluate()
	if lookedUpKey == "" {
		t.Fatal("expected store lookup")
	}
	// Lookup key is the SHA-256 of the lowercased email — no PII in the store.
	expected := utils.EmailSendsKey("user@example.com")
	if lookedUpKey != expected {
		t.Fatalf("expected hashed key %q, got %q", expected, lookedUpKey)
	}
}
