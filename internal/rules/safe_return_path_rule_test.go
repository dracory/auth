package rules

import "testing"

func TestSafeReturnPathRule_EmptyPasses(t *testing.T) {
	r := NewSafeReturnPathRule("")
	if r.Fails() {
		t.Fatal("expected empty path to pass")
	}
}

func TestSafeReturnPathRule_RelativePasses(t *testing.T) {
	for _, p := range []string{"/dashboard", "/users/profile?tab=1", "/"} {
		if r := NewSafeReturnPathRule(p); r.Fails() {
			t.Fatalf("expected %q to pass", p)
		}
	}
}

func TestSafeReturnPathRule_ExternalFails(t *testing.T) {
	for _, p := range []string{
		"https://evil.com", "http://evil.com", "//evil.com/x", "javascript:alert(1)",
	} {
		if r := NewSafeReturnPathRule(p); !r.Fails() {
			t.Fatalf("expected %q to fail", p)
		}
	}
}
