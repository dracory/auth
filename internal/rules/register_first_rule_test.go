package rules

import (
	"context"
	"errors"
	"testing"

	"github.com/dracory/auth/types"
)

func TestRegisterFirstRule_Disabled(t *testing.T) {
	r := NewRegisterFirstRule(false, nil, nil, context.Background(), "a@b.com", types.UserAuthOptions{})
	if !r.Fails() {
		t.Fatal("expected rule to fail when policy disabled")
	}
	if r.Degraded() {
		t.Fatal("disabled policy must not degrade — silent non-application")
	}
}

func TestRegisterFirstRule_MissingUserExists(t *testing.T) {
	r := NewRegisterFirstRule(true, nil, func(k, v string, e int) error { return nil },
		context.Background(), "a@b.com", types.UserAuthOptions{})
	if !r.Fails() {
		t.Fatal("expected fail without UserExists")
	}
	if !r.Degraded() || r.Reason() == "" {
		t.Fatalf("expected degraded with reason, got degraded=%v reason=%q", r.Degraded(), r.Reason())
	}
}

func TestRegisterFirstRule_MissingKeySet(t *testing.T) {
	r := NewRegisterFirstRule(true,
		func(ctx context.Context, email string, o types.UserAuthOptions) (bool, error) { return false, nil },
		nil, context.Background(), "a@b.com", types.UserAuthOptions{})
	if !r.Fails() || !r.Degraded() {
		t.Fatal("expected degraded failure without TemporaryKeySet")
	}
}

func TestRegisterFirstRule_CheckError(t *testing.T) {
	r := NewRegisterFirstRule(true,
		func(ctx context.Context, email string, o types.UserAuthOptions) (bool, error) {
			return false, errors.New("store down")
		},
		func(k, v string, e int) error { return nil },
		context.Background(), "a@b.com", types.UserAuthOptions{})
	if !r.Fails() || !r.Degraded() {
		t.Fatal("expected degraded failure on existence-check error")
	}
}

func TestRegisterFirstRule_ExistingAccount(t *testing.T) {
	r := NewRegisterFirstRule(true,
		func(ctx context.Context, email string, o types.UserAuthOptions) (bool, error) { return true, nil },
		func(k, v string, e int) error { return nil },
		context.Background(), "a@b.com", types.UserAuthOptions{})
	if !r.Fails() {
		t.Fatal("expected fail for existing account (normal login)")
	}
	if r.Degraded() {
		t.Fatal("existing account is a normal fail, not degradation")
	}
}

func TestRegisterFirstRule_Applies(t *testing.T) {
	r := NewRegisterFirstRule(true,
		func(ctx context.Context, email string, o types.UserAuthOptions) (bool, error) { return false, nil },
		func(k, v string, e int) error { return nil },
		context.Background(), "a@b.com", types.UserAuthOptions{})
	if r.Fails() {
		t.Fatal("expected rule to apply for new email")
	}
	if r.Degraded() {
		t.Fatal("applied rule must not be degraded")
	}
}
