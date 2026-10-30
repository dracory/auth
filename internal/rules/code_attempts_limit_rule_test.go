package rules

import (
	"errors"
	"fmt"
	"testing"

	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
)

func newAttemptsStore() map[string]string {
	return map[string]string{}
}

func attemptsGet(store map[string]string) func(string) (string, error) {
	return func(k string) (string, error) {
		v, ok := store[k]
		if !ok {
			return "", errors.New("not found")
		}
		return v, nil
	}
}

func attemptsSet(store map[string]string) func(string, string, int) error {
	return func(k, v string, _ int) error {
		store[k] = v
		return nil
	}
}

func attemptsDelete(store map[string]string) func(string) error {
	return func(k string) error {
		delete(store, k)
		return nil
	}
}

func TestCodeAttemptsLimitRule_UnderLimitPasses(t *testing.T) {
	store := newAttemptsStore()
	store["attempts"] = fmt.Sprintf("%d", utils.MaxCodeAttempts-1)
	store["payload"] = "email:code"

	r := NewCodeAttemptsLimitRule(attemptsGet(store), attemptsDelete(store), attemptsSet(store), "payload", "attempts")
	if r.Fails() {
		t.Fatal("expected attempts below limit to pass")
	}
	if r.Attempts() != utils.MaxCodeAttempts-1 {
		t.Fatalf("expected Attempts()=%d, got %d", utils.MaxCodeAttempts-1, r.Attempts())
	}
	if _, ok := store["payload"]; !ok {
		t.Fatal("expected payload key to be preserved under the limit")
	}
}

func TestCodeAttemptsLimitRule_AtLimitFailsAndInvalidates(t *testing.T) {
	store := newAttemptsStore()
	store["attempts"] = fmt.Sprintf("%d", utils.MaxCodeAttempts)
	store["payload"] = "email:code"

	r := NewCodeAttemptsLimitRule(attemptsGet(store), attemptsDelete(store), attemptsSet(store), "payload", "attempts")
	if !r.Fails() {
		t.Fatal("expected attempts at limit to fail")
	}
	if r.FailMessageFirst() != types.MsgTooManyAttempts {
		t.Fatalf("expected %q, got %q", types.MsgTooManyAttempts, r.FailMessageFirst())
	}
	if _, ok := store["payload"]; ok {
		t.Fatal("expected payload key to be invalidated at the limit")
	}
	if _, ok := store["attempts"]; ok {
		t.Fatal("expected attempts key to be invalidated at the limit")
	}
}

func TestCodeAttemptsLimitRule_MissingCounterFailsOpen(t *testing.T) {
	store := newAttemptsStore()

	r := NewCodeAttemptsLimitRule(attemptsGet(store), nil, attemptsSet(store), "payload", "attempts")
	if r.Fails() {
		t.Fatal("expected missing counter to pass (fail-open)")
	}
	if r.Attempts() != 0 {
		t.Fatalf("expected Attempts()=0, got %d", r.Attempts())
	}
}

func TestCodeAttemptsLimitRule_CorruptCounterFailsOpen(t *testing.T) {
	store := newAttemptsStore()
	store["attempts"] = "not-a-number"

	r := NewCodeAttemptsLimitRule(attemptsGet(store), nil, attemptsSet(store), "payload", "attempts")
	if r.Fails() {
		t.Fatal("expected corrupt counter to pass (fail-open)")
	}
	if r.Attempts() != 0 {
		t.Fatalf("expected Attempts()=0, got %d", r.Attempts())
	}
}

func TestCodeAttemptsLimitRule_NilGetPasses(t *testing.T) {
	r := NewCodeAttemptsLimitRule(nil, nil, nil, "payload", "attempts")
	if r.Fails() {
		t.Fatal("expected nil store to pass")
	}
}

func TestCodeAttemptsLimitRule_InvalidatesViaSetFallback(t *testing.T) {
	store := newAttemptsStore()
	store["attempts"] = fmt.Sprintf("%d", utils.MaxCodeAttempts+3)
	store["payload"] = "email:code"

	r := NewCodeAttemptsLimitRule(attemptsGet(store), nil, attemptsSet(store), "payload", "attempts")
	if !r.Fails() {
		t.Fatal("expected attempts above limit to fail")
	}
	// Fallback convention: "empty value, 0 expiry"
	if v, ok := store["payload"]; !ok || v != "" {
		t.Fatal("expected payload key to be emptied via set fallback")
	}
}
