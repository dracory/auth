package utils

import "testing"

func TestLoginCodeLength_DefaultAndHardened(t *testing.T) {
	tests := []struct {
		name          string
		extraHardened bool
		want          int
	}{
		{"default_length", false, 8},
		{"hardened_length", true, 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LoginCodeLength(tt.extraHardened)
			if got != tt.want {
				t.Fatalf("LoginCodeLength(%v) = %d, want %d", tt.extraHardened, got, tt.want)
			}
		})
	}
}

func TestLoginCodeGamma_DefaultAndHardened(t *testing.T) {
	tests := []struct {
		name          string
		extraHardened bool
		minLen        int
		mustContain   []rune
	}{
		{
			"default_gamma",
			false,
			20,
			[]rune{'B', 'C', 'D'},
		},
		{
			"hardened_gamma",
			true,
			43,
			[]rune{'0', '1', '9'},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gamma := LoginCodeGamma(tt.extraHardened)
			if len(gamma) < tt.minLen {
				t.Fatalf("gamma length too short: got %d, want >= %d", len(gamma), tt.minLen)
			}
			for _, ch := range tt.mustContain {
				if !containsRune(gamma, ch) {
					t.Fatalf("gamma %q does not contain required rune %q", gamma, ch)
				}
			}
		})
	}
}

func TestGenerateVerificationCode_DefaultAndHardened(t *testing.T) {
	tests := []struct {
		name          string
		extraHardened bool
	}{
		{"default_verification_code", false},
		{"hardened_verification_code", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := GenerateVerificationCode(tt.extraHardened)
			if err != nil {
				t.Fatalf("GenerateVerificationCode(%v) returned error: %v", tt.extraHardened, err)
			}

			expectedLen := LoginCodeLength(tt.extraHardened)
			if len(code) != expectedLen {
				t.Fatalf("expected code length %d, got %d", expectedLen, len(code))
			}

			gamma := LoginCodeGamma(tt.extraHardened)
			for _, ch := range code {
				if !containsRune(gamma, ch) {
					t.Fatalf("code %q contains invalid character %q (not in gamma %q)", code, ch, gamma)
				}
			}
		})
	}
}

func TestGeneratePasswordResetToken(t *testing.T) {
	token, err := GeneratePasswordResetToken()
	if err != nil {
		t.Fatalf("GeneratePasswordResetToken returned error: %v", err)
	}

	if len(token) != 32 {
		t.Fatalf("expected token length 32, got %d", len(token))
	}

	const gamma = "BCDFGHJKLMNPQRSTVXYZ"
	for _, ch := range token {
		if !containsRune(gamma, ch) {
			t.Fatalf("token %q contains invalid character %q (not in gamma %q)", token, ch, gamma)
		}
	}
}

func containsRune(s string, r rune) bool {
	for _, ch := range s {
		if ch == r {
			return true
		}
	}
	return false
}
