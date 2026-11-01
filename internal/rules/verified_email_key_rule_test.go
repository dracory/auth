package rules

import (
	"errors"
	"testing"
)

func TestVerifiedEmailKeyRule_EmptyKey(t *testing.T) {
	r := NewVerifiedEmailKeyRule(func(k string) (string, error) { return "a@b.com", nil }, "")
	if !r.Fails() {
		t.Fatal("expected fail for empty key")
	}
	if r.Email() != "" {
		t.Fatal("expected no email")
	}
}

func TestVerifiedEmailKeyRule_NilStore(t *testing.T) {
	r := NewVerifiedEmailKeyRule(nil, "tok")
	if !r.Fails() {
		t.Fatal("expected fail with nil store")
	}
}

func TestVerifiedEmailKeyRule_Unresolvable(t *testing.T) {
	r := NewVerifiedEmailKeyRule(func(k string) (string, error) { return "", errors.New("expired") }, "tok")
	if !r.Fails() {
		t.Fatal("expected fail for unresolvable key")
	}
	if r.FailMessageFirst() == "" {
		t.Fatal("expected expired-link fail message")
	}
}

func TestVerifiedEmailKeyRule_Resolves(t *testing.T) {
	r := NewVerifiedEmailKeyRule(func(k string) (string, error) {
		if k == "tok-1" {
			return "v@example.com", nil
		}
		return "", errors.New("miss")
	}, "tok-1")
	if r.Fails() {
		t.Fatal("expected pass for resolvable key")
	}
	if r.Email() != "v@example.com" {
		t.Fatalf("expected resolved email, got %q", r.Email())
	}
}
