package rules

import (
	"testing"

	"github.com/dracory/auth/types"
)

func TestPasswordStrengthRule_NilConfigPasses(t *testing.T) {
	r := NewPasswordStrengthRule("any", nil)
	if r.Fails() {
		t.Fatal("expected nil config to pass any password")
	}
}

func TestPasswordStrengthRule_ValidPasswordPasses(t *testing.T) {
	cfg := &types.PasswordStrengthConfig{
		MinLength:    8,
		RequireDigit: true,
	}
	r := NewPasswordStrengthRule("password123", cfg)
	if r.Fails() {
		t.Fatal("expected valid password to pass")
	}
}

func TestPasswordStrengthRule_ShortPasswordFails(t *testing.T) {
	cfg := &types.PasswordStrengthConfig{
		MinLength: 8,
	}
	r := NewPasswordStrengthRule("short", cfg)
	if r.Passes() {
		t.Fatal("expected short password to fail")
	}
	if r.FailMessageFirst() == "" {
		t.Fatal("expected non-empty fail message")
	}
}
