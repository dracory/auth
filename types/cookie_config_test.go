package types

import (
	"net/http"
	"testing"
)

func TestWithSecure(t *testing.T) {
	cfg := &CookieConfig{}
	WithSecure(true)(cfg)
	if cfg.Secure != CookieSecure {
		t.Fatalf("expected Secure %q, got %q", CookieSecure, cfg.Secure)
	}
	WithSecure(false)(cfg)
	if cfg.Secure != CookieInsecure {
		t.Fatalf("expected Secure %q, got %q", CookieInsecure, cfg.Secure)
	}
}

func TestWithHttpOnly(t *testing.T) {
	cfg := &CookieConfig{}
	WithHttpOnly(true)(cfg)
	if cfg.HttpOnly != CookieHttpOnly {
		t.Fatalf("expected HttpOnly %q, got %q", CookieHttpOnly, cfg.HttpOnly)
	}
	WithHttpOnly(false)(cfg)
	if cfg.HttpOnly != CookieHttpWritable {
		t.Fatalf("expected HttpOnly %q, got %q", CookieHttpWritable, cfg.HttpOnly)
	}
}

func TestWithSameSite(t *testing.T) {
	cfg := &CookieConfig{}
	WithSameSite(http.SameSiteStrictMode)(cfg)
	if cfg.SameSite != http.SameSiteStrictMode {
		t.Fatalf("expected SameSite to be %d, got %d", http.SameSiteStrictMode, cfg.SameSite)
	}
	WithSameSite(http.SameSiteLaxMode)(cfg)
	if cfg.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected SameSite to be %d, got %d", http.SameSiteLaxMode, cfg.SameSite)
	}
}

func TestWithMaxAge(t *testing.T) {
	cfg := &CookieConfig{}
	WithMaxAge(3600)(cfg)
	if cfg.MaxAge != 3600 {
		t.Fatalf("expected MaxAge to be 3600, got %d", cfg.MaxAge)
	}
}

func TestWithDomain(t *testing.T) {
	cfg := &CookieConfig{}
	WithDomain("example.com")(cfg)
	if cfg.Domain != "example.com" {
		t.Fatalf("expected Domain to be 'example.com', got %q", cfg.Domain)
	}
}

func TestWithPath(t *testing.T) {
	cfg := &CookieConfig{}
	WithPath("/auth")(cfg)
	if cfg.Path != "/auth" {
		t.Fatalf("expected Path to be '/auth', got %q", cfg.Path)
	}
}

func TestWithCookieConfig(t *testing.T) {
	cfg := &CookieConfig{}
	replacement := CookieConfig{
		HttpOnly: CookieHttpOnly,
		Secure:   CookieSecure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7200,
		Domain:   "test.com",
		Path:     "/api",
	}
	WithCookieConfig(replacement)(cfg)
	if cfg.HttpOnly != replacement.HttpOnly {
		t.Fatal("expected HttpOnly to match replacement")
	}
	if cfg.Secure != replacement.Secure {
		t.Fatal("expected Secure to match replacement")
	}
	if cfg.SameSite != replacement.SameSite {
		t.Fatal("expected SameSite to match replacement")
	}
	if cfg.MaxAge != replacement.MaxAge {
		t.Fatal("expected MaxAge to match replacement")
	}
	if cfg.Domain != replacement.Domain {
		t.Fatal("expected Domain to match replacement")
	}
	if cfg.Path != replacement.Path {
		t.Fatal("expected Path to match replacement")
	}
}

func TestResolveSecure(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{CookieUnset, true},
		{CookieSecure, true},
		{CookieInsecure, false},
		{"", true},
	}
	for _, tc := range tests {
		if got := ResolveSecure(tc.input); got != tc.expected {
			t.Fatalf("ResolveSecure(%q) = %v, want %v", tc.input, got, tc.expected)
		}
	}
}

func TestResolveHttpOnly(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{CookieUnset, true},
		{CookieHttpOnly, true},
		{CookieHttpWritable, false},
		{"", true},
	}
	for _, tc := range tests {
		if got := ResolveHttpOnly(tc.input); got != tc.expected {
			t.Fatalf("ResolveHttpOnly(%q) = %v, want %v", tc.input, got, tc.expected)
		}
	}
}
