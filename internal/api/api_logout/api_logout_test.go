package api_logout

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// helper to build a POST request without body
func makePostRequest(t *testing.T, path string) (*httptest.ResponseRecorder, *http.Request) {
	req, err := http.NewRequest("POST", path, nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	return recorder, req
}

func TestApiLogoutNoToken(t *testing.T) {
	// No cookie or header set, so token retrieval should fail/return empty
	deps := Dependencies{
		UseCookies: false,
		AuthTokenRetrieve: func(r *http.Request, useCookies bool) string {
			return ""
		},
		UserFromToken: func(ctx context.Context, token string) (string, error) {
			if token != "" {
				t.Fatalf("expected empty token, got %q", token)
			}
			// Mirror original behaviour: empty token simply yields no user
			return "", nil
		},
		LogoutUser: func(ctx context.Context, userID string) error {
			// Should not be called when userID is empty
			t.Fatalf("LogoutUser should not be called when userID is empty")
			return nil
		},
	}

	recorder, req := makePostRequest(t, "/api/logout")
	ApiLogout(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"success\"") {
		t.Fatalf("expected status success, got %q", body)
	}
	if !strings.Contains(body, "\"message\":\"logout success\"") {
		t.Fatalf("expected logout success message, got %q", body)
	}
}

func TestApiLogoutTokenValidationError(t *testing.T) {
	deps := Dependencies{
		UseCookies: false,
		AuthTokenRetrieve: func(r *http.Request, useCookies bool) string {
			// Simulate token present via header/cookie
			return "some-token"
		},
		UserFromToken: func(ctx context.Context, token string) (string, error) {
			if token != "some-token" {
				t.Fatalf("expected token 'some-token', got %q", token)
			}
			// Simulate token validation error (e.g. DB error)
			return "", errors.New("db error")
		},
		LogoutUser: func(ctx context.Context, userID string) error {
			// Should not be called when token validation fails
			t.Fatalf("LogoutUser should not be called when token validation fails")
			return nil
		},
	}

	recorder, req := makePostRequest(t, "/api/logout")
	ApiLogout(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Logout failed. Please try again later\"") {
		t.Fatalf("expected logout failed message, got %q", body)
	}
}
