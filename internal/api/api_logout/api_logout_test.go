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

func TestLogoutError_Error(t *testing.T) {
	var e *LogoutError
	if e.Error() != "" {
		t.Fatalf("expected empty string for nil receiver, got %q", e.Error())
	}

	e = &LogoutError{Code: LogoutErrorCodeTokenLookup}
	if e.Error() != "token_lookup" {
		t.Fatalf("expected 'token_lookup', got %q", e.Error())
	}

	e = &LogoutError{Err: errors.New("some error")}
	if e.Error() != "some error" {
		t.Fatalf("expected 'some error', got %q", e.Error())
	}
}

func TestApiLogoutNilAuthTokenRetrieve(t *testing.T) {
	deps := Dependencies{
		UseCookies:        false,
		AuthTokenRetrieve: nil,
	}
	recorder, req := makePostRequest(t, "/api/logout")
	ApiLogout(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"error\"") {
		t.Fatalf("expected status error, got %q", body)
	}
	if !strings.Contains(body, "Internal server error") {
		t.Fatalf("expected internal server error, got %q", body)
	}
}

func TestApiLogoutSuccessWithCookies(t *testing.T) {
	cookieRemoved := false
	deps := Dependencies{
		UseCookies: true,
		AuthTokenRetrieve: func(r *http.Request, useCookies bool) string {
			return "some-token"
		},
		UserFromToken: func(ctx context.Context, token string) (string, error) {
			return "user-1", nil
		},
		LogoutUser: func(ctx context.Context, userID string) error {
			return nil
		},
		RemoveAuthCookie: func(w http.ResponseWriter, r *http.Request) {
			cookieRemoved = true
		},
	}
	recorder, req := makePostRequest(t, "/api/logout")
	ApiLogout(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"success\"") {
		t.Fatalf("expected success, got %q", body)
	}
	if !cookieRemoved {
		t.Fatal("expected RemoveAuthCookie to be called")
	}
}

func TestApiLogoutSuccessWithoutCookies(t *testing.T) {
	cookieRemoved := false
	deps := Dependencies{
		UseCookies: false,
		AuthTokenRetrieve: func(r *http.Request, useCookies bool) string {
			return "some-token"
		},
		UserFromToken: func(ctx context.Context, token string) (string, error) {
			return "user-1", nil
		},
		LogoutUser: func(ctx context.Context, userID string) error {
			return nil
		},
		RemoveAuthCookie: func(w http.ResponseWriter, r *http.Request) {
			cookieRemoved = true
		},
	}
	recorder, req := makePostRequest(t, "/api/logout")
	ApiLogout(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"success\"") {
		t.Fatalf("expected success, got %q", body)
	}
	if cookieRemoved {
		t.Fatal("expected RemoveAuthCookie NOT to be called when UseCookies is false")
	}
}

func TestApiLogoutLogoutUserError(t *testing.T) {
	deps := Dependencies{
		UseCookies: false,
		AuthTokenRetrieve: func(r *http.Request, useCookies bool) string {
			return "some-token"
		},
		UserFromToken: func(ctx context.Context, token string) (string, error) {
			return "user-1", nil
		},
		LogoutUser: func(ctx context.Context, userID string) error {
			return errors.New("logout db error")
		},
	}
	recorder, req := makePostRequest(t, "/api/logout")
	ApiLogout(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "Logout failed") {
		t.Fatalf("expected logout failed, got %q", body)
	}
}

func TestApiLogoutUserFromTokenNil(t *testing.T) {
	deps := Dependencies{
		UseCookies: false,
		AuthTokenRetrieve: func(r *http.Request, useCookies bool) string {
			return "some-token"
		},
		UserFromToken: nil,
	}
	recorder, req := makePostRequest(t, "/api/logout")
	ApiLogout(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "Logout failed") {
		t.Fatalf("expected logout failed, got %q", body)
	}
}

func TestApiLogoutLogoutUserNil(t *testing.T) {
	deps := Dependencies{
		UseCookies: false,
		AuthTokenRetrieve: func(r *http.Request, useCookies bool) string {
			return "some-token"
		},
		UserFromToken: func(ctx context.Context, token string) (string, error) {
			return "user-1", nil
		},
		LogoutUser: nil,
	}
	recorder, req := makePostRequest(t, "/api/logout")
	ApiLogout(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "Logout failed") {
		t.Fatalf("expected logout failed, got %q", body)
	}
}

func TestApiLogoutEmptyUserIDSuccess(t *testing.T) {
	deps := Dependencies{
		UseCookies: false,
		AuthTokenRetrieve: func(r *http.Request, useCookies bool) string {
			return "some-token"
		},
		UserFromToken: func(ctx context.Context, token string) (string, error) {
			return "", nil
		},
		LogoutUser: func(ctx context.Context, userID string) error {
			t.Fatal("LogoutUser should not be called for empty userID")
			return nil
		},
	}
	recorder, req := makePostRequest(t, "/api/logout")
	ApiLogout(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"success\"") {
		t.Fatalf("expected success for empty userID, got %q", body)
	}
}
