package rules

import (
	"testing"

	"github.com/dracory/auth/types"
)

func TestNonceRule_EmptyFails(t *testing.T) {
	r := NewNonceRule("")
	if r.Passes() {
		t.Fatal("expected empty nonce to fail")
	}
	if r.FailMessageFirst() != types.MsgNonceRequired {
		t.Fatalf("expected required message, got %q", r.FailMessageFirst())
	}
}

func TestNonceRule_InvalidFormatFails(t *testing.T) {
	r := NewNonceRule("not-hex")
	if r.Passes() {
		t.Fatal("expected malformed nonce to fail")
	}
	if r.FailMessageFirst() != types.MsgNonceInvalid {
		t.Fatalf("expected invalid message, got %q", r.FailMessageFirst())
	}
}

func TestNonceRule_ShortHexFails(t *testing.T) {
	r := NewNonceRule("abcdef0123")
	if r.Passes() {
		t.Fatal("expected short hex to fail")
	}
}

func TestNonceRule_ValidPasses(t *testing.T) {
	r := NewNonceRule("0123456789abcdef0123456789abcdef")
	if r.Fails() {
		t.Fatal("expected valid nonce to pass")
	}
}
