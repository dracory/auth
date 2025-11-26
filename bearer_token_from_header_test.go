package auth

import "testing"

func TestBearerTokenFromHeader_EmptyHeaderReturnsEmpty(t *testing.T) {
	if token := BearerTokenFromHeader(""); token != "" {
		t.Fatalf("expected empty token, got %q", token)
	}
}

func TestBearerTokenFromHeader_NoBearerPrefixReturnsEmpty(t *testing.T) {
	header := "Basic abcdef123456"
	if token := BearerTokenFromHeader(header); token != "" {
		t.Fatalf("expected empty token, got %q", token)
	}
}

func TestBearerTokenFromHeader_MalformedBearerReturnsEmpty(t *testing.T) {
	header := "Bearer"
	if token := BearerTokenFromHeader(header); token != "" {
		t.Fatalf("expected empty token, got %q", token)
	}
}

func TestBearerTokenFromHeader_ValidBearerToken(t *testing.T) {
	header := "Bearer my-token"
	if token := BearerTokenFromHeader(header); token != "my-token" {
		t.Fatalf("expected token %q, got %q", "my-token", token)
	}
}

func TestBearerTokenFromHeader_AllowsExtraSpaces(t *testing.T) {
	header := "  Bearer   spaced-token  "
	if token := BearerTokenFromHeader(header); token != "spaced-token" {
		t.Fatalf("expected token %q, got %q", "spaced-token", token)
	}
}
