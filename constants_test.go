package auth

import "testing"

func TestContextKey_String(t *testing.T) {
	c := contextKey("test-key")
	if c.String() != "test-key" {
		t.Fatalf("expected %q, got %q", "test-key", c.String())
	}
}

func TestContextKey_EmptyString(t *testing.T) {
	c := contextKey("")
	if c.String() != "" {
		t.Fatalf("expected empty string, got %q", c.String())
	}
}

func TestContextKey_KeyEndpoint(t *testing.T) {
	if keyEndpoint.String() != "endpoint" {
		t.Fatalf("expected %q, got %q", "endpoint", keyEndpoint.String())
	}
}
