package api_authknight_callback

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dracory/auth/types"
)

// mockAuthKnightServer creates a test HTTP server that simulates AuthKnight's /api/who endpoint.
func mockAuthKnightServer(email string, status string, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		once := r.FormValue("once")
		resp := authKnightWhoResponse{
			Status:  status,
			Message: "success",
		}
		resp.Data.Email = email
		resp.Data.BackURL = "http://localhost:8083/"

		_ = once // just consume it
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func TestApiAuthKnightCallback_ExistingUser(t *testing.T) {
	var storedToken string
	var storedUserID string
	var cookieSet bool

	server := mockAuthKnightServer("user@example.com", "success", http.StatusOK)
	defer server.Close()

	deps := Dependencies{
		BaseURL:   server.URL,
		HTTPTimeout: 5 * time.Second,
		UserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
			if email != "user@example.com" {
				return "", fmt.Errorf("user not found")
			}
			return "user-123", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
			storedToken = token
			storedUserID = userID
			return nil
		},
		TemporaryKeySet: func(key, value string, expiresSeconds int) error {
			t.Fatal("should not store temp key for existing user")
			return nil
		},
		TemporaryKeyGet: func(key string) (string, error) {
			return "", nil
		},
		UrlRedirectOnSuccess: "/dashboard",
		RegisterURL:          "/register",
		UseCookies:           true,
		SetAuthCookie: func(w http.ResponseWriter, r *http.Request, token string) {
			cookieSet = true
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/api/authknight/callback?once=test-token", nil)
	recorder := httptest.NewRecorder()

	ApiAuthKnightCallback(recorder, req, deps)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, recorder.Code)
	}

	location := recorder.Header().Get("Location")
	if location != "/dashboard" {
		t.Fatalf("expected redirect to /dashboard, got %s", location)
	}

	if storedUserID != "user-123" {
		t.Fatalf("expected stored userID 'user-123', got '%s'", storedUserID)
	}

	if storedToken == "" {
		t.Fatal("expected non-empty stored token")
	}

	if !cookieSet {
		t.Fatal("expected cookie to be set")
	}
}

func TestApiAuthKnightCallback_NewUser(t *testing.T) {
	var tempKeyStored string
	var tempValueStored string

	server := mockAuthKnightServer("newuser@example.com", "success", http.StatusOK)
	defer server.Close()

	deps := Dependencies{
		BaseURL:   server.URL,
		HTTPTimeout: 5 * time.Second,
		UserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
			return "", fmt.Errorf("user not found")
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
			t.Fatal("should not store auth token for new user")
			return nil
		},
		TemporaryKeySet: func(key, value string, expiresSeconds int) error {
			tempKeyStored = key
			tempValueStored = value
			return nil
		},
		TemporaryKeyGet: func(key string) (string, error) {
			return "", nil
		},
		UrlRedirectOnSuccess: "/dashboard",
		RegisterURL:          "/register",
		UseCookies:           true,
		SetAuthCookie:        func(w http.ResponseWriter, r *http.Request, token string) {},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/api/authknight/callback?once=test-token", nil)
	recorder := httptest.NewRecorder()

	ApiAuthKnightCallback(recorder, req, deps)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, recorder.Code)
	}

	location := recorder.Header().Get("Location")
	if !strings.HasPrefix(location, "/register?ak_key=") {
		t.Fatalf("expected redirect to /register?ak_key=..., got %s", location)
	}

	if tempKeyStored == "" {
		t.Fatal("expected non-empty temp key stored")
	}

	if tempValueStored != "newuser@example.com" {
		t.Fatalf("expected temp value 'newuser@example.com', got '%s'", tempValueStored)
	}
}

func TestApiAuthKnightCallback_MissingOnceParameter(t *testing.T) {
	deps := Dependencies{
		BaseURL:   "http://localhost",
		HTTPTimeout: 5 * time.Second,
		UserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
			return "", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
			return nil
		},
		TemporaryKeySet: func(key, value string, expiresSeconds int) error { return nil },
		TemporaryKeyGet: func(key string) (string, error) { return "", nil },
		UrlRedirectOnSuccess: "/dashboard",
		RegisterURL:          "/register",
		UseCookies:           true,
		SetAuthCookie:        func(w http.ResponseWriter, r *http.Request, token string) {},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/api/authknight/callback", nil)
	recorder := httptest.NewRecorder()

	ApiAuthKnightCallback(recorder, req, deps)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, recorder.Code)
	}

	location := recorder.Header().Get("Location")
	if location != "/dashboard" {
		t.Fatalf("expected redirect to /dashboard, got %s", location)
	}
}

func TestApiAuthKnightCallback_AuthKnightAPIError(t *testing.T) {
	server := mockAuthKnightServer("", "error", http.StatusOK)
	defer server.Close()

	deps := Dependencies{
		BaseURL:   server.URL,
		HTTPTimeout: 5 * time.Second,
		UserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
			t.Fatal("should not call UserFindByEmail on API error")
			return "", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
			return nil
		},
		TemporaryKeySet: func(key, value string, expiresSeconds int) error { return nil },
		TemporaryKeyGet: func(key string) (string, error) { return "", nil },
		UrlRedirectOnSuccess: "/dashboard",
		RegisterURL:          "/register",
		UseCookies:           true,
		SetAuthCookie:        func(w http.ResponseWriter, r *http.Request, token string) {},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/api/authknight/callback?once=test-token", nil)
	recorder := httptest.NewRecorder()

	ApiAuthKnightCallback(recorder, req, deps)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, recorder.Code)
	}

	location := recorder.Header().Get("Location")
	if location != "/dashboard" {
		t.Fatalf("expected redirect to /dashboard, got %s", location)
	}
}

func TestApiAuthKnightCallback_AuthKnightAPIUnreachable(t *testing.T) {
	deps := Dependencies{
		BaseURL:   "http://127.0.0.1:1", // unreachable
		HTTPTimeout: 1 * time.Second,
		UserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
			t.Fatal("should not call UserFindByEmail on unreachable API")
			return "", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
			return nil
		},
		TemporaryKeySet: func(key, value string, expiresSeconds int) error { return nil },
		TemporaryKeyGet: func(key string) (string, error) { return "", nil },
		UrlRedirectOnSuccess: "/dashboard",
		RegisterURL:          "/register",
		UseCookies:           true,
		SetAuthCookie:        func(w http.ResponseWriter, r *http.Request, token string) {},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/api/authknight/callback?once=test-token", nil)
	recorder := httptest.NewRecorder()

	ApiAuthKnightCallback(recorder, req, deps)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, recorder.Code)
	}

	location := recorder.Header().Get("Location")
	if location != "/dashboard" {
		t.Fatalf("expected redirect to /dashboard, got %s", location)
	}
}

func TestApiAuthKnightCallback_CustomRedirectURL(t *testing.T) {
	server := mockAuthKnightServer("user@example.com", "success", http.StatusOK)
	defer server.Close()

	deps := Dependencies{
		BaseURL:   server.URL,
		HTTPTimeout: 5 * time.Second,
		UserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
			return "user-456", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
			return nil
		},
		TemporaryKeySet: func(key, value string, expiresSeconds int) error { return nil },
		TemporaryKeyGet: func(key string) (string, error) { return "", nil },
		RedirectURL: func(ctx context.Context, userID string) string {
			return "/custom-redirect/" + userID
		},
		UrlRedirectOnSuccess: "/dashboard",
		RegisterURL:          "/register",
		UseCookies:           true,
		SetAuthCookie:        func(w http.ResponseWriter, r *http.Request, token string) {},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/api/authknight/callback?once=test-token", nil)
	recorder := httptest.NewRecorder()

	ApiAuthKnightCallback(recorder, req, deps)

	if recorder.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, recorder.Code)
	}

	location := recorder.Header().Get("Location")
	if location != "/custom-redirect/user-456" {
		t.Fatalf("expected redirect to /custom-redirect/user-456, got %s", location)
	}
}

func TestApiAuthKnightCallback_NewUserRegisterURLWithExistingQuery(t *testing.T) {
	server := mockAuthKnightServer("newuser@example.com", "success", http.StatusOK)
	defer server.Close()

	deps := Dependencies{
		BaseURL:   server.URL,
		HTTPTimeout: 5 * time.Second,
		UserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
			return "", fmt.Errorf("not found")
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
			return nil
		},
		TemporaryKeySet: func(key, value string, expiresSeconds int) error { return nil },
		TemporaryKeyGet: func(key string) (string, error) { return "", nil },
		UrlRedirectOnSuccess: "/dashboard",
		RegisterURL:          "/register?foo=bar",
		UseCookies:           true,
		SetAuthCookie:        func(w http.ResponseWriter, r *http.Request, token string) {},
	}

	req := httptest.NewRequest(http.MethodGet, "/auth/api/authknight/callback?once=test-token", nil)
	recorder := httptest.NewRecorder()

	ApiAuthKnightCallback(recorder, req, deps)

	location := recorder.Header().Get("Location")
	parsed, err := url.Parse(location)
	if err != nil {
		t.Fatalf("failed to parse redirect URL: %v", err)
	}
	if parsed.Path != "/register" {
		t.Fatalf("expected path /register, got %s", parsed.Path)
	}
	if parsed.Query().Get("foo") != "bar" {
		t.Fatalf("expected foo=bar in query, got %s", parsed.Query().Get("foo"))
	}
	if parsed.Query().Get("ak_key") == "" {
		t.Fatal("expected ak_key in query")
	}
}

func TestVerifyOnceToken_Success(t *testing.T) {
	server := mockAuthKnightServer("test@example.com", "success", http.StatusOK)
	defer server.Close()

	email, err := verifyOnceToken(context.Background(), server.URL, "test-token", 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if email != "test@example.com" {
		t.Fatalf("expected email 'test@example.com', got '%s'", email)
	}
}

func TestVerifyOnceToken_FailureStatus(t *testing.T) {
	server := mockAuthKnightServer("", "error", http.StatusOK)
	defer server.Close()

	_, err := verifyOnceToken(context.Background(), server.URL, "test-token", 5*time.Second)
	if err == nil {
		t.Fatal("expected error for failure status")
	}
}

func TestVerifyOnceToken_Unreachable(t *testing.T) {
	_, err := verifyOnceToken(context.Background(), "http://127.0.0.1:1", "test-token", 1*time.Second)
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}

func TestApiAuthKnightCallback_ConcurrentRequests(t *testing.T) {
	server := mockAuthKnightServer("user@example.com", "success", http.StatusOK)
	defer server.Close()

	var mu sync.Mutex
	tokenCount := 0

	deps := Dependencies{
		BaseURL:   server.URL,
		HTTPTimeout: 5 * time.Second,
		UserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
			return "user-123", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
			mu.Lock()
			tokenCount++
			mu.Unlock()
			return nil
		},
		TemporaryKeySet: func(key, value string, expiresSeconds int) error { return nil },
		TemporaryKeyGet: func(key string) (string, error) { return "", nil },
		UrlRedirectOnSuccess: "/dashboard",
		RegisterURL:          "/register",
		UseCookies:           true,
		SetAuthCookie:        func(w http.ResponseWriter, r *http.Request, token string) {},
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/auth/api/authknight/callback?once=test-token", nil)
			recorder := httptest.NewRecorder()
			ApiAuthKnightCallback(recorder, req, deps)
		}()
	}
	wg.Wait()

	mu.Lock()
	if tokenCount != 5 {
		t.Fatalf("expected 5 tokens stored, got %d", tokenCount)
	}
	mu.Unlock()
}
