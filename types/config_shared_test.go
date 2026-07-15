package types

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestConfigShared_Fields(t *testing.T) {
	c := ConfigShared{
		EnableRegistration:      true,
		Endpoint:                "/auth",
		UrlRedirectOnSuccess:    "/dashboard",
		UseCookies:              true,
		UseLocalStorage:         false,
		FuncLayout:              func(content string) string { return content },
		FuncTemporaryKeyGet:     func(key string) (string, error) { return "", nil },
		FuncTemporaryKeySet:     func(key string, value string, expiresSeconds int) error { return nil },
		FuncUserStoreAuthToken:  func(ctx context.Context, sessionID string, userID string, options UserAuthOptions) error { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options UserAuthOptions) (string, error) { return "", nil },
		FuncUserLogout:          func(ctx context.Context, userID string, options UserAuthOptions) error { return nil },
		CookieConfig:            &CookieConfig{Name: "test-cookie"},
		Logger:                  slog.Default(),
		ObservabilityHooks:      NoopObservabilityHooks{},
	}

	if !c.EnableRegistration {
		t.Fatal("EnableRegistration should be true")
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
	if c.UseLocalStorage {
		t.Fatal("UseLocalStorage should be false")
	}
	if c.FuncLayout == nil {
		t.Fatal("FuncLayout should not be nil")
	}
	if c.FuncTemporaryKeyGet == nil {
		t.Fatal("FuncTemporaryKeyGet should not be nil")
	}
	if c.FuncTemporaryKeySet == nil {
		t.Fatal("FuncTemporaryKeySet should not be nil")
	}
	if c.FuncUserStoreAuthToken == nil {
		t.Fatal("FuncUserStoreAuthToken should not be nil")
	}
	if c.FuncUserFindByAuthToken == nil {
		t.Fatal("FuncUserFindByAuthToken should not be nil")
	}
	if c.FuncUserLogout == nil {
		t.Fatal("FuncUserLogout should not be nil")
	}
	if c.CookieConfig == nil || c.CookieConfig.Name != "test-cookie" {
		t.Fatal("CookieConfig should be set with name 'test-cookie'")
	}
	if c.Logger == nil {
		t.Fatal("Logger should not be nil")
	}
	if c.ObservabilityHooks == nil {
		t.Fatal("ObservabilityHooks should not be nil")
	}
}

func TestConfigShared_ZeroValue(t *testing.T) {
	var c ConfigShared
	if c.EnableRegistration {
		t.Fatal("EnableRegistration should be false by default")
	}
	if c.Endpoint != "" {
		t.Fatalf("expected empty Endpoint, got '%s'", c.Endpoint)
	}
	if c.CookieConfig != nil {
		t.Fatal("CookieConfig should be nil by default")
	}
	if c.Logger != nil {
		t.Fatal("Logger should be nil by default")
	}
}

func TestConfigRateLimiting_Fields(t *testing.T) {
	c := ConfigRateLimiting{
		DisableRateLimit:   false,
		FuncCheckRateLimit: func(ip string, endpoint string) (bool, time.Duration, error) { return true, 0, nil },
		MaxLoginAttempts:   5,
		LockoutDuration:    15 * time.Minute,
	}

	if c.DisableRateLimit {
		t.Fatal("DisableRateLimit should be false")
	}
	if c.FuncCheckRateLimit == nil {
		t.Fatal("FuncCheckRateLimit should not be nil")
	}
	if c.MaxLoginAttempts != 5 {
		t.Fatalf("expected MaxLoginAttempts 5, got %d", c.MaxLoginAttempts)
	}
	if c.LockoutDuration != 15*time.Minute {
		t.Fatalf("expected LockoutDuration 15m, got %v", c.LockoutDuration)
	}
}

func TestConfigRateLimiting_ZeroValue(t *testing.T) {
	var c ConfigRateLimiting
	if c.DisableRateLimit {
		t.Fatal("DisableRateLimit should be false by default")
	}
	if c.MaxLoginAttempts != 0 {
		t.Fatalf("expected MaxLoginAttempts 0, got %d", c.MaxLoginAttempts)
	}
	if c.LockoutDuration != 0 {
		t.Fatalf("expected LockoutDuration 0, got %v", c.LockoutDuration)
	}
}

func TestConfigCSRF_Fields(t *testing.T) {
	c := ConfigCSRF{
		EnableCSRFProtection: true,
		CSRFSecret:           "super-secret",
	}

	if !c.EnableCSRFProtection {
		t.Fatal("EnableCSRFProtection should be true")
	}
	if c.CSRFSecret != "super-secret" {
		t.Fatalf("expected CSRFSecret 'super-secret', got '%s'", c.CSRFSecret)
	}
}

func TestConfigCSRF_ZeroValue(t *testing.T) {
	var c ConfigCSRF
	if c.EnableCSRFProtection {
		t.Fatal("EnableCSRFProtection should be false by default")
	}
	if c.CSRFSecret != "" {
		t.Fatalf("expected empty CSRFSecret, got '%s'", c.CSRFSecret)
	}
}

func TestConfigImpersonation_Fields(t *testing.T) {
	c := ConfigImpersonation{
		EnableImpersonation: true,
		FuncCanImpersonate: func(ctx context.Context, adminUserID string, targetUserID string) (bool, error) {
			return true, nil
		},
		FuncImpersonationStart: func(ctx context.Context, adminUserID string, targetUserID string) error { return nil },
		FuncImpersonationStop:  func(ctx context.Context, adminUserID string, targetUserID string) error { return nil },
	}

	if !c.EnableImpersonation {
		t.Fatal("EnableImpersonation should be true")
	}
	if c.FuncCanImpersonate == nil {
		t.Fatal("FuncCanImpersonate should not be nil")
	}
	if c.FuncImpersonationStart == nil {
		t.Fatal("FuncImpersonationStart should not be nil")
	}
	if c.FuncImpersonationStop == nil {
		t.Fatal("FuncImpersonationStop should not be nil")
	}
}

func TestConfigImpersonation_ZeroValue(t *testing.T) {
	var c ConfigImpersonation
	if c.EnableImpersonation {
		t.Fatal("EnableImpersonation should be false by default")
	}
}

func TestConfigPasswordless_FieldPromotion(t *testing.T) {
	c := ConfigPasswordless{
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

func TestConfigUsernameAndPassword_FieldPromotion(t *testing.T) {
	c := ConfigUsernameAndPassword{
		ConfigShared: ConfigShared{
			Endpoint:             "/auth",
			UrlRedirectOnSuccess: "/dashboard",
			UseCookies:           true,
			FuncUserLogout:       func(ctx context.Context, userID string, options UserAuthOptions) error { return nil },
		},
		ConfigRateLimiting: ConfigRateLimiting{
			MaxLoginAttempts: 10,
		},
		ConfigCSRF: ConfigCSRF{
			EnableCSRFProtection: false,
		},
		ConfigImpersonation: ConfigImpersonation{
			EnableImpersonation: false,
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
	if c.FuncUserLogout == nil {
		t.Fatal("expected promoted FuncUserLogout to be set")
	}
	if c.MaxLoginAttempts != 10 {
		t.Fatalf("expected promoted MaxLoginAttempts 10, got %d", c.MaxLoginAttempts)
	}
	if c.EnableCSRFProtection {
		t.Fatal("expected promoted EnableCSRFProtection false")
	}
	if c.EnableImpersonation {
		t.Fatal("expected promoted EnableImpersonation false")
	}
}
