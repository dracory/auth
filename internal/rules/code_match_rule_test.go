package rules

import (
	"testing"

	"github.com/dracory/auth/types"
)

func TestCodeMatchRule_MatchPasses(t *testing.T) {
	r := NewCodeMatchRule("a@b.c:123456", "a@b.c:123456")
	if r.Fails() {
		t.Fatal("expected matching values to pass")
	}
}

func TestCodeMatchRule_MismatchFails(t *testing.T) {
	r := NewCodeMatchRule("a@b.c:123456", "a@b.c:654321")
	if !r.Fails() {
		t.Fatal("expected mismatched values to fail")
	}
	if r.FailMessageFirst() != types.MsgVerificationCodeExpired {
		t.Fatalf("expected %q, got %q", types.MsgVerificationCodeExpired, r.FailMessageFirst())
	}
}

func TestCodeMatchRule_EmptyStoredFails(t *testing.T) {
	r := NewCodeMatchRule("", "a@b.c:123456")
	if !r.Fails() {
		t.Fatal("expected empty stored value to fail")
	}
}

func TestCodeMatchRule_DifferentLengthFails(t *testing.T) {
	r := NewCodeMatchRule("short", "much-longer-value")
	if !r.Fails() {
		t.Fatal("expected different-length values to fail")
	}
}
