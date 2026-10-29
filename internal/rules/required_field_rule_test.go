package rules

import (
	"testing"

	"github.com/dracory/auth/types"
)

func TestRequiredFieldRule_EmptyFails(t *testing.T) {
	r := NewRequiredFieldRule("", types.MsgEmailRequired)
	if r.Passes() {
		t.Fatal("expected empty field to fail")
	}
	if r.FailMessageFirst() != types.MsgEmailRequired {
		t.Fatalf("expected supplied message, got %q", r.FailMessageFirst())
	}
}

func TestRequiredFieldRule_NonEmptyPasses(t *testing.T) {
	r := NewRequiredFieldRule("x", types.MsgEmailRequired)
	if r.Fails() {
		t.Fatal("expected non-empty field to pass")
	}
}
