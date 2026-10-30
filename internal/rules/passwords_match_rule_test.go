package rules

import (
	"testing"

	"github.com/dracory/auth/types"
)

func TestPasswordsMatchRule_MatchPasses(t *testing.T) {
	r := NewPasswordsMatchRule("Str0ng!Pass", "Str0ng!Pass")
	if r.Fails() {
		t.Fatal("expected matching passwords to pass")
	}
}

func TestPasswordsMatchRule_BothEmptyPasses(t *testing.T) {
	r := NewPasswordsMatchRule("", "")
	if r.Fails() {
		t.Fatal("expected two empty passwords to pass (required-field rule covers emptiness)")
	}
}

func TestPasswordsMatchRule_MismatchFails(t *testing.T) {
	r := NewPasswordsMatchRule("Str0ng!Pass", "Str0ng!Pass!")
	if !r.Fails() {
		t.Fatal("expected mismatched passwords to fail")
	}
	if r.FailMessageFirst() != types.MsgPasswordsDoNotMatch {
		t.Fatalf("expected %q, got %q", types.MsgPasswordsDoNotMatch, r.FailMessageFirst())
	}
}
