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

func containsRune(s string, r rune) bool {
	for _, ch := range s {
		if ch == r {
			return true
		}
	}
	return false
}
