package utils

import (
	"testing"

	authtypes "github.com/dracory/auth/types"
)

func TestValidatePasswordStrength_NilConfigAllowsAnyPassword(t *testing.T) {
	if err := ValidatePasswordStrength("a", nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidatePasswordStrength_MinLength(t *testing.T) {
	cfg := &authtypes.PasswordStrengthConfig{MinLength: 8}

	if err := ValidatePasswordStrength("short", cfg); err == nil {
		t.Fatalf("expected error for short password, got nil")
	}

	if err := ValidatePasswordStrength("longenough", cfg); err != nil {
		t.Fatalf("expected no error for sufficient length, got %v", err)
	}
}

func TestValidatePasswordStrength_CharacterRequirements(t *testing.T) {
	cfg := &authtypes.PasswordStrengthConfig{
		MinLength:        1,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
		RequireSpecial:   true,
	}

	cases := []struct {
		name    string
		pwd     string
		wantErr bool
	}{
		{"missing all", "", true},
		{"missing upper", "a1!", true},
		{"missing lower", "A1!", true},
		{"missing digit", "Aa!", true},
		{"missing special", "Aa1", true},
		{"all present", "Aa1!", false},
	}

	for _, tc := range cases {
		if err := ValidatePasswordStrength(tc.pwd, cfg); (err != nil) != tc.wantErr {
			t.Fatalf("%s: expected error=%v, got %v", tc.name, tc.wantErr, err)
		}
	}
}

func TestValidatePasswordStrength_ForbidCommonWords(t *testing.T) {
	cfg := &authtypes.PasswordStrengthConfig{
		ForbidCommonWords: true,
	}

	if err := ValidatePasswordStrength("password", cfg); err == nil {
		t.Fatalf("expected error for common password, got nil")
	}

	if err := ValidatePasswordStrength("uniquepassword", cfg); err != nil {
		t.Fatalf("expected no error for non-common password, got %v", err)
	}
}
