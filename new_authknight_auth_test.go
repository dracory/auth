package auth

import (
	"context"
	"testing"
	"time"

	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/types"
)

func newValidAuthKnightConfig() types.ConfigAuthKnight {
	return types.ConfigAuthKnight{
		ConfigShared: types.ConfigShared{
			Endpoint:             "/auth",
			UrlRedirectOnSuccess: "/dashboard",
			UseCookies:           true,
			FuncTemporaryKeyGet:  func(key string) (string, error) { return "", nil },
			FuncTemporaryKeySet:  func(key string, value string, expiresSeconds int) error { return nil },
			FuncUserFindByAuthToken: func(ctx context.Context, token string, options types.UserAuthOptions) (string, error) {
				return "", nil
			},
			FuncUserLogout: func(ctx context.Context, userID string, options types.UserAuthOptions) error {
				return nil
			},
			FuncUserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
				return nil
			},
		},
		FuncUserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
			return "user-1", nil
		},
		FuncUserRegister: func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (string, error) {
			return "user-2", nil
		},
		HTTPTimeout: 15 * time.Second,
	}
}

func TestNewAuthKnightAuth_Success(t *testing.T) {
	config := newValidAuthKnightConfig()
	a, err := NewAuthKnightAuth(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !a.IsAuthKnight() {
		t.Fatal("expected IsAuthKnight to be true")
	}

	if a.GetEndpoint() != "/auth" {
		t.Fatalf("expected endpoint '/auth', got '%s'", a.GetEndpoint())
	}

	if a.LinkRedirectOnSuccess() != "/dashboard" {
		t.Fatalf("expected redirect '/dashboard', got '%s'", a.LinkRedirectOnSuccess())
	}

	if a.GetAuthKnightHTTPTimeout() != 15*time.Second {
		t.Fatalf("expected HTTPTimeout 15s, got %v", a.GetAuthKnightHTTPTimeout())
	}

	if a.GetAuthKnightUserFindByEmail() == nil {
		t.Fatal("expected GetAuthKnightUserFindByEmail to be non-nil")
	}

	if a.GetAuthKnightUserRegister() == nil {
		t.Fatal("expected GetAuthKnightUserRegister to be non-nil")
	}
}

func TestNewAuthKnightAuth_DefaultHTTPTimeout(t *testing.T) {
	config := newValidAuthKnightConfig()
	config.HTTPTimeout = 0
	a, err := NewAuthKnightAuth(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.GetAuthKnightHTTPTimeout() != 10*time.Second {
		t.Fatalf("expected default HTTPTimeout 10s, got %v", a.GetAuthKnightHTTPTimeout())
	}
}

func TestNewAuthKnightAuth_MissingEndpoint(t *testing.T) {
	config := newValidAuthKnightConfig()
	config.Endpoint = ""
	_, err := NewAuthKnightAuth(config)
	if err == nil {
		t.Fatal("expected error for missing Endpoint")
	}
}

func TestNewAuthKnightAuth_MissingUrlRedirectOnSuccess(t *testing.T) {
	config := newValidAuthKnightConfig()
	config.UrlRedirectOnSuccess = ""
	_, err := NewAuthKnightAuth(config)
	if err == nil {
		t.Fatal("expected error for missing UrlRedirectOnSuccess")
	}
}

func TestNewAuthKnightAuth_MissingFuncUserFindByEmail(t *testing.T) {
	config := newValidAuthKnightConfig()
	config.FuncUserFindByEmail = nil
	_, err := NewAuthKnightAuth(config)
	if err == nil {
		t.Fatal("expected error for missing FuncUserFindByEmail")
	}
}

func TestNewAuthKnightAuth_MissingFuncUserRegister(t *testing.T) {
	config := newValidAuthKnightConfig()
	config.FuncUserRegister = nil
	_, err := NewAuthKnightAuth(config)
	if err == nil {
		t.Fatal("expected error for missing FuncUserRegister")
	}
}

func TestNewAuthKnightAuth_MissingFuncUserStoreAuthToken(t *testing.T) {
	config := newValidAuthKnightConfig()
	config.FuncUserStoreAuthToken = nil
	_, err := NewAuthKnightAuth(config)
	if err == nil {
		t.Fatal("expected error for missing FuncUserStoreAuthToken")
	}
}

func TestNewAuthKnightAuth_MissingFuncTemporaryKeyGet(t *testing.T) {
	config := newValidAuthKnightConfig()
	config.FuncTemporaryKeyGet = nil
	_, err := NewAuthKnightAuth(config)
	if err == nil {
		t.Fatal("expected error for missing FuncTemporaryKeyGet")
	}
}

func TestNewAuthKnightAuth_MissingFuncTemporaryKeySet(t *testing.T) {
	config := newValidAuthKnightConfig()
	config.FuncTemporaryKeySet = nil
	_, err := NewAuthKnightAuth(config)
	if err == nil {
		t.Fatal("expected error for missing FuncTemporaryKeySet")
	}
}

func TestNewAuthKnightAuth_MissingFuncUserFindByAuthToken(t *testing.T) {
	config := newValidAuthKnightConfig()
	config.FuncUserFindByAuthToken = nil
	_, err := NewAuthKnightAuth(config)
	if err == nil {
		t.Fatal("expected error for missing FuncUserFindByAuthToken")
	}
}

func TestNewAuthKnightAuth_MissingFuncUserLogout(t *testing.T) {
	config := newValidAuthKnightConfig()
	config.FuncUserLogout = nil
	_, err := NewAuthKnightAuth(config)
	if err == nil {
		t.Fatal("expected error for missing FuncUserLogout")
	}
}

func TestNewAuthKnightAuth_CSRFWithoutSecret(t *testing.T) {
	config := newValidAuthKnightConfig()
	config.EnableCSRFProtection = true
	config.CSRFSecret = ""
	_, err := NewAuthKnightAuth(config)
	if err == nil {
		t.Fatal("expected error for CSRF enabled without secret")
	}
}

func TestNewAuthKnightAuth_WithCSRF(t *testing.T) {
	config := newValidAuthKnightConfig()
	config.EnableCSRFProtection = true
	config.CSRFSecret = "test-secret"
	a, err := NewAuthKnightAuth(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	concrete, ok := a.(*authImplementation)
	if !ok {
		t.Fatal("expected *authImplementation")
	}

	if concrete.csrfSecret != "test-secret" {
		t.Fatalf("expected csrfSecret 'test-secret', got '%s'", concrete.csrfSecret)
	}
	if concrete.funcCSRFTokenGenerate == nil {
		t.Fatal("expected funcCSRFTokenGenerate to be set")
	}
	if concrete.funcCSRFTokenValidate == nil {
		t.Fatal("expected funcCSRFTokenValidate to be set")
	}
}

func TestNewAuthKnightAuth_LinkAuthKnightCallback(t *testing.T) {
	config := newValidAuthKnightConfig()
	config.Endpoint = "http://localhost/auth"
	a, err := NewAuthKnightAuth(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "http://localhost/auth/api/authknight/callback"
	got := a.LinkAuthKnightCallback()
	if got != expected {
		t.Fatalf("expected '%s', got '%s'", expected, got)
	}
}

func TestNewAuthKnightAuth_LinkAuthKnightLogin(t *testing.T) {
	config := newValidAuthKnightConfig()
	a, err := NewAuthKnightAuth(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := a.LinkAuthKnightLogin("http://localhost/auth/api/authknight/callback", "/dashboard")
	if got == "" {
		t.Fatal("expected non-empty login URL")
	}
}

func TestNewAuthKnightAuth_NotPasswordless(t *testing.T) {
	config := newValidAuthKnightConfig()
	a, err := NewAuthKnightAuth(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.IsPasswordless() {
		t.Fatal("expected IsPasswordless to be false for AuthKnight")
	}
}

func TestNewAuthKnightAuth_Router(t *testing.T) {
	config := newValidAuthKnightConfig()
	a, err := NewAuthKnightAuth(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	router := a.Router()
	if router == nil {
		t.Fatal("expected non-nil router")
	}
}

func TestNewAuthKnightAuth_WithRedirectURL(t *testing.T) {
	config := newValidAuthKnightConfig()
	config.FuncRedirectURL = func(ctx context.Context, userID string) string {
		return "/custom/" + userID
	}
	a, err := NewAuthKnightAuth(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.GetAuthKnightRedirectURL() == nil {
		t.Fatal("expected GetAuthKnightRedirectURL to be non-nil")
	}
}

func TestNewAuthKnightAuth_WithRateLimiting(t *testing.T) {
	config := newValidAuthKnightConfig()
	config.DisableRateLimit = false
	config.MaxLoginAttempts = 10
	config.LockoutDuration = 30 * time.Minute
	a, err := NewAuthKnightAuth(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	concrete, ok := a.(*authImplementation)
	if !ok {
		t.Fatal("expected *authImplementation")
	}
	if concrete.rateLimiter == nil {
		t.Fatal("expected rate limiter to be initialized")
	}
}

func TestNewAuthKnightAuth_DisableRateLimit(t *testing.T) {
	config := newValidAuthKnightConfig()
	config.DisableRateLimit = true
	a, err := NewAuthKnightAuth(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	concrete, ok := a.(*authImplementation)
	if !ok {
		t.Fatal("expected *authImplementation")
	}
	if !concrete.disableRateLimit {
		t.Fatal("expected disableRateLimit to be true")
	}
}

func TestNewAuthKnightAuth_TestUtilsConfig(t *testing.T) {
	config := testutils.NewAuthKnightConfigForTest()
	a, err := NewAuthKnightAuth(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !a.IsAuthKnight() {
		t.Fatal("expected IsAuthKnight to be true")
	}
}
