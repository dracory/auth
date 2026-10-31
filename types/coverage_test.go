package types

import (
	"errors"
	"testing"
)

func TestWithCookieName(t *testing.T) {
	c := CookieConfig{}
	WithCookieName("session_token")(&c)
	if c.Name != "session_token" {
		t.Fatalf("expected cookie name set, got %q", c.Name)
	}
}

func TestNoopObservabilityHooks_AllMethodsCallable(t *testing.T) {
	var h ObservabilityHooks = NoopObservabilityHooks{}
	h.RecordLoginAttempt("a@b.com", true, nil)
	h.RecordRegistrationAttempt(false, errors.New("x"))
	h.RecordRateLimitHit("1.2.3.4", "login")
	h.RecordSessionCreated("u1")
	h.RecordPasswordReset(true, nil)
	h.RecordImpersonationStart("admin", "u1")
	h.RecordImpersonationStop("admin", "u1")
}

type fakeAuth struct{ AuthSharedInterface }

type fakeWrapper struct {
	AuthSharedInterface
	inner AuthSharedInterface
}

func (w fakeWrapper) Unwrap() AuthSharedInterface { return w.inner }

func TestUnwrapAuth(t *testing.T) {
	base := fakeAuth{}
	if got := UnwrapAuth(base); got != base {
		t.Fatal("expected unwrapped impl returned as-is")
	}

	inner := fakeAuth{}
	wrapped := fakeWrapper{inner: inner}
	if got := UnwrapAuth(wrapped); got != inner {
		t.Fatal("expected decorator unwrapped to inner impl")
	}
}
