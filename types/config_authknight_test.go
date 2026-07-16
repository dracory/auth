package types

import (
	"context"
	"testing"
	"time"
)

func TestConfigAuthKnight_Fields(t *testing.T) {
	c := ConfigAuthKnight{
		ConfigShared: ConfigShared{
			Endpoint:             "/auth",
			UrlRedirectOnSuccess: "/dashboard",
			UseCookies:           true,
		},
		ConfigRateLimiting: ConfigRateLimiting{
			MaxLoginAttempts: 5,
		},
		ConfigCSRF: ConfigCSRF{
			EnableCSRFProtection: true,
			CSRFSecret:           "secret",
		},
		ConfigImpersonation: ConfigImpersonation{
			EnableImpersonation: false,
		},
		FuncUserFindByEmail: func(ctx context.Context, email string, options UserAuthOptions) (string, error) {
			return "user-1", nil
		},
		FuncUserRegister: func(ctx context.Context, email, firstName, lastName string, options UserAuthOptions) (string, error) {
			return "user-2", nil
		},
		FuncRedirectURL: func(ctx context.Context, userID string) string {
			return "/custom-redirect"
		},
		HTTPTimeout: 15 * time.Second,
	}

	if c.Endpoint != "/auth" {
		t.Fatalf("expected Endpoint '/auth', got '%s'", c.Endpoint)
	}
	if c.UrlRedirectOnSuccess != "/dashboard" {
		t.Fatalf("expected UrlRedirectOnSuccess '/dashboard', got '%s'", c.UrlRedirectOnSuccess)
	}
	if !c.UseCookies {
		t.Fatal("UseCookies should be true")
	}
	if c.MaxLoginAttempts != 5 {
		t.Fatalf("expected MaxLoginAttempts 5, got %d", c.MaxLoginAttempts)
	}
	if !c.EnableCSRFProtection {
		t.Fatal("EnableCSRFProtection should be true")
	}
	if c.CSRFSecret != "secret" {
		t.Fatalf("expected CSRFSecret 'secret', got '%s'", c.CSRFSecret)
	}
	if c.FuncUserFindByEmail == nil {
		t.Fatal("FuncUserFindByEmail should not be nil")
	}
	if c.FuncUserRegister == nil {
		t.Fatal("FuncUserRegister should not be nil")
	}
	if c.FuncRedirectURL == nil {
		t.Fatal("FuncRedirectURL should not be nil")
	}
	if c.HTTPTimeout != 15*time.Second {
		t.Fatalf("expected HTTPTimeout 15s, got %v", c.HTTPTimeout)
	}
}

func TestConfigAuthKnight_ZeroValue(t *testing.T) {
	var c ConfigAuthKnight
	if c.Endpoint != "" {
		t.Fatalf("expected empty Endpoint, got '%s'", c.Endpoint)
	}
	if c.FuncUserFindByEmail != nil {
		t.Fatal("FuncUserFindByEmail should be nil by default")
	}
	if c.FuncUserRegister != nil {
		t.Fatal("FuncUserRegister should be nil by default")
	}
	if c.FuncRedirectURL != nil {
		t.Fatal("FuncRedirectURL should be nil by default")
	}
	if c.HTTPTimeout != 0 {
		t.Fatalf("expected HTTPTimeout 0, got %v", c.HTTPTimeout)
	}
}

func TestConfigAuthKnight_FieldPromotion(t *testing.T) {
	c := ConfigAuthKnight{
		ConfigShared: ConfigShared{
			Endpoint:             "/auth",
			UrlRedirectOnSuccess: "/dashboard",
			UseCookies:           true,
		},
		ConfigRateLimiting: ConfigRateLimiting{
			MaxLoginAttempts: 3,
		},
		ConfigCSRF: ConfigCSRF{
			EnableCSRFProtection: true,
			CSRFSecret:           "secret",
		},
		ConfigImpersonation: ConfigImpersonation{
			EnableImpersonation: true,
		},
	}

	if c.Endpoint != "/auth" {
		t.Fatalf("expected promoted Endpoint '/auth', got '%s'", c.Endpoint)
	}
	if c.UrlRedirectOnSuccess != "/dashboard" {
		t.Fatalf("expected promoted UrlRedirectOnSuccess '/dashboard', got '%s'", c.UrlRedirectOnSuccess)
	}
	if !c.UseCookies {
		t.Fatal("expected promoted UseCookies true")
	}
	if c.MaxLoginAttempts != 3 {
		t.Fatalf("expected promoted MaxLoginAttempts 3, got %d", c.MaxLoginAttempts)
	}
	if !c.EnableCSRFProtection {
		t.Fatal("expected promoted EnableCSRFProtection true")
	}
	if c.CSRFSecret != "secret" {
		t.Fatalf("expected promoted CSRFSecret 'secret', got '%s'", c.CSRFSecret)
	}
	if !c.EnableImpersonation {
		t.Fatal("expected promoted EnableImpersonation true")
	}
}
