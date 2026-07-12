package api_impersonate_start

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth/types"
)

type mockObservabilityHooks struct {
	impersonationStarts []struct{ admin, target string }
	sessionsCreated     []string
}

func (m *mockObservabilityHooks) RecordLoginAttempt(string, bool, error) {}
func (m *mockObservabilityHooks) RecordRegistrationAttempt(bool, error)  {}
func (m *mockObservabilityHooks) RecordRateLimitHit(string, string)      {}
func (m *mockObservabilityHooks) RecordSessionCreated(userID string) {
	m.sessionsCreated = append(m.sessionsCreated, userID)
}
func (m *mockObservabilityHooks) RecordPasswordReset(bool, error) {}
func (m *mockObservabilityHooks) RecordImpersonationStart(admin, target string) {
	m.impersonationStarts = append(m.impersonationStarts, struct{ admin, target string }{admin, target})
}
func (m *mockObservabilityHooks) RecordImpersonationStop(string, string) {}

func makePostRequestWithToken(t *testing.T, path, authToken string, values url.Values, useCookies bool) (*httptest.ResponseRecorder, *http.Request) {
	body := strings.NewReader(values.Encode())
	req, err := http.NewRequest("POST", path, body)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if useCookies {
		req.AddCookie(&http.Cookie{Name: types.CookieName, Value: authToken})
	} else {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	return httptest.NewRecorder(), req
}

func TestApiImpersonateStartNoAuthToken(t *testing.T) {
	deps := Dependencies{
		CanImpersonate: func(ctx context.Context, admin, target string) (bool, error) { return true, nil },
		UserFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
			return "admin-123", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, opts types.UserAuthOptions) error { return nil },
		TemporaryKeyGet:    func(key string) (string, error) { return "", nil },
		TemporaryKeySet:    func(key string, value string, expires int) error { return nil },
	}

	values := url.Values{"user_id": {"target-456"}}
	recorder, req := makePostRequestWithToken(t, "/api/impersonate/start", "", values, false)
	ApiImpersonateStart(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"unauthenticated"`) {
		t.Fatalf("expected unauthenticated status, got %q", body)
	}
}

func TestApiImpersonateStartUserNotFound(t *testing.T) {
	deps := Dependencies{
		CanImpersonate: func(ctx context.Context, admin, target string) (bool, error) { return true, nil },
		UserFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
			return "", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, opts types.UserAuthOptions) error { return nil },
		TemporaryKeyGet:    func(key string) (string, error) { return "", nil },
		TemporaryKeySet:    func(key string, value string, expires int) error { return nil },
	}

	values := url.Values{"user_id": {"target-456"}}
	recorder, req := makePostRequestWithToken(t, "/api/impersonate/start", "admin-token", values, false)
	ApiImpersonateStart(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"unauthenticated"`) {
		t.Fatalf("expected unauthenticated status, got %q", body)
	}
}

func TestApiImpersonateStartMissingUserID(t *testing.T) {
	deps := Dependencies{
		CanImpersonate: func(ctx context.Context, admin, target string) (bool, error) { return true, nil },
		UserFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
			return "admin-123", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, opts types.UserAuthOptions) error { return nil },
		TemporaryKeyGet:    func(key string) (string, error) { return "", nil },
		TemporaryKeySet:    func(key string, value string, expires int) error { return nil },
	}

	values := url.Values{}
	recorder, req := makePostRequestWithToken(t, "/api/impersonate/start", "admin-token", values, false)
	ApiImpersonateStart(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"error"`) {
		t.Fatalf("expected error status, got %q", body)
	}
}

func TestApiImpersonateStartAlreadyImpersonating(t *testing.T) {
	deps := Dependencies{
		CanImpersonate: func(ctx context.Context, admin, target string) (bool, error) { return true, nil },
		UserFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
			return "admin-123", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, opts types.UserAuthOptions) error { return nil },
		TemporaryKeyGet: func(key string) (string, error) {
			if strings.HasPrefix(key, "imp:") {
				return "original-admin-token", nil
			}
			return "", nil
		},
		TemporaryKeySet: func(key string, value string, expires int) error { return nil },
	}

	values := url.Values{"user_id": {"target-456"}}
	recorder, req := makePostRequestWithToken(t, "/api/impersonate/start", "admin-token", values, false)
	ApiImpersonateStart(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "already impersonating") {
		t.Fatalf("expected already impersonating message, got %q", body)
	}
}

func TestApiImpersonateStartCanImpersonateFalse(t *testing.T) {
	deps := Dependencies{
		CanImpersonate: func(ctx context.Context, admin, target string) (bool, error) { return false, nil },
		UserFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
			return "admin-123", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, opts types.UserAuthOptions) error { return nil },
		TemporaryKeyGet:    func(key string) (string, error) { return "", nil },
		TemporaryKeySet:    func(key string, value string, expires int) error { return nil },
	}

	values := url.Values{"user_id": {"target-456"}}
	recorder, req := makePostRequestWithToken(t, "/api/impersonate/start", "admin-token", values, false)
	ApiImpersonateStart(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "not allowed") {
		t.Fatalf("expected forbidden message, got %q", body)
	}
}

func TestApiImpersonateStartCanImpersonateError(t *testing.T) {
	deps := Dependencies{
		CanImpersonate: func(ctx context.Context, admin, target string) (bool, error) { return false, errors.New("db error") },
		UserFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
			return "admin-123", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, opts types.UserAuthOptions) error { return nil },
		TemporaryKeyGet:    func(key string) (string, error) { return "", nil },
		TemporaryKeySet:    func(key string, value string, expires int) error { return nil },
	}

	values := url.Values{"user_id": {"target-456"}}
	recorder, req := makePostRequestWithToken(t, "/api/impersonate/start", "admin-token", values, false)
	ApiImpersonateStart(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"error"`) {
		t.Fatalf("expected error status, got %q", body)
	}
}

func TestApiImpersonateStartTemporaryKeySetFails(t *testing.T) {
	deps := Dependencies{
		CanImpersonate: func(ctx context.Context, admin, target string) (bool, error) { return true, nil },
		UserFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
			return "admin-123", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, opts types.UserAuthOptions) error { return nil },
		TemporaryKeyGet:    func(key string) (string, error) { return "", nil },
		TemporaryKeySet:    func(key string, value string, expires int) error { return errors.New("redis down") },
	}

	values := url.Values{"user_id": {"target-456"}}
	recorder, req := makePostRequestWithToken(t, "/api/impersonate/start", "admin-token", values, false)
	ApiImpersonateStart(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"error"`) {
		t.Fatalf("expected error status, got %q", body)
	}
}

func TestApiImpersonateStartUserStoreAuthTokenFails(t *testing.T) {
	deps := Dependencies{
		CanImpersonate: func(ctx context.Context, admin, target string) (bool, error) { return true, nil },
		UserFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
			return "admin-123", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, opts types.UserAuthOptions) error {
			return errors.New("db error")
		},
		TemporaryKeyGet: func(key string) (string, error) { return "", nil },
		TemporaryKeySet: func(key string, value string, expires int) error { return nil },
	}

	values := url.Values{"user_id": {"target-456"}}
	recorder, req := makePostRequestWithToken(t, "/api/impersonate/start", "admin-token", values, false)
	ApiImpersonateStart(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"error"`) {
		t.Fatalf("expected error status, got %q", body)
	}
}

func TestApiImpersonateStartHappyPath(t *testing.T) {
	hooks := &mockObservabilityHooks{}
	var storedKey, storedValue string
	var storedExpires int
	var storedToken, storedUserID string

	deps := Dependencies{
		CanImpersonate: func(ctx context.Context, admin, target string) (bool, error) { return true, nil },
		UserFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
			if token == "admin-token" {
				return "admin-123", nil
			}
			return "", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, opts types.UserAuthOptions) error {
			storedToken = token
			storedUserID = userID
			return nil
		},
		TemporaryKeyGet: func(key string) (string, error) { return "", nil },
		TemporaryKeySet: func(key string, value string, expires int) error {
			storedKey = key
			storedValue = value
			storedExpires = expires
			return nil
		},
		ObservabilityHooks: hooks,
		ImpersonationStart: func(ctx context.Context, admin, target string) error {
			if admin != "admin-123" || target != "target-456" {
				t.Errorf("unexpected admin/target: %s/%s", admin, target)
			}
			return nil
		},
	}

	values := url.Values{"user_id": {"target-456"}}
	recorder, req := makePostRequestWithToken(t, "/api/impersonate/start", "admin-token", values, false)
	ApiImpersonateStart(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"success"`) {
		t.Fatalf("expected success status, got %q", body)
	}
	if !strings.Contains(body, `"token"`) {
		t.Fatalf("expected token in response, got %q", body)
	}
	if !strings.HasPrefix(storedKey, "imp:") {
		t.Fatalf("expected stored key to have imp: prefix, got %q", storedKey)
	}
	if storedValue != "admin-token" {
		t.Fatalf("expected stored value to be admin-token, got %q", storedValue)
	}
	if storedUserID != "target-456" {
		t.Fatalf("expected stored userID to be target-456, got %q", storedUserID)
	}
	if storedToken == "" {
		t.Fatal("expected non-empty token to be stored")
	}
	if storedExpires != 7200 {
		t.Fatalf("expected expires to be 7200 (2h in seconds), got %d", storedExpires)
	}
	if len(hooks.impersonationStarts) != 1 || hooks.impersonationStarts[0].admin != "admin-123" || hooks.impersonationStarts[0].target != "target-456" {
		t.Fatalf("expected 1 impersonation start hook with admin-123/target-456, got %+v", hooks.impersonationStarts)
	}
	if len(hooks.sessionsCreated) != 1 || hooks.sessionsCreated[0] != "target-456" {
		t.Fatalf("expected 1 session created for target-456, got %+v", hooks.sessionsCreated)
	}
}

func TestApiImpersonateStartMissingUserIDMessage(t *testing.T) {
	deps := Dependencies{
		CanImpersonate: func(ctx context.Context, admin, target string) (bool, error) { return true, nil },
		UserFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
			return "admin-123", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, opts types.UserAuthOptions) error { return nil },
		TemporaryKeyGet:    func(key string) (string, error) { return "", nil },
		TemporaryKeySet:    func(key string, value string, expires int) error { return nil },
	}

	values := url.Values{}
	recorder, req := makePostRequestWithToken(t, "/api/impersonate/start", "admin-token", values, false)
	ApiImpersonateStart(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "user_id is required") {
		t.Fatalf("expected user_id required message, got %q", body)
	}
}

func TestApiImpersonateStartUserStoreAuthTokenFailCleansUpKey(t *testing.T) {
	var cleanupKey string
	var cleanupValue string
	var cleanupExpires int
	callCount := 0

	deps := Dependencies{
		CanImpersonate: func(ctx context.Context, admin, target string) (bool, error) { return true, nil },
		UserFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
			return "admin-123", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, opts types.UserAuthOptions) error {
			return errors.New("db error")
		},
		TemporaryKeyGet: func(key string) (string, error) { return "", nil },
		TemporaryKeySet: func(key string, value string, expires int) error {
			callCount++
			if callCount == 2 {
				cleanupKey = key
				cleanupValue = value
				cleanupExpires = expires
			}
			return nil
		},
	}

	values := url.Values{"user_id": {"target-456"}}
	recorder, req := makePostRequestWithToken(t, "/api/impersonate/start", "admin-token", values, false)
	ApiImpersonateStart(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"error"`) {
		t.Fatalf("expected error status, got %q", body)
	}
	if callCount != 2 {
		t.Fatalf("expected TemporaryKeySet to be called twice (store + cleanup), got %d", callCount)
	}
	if !strings.HasPrefix(cleanupKey, "imp:") {
		t.Fatalf("expected cleanup key to have imp: prefix, got %q", cleanupKey)
	}
	if cleanupValue != "" {
		t.Fatalf("expected cleanup value to be empty, got %q", cleanupValue)
	}
	if cleanupExpires != 1 {
		t.Fatalf("expected cleanup expires to be 1, got %d", cleanupExpires)
	}
}

func TestApiImpersonateStartWithCookies(t *testing.T) {
	var cookieSet bool
	deps := Dependencies{
		CanImpersonate: func(ctx context.Context, admin, target string) (bool, error) { return true, nil },
		UserFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
			return "admin-123", nil
		},
		UserStoreAuthToken: func(ctx context.Context, token, userID string, opts types.UserAuthOptions) error { return nil },
		TemporaryKeyGet:    func(key string) (string, error) { return "", nil },
		TemporaryKeySet:    func(key string, value string, expires int) error { return nil },
		UseCookies:         true,
		SetAuthCookie: func(w http.ResponseWriter, r *http.Request, token string) {
			cookieSet = true
		},
	}

	values := url.Values{"user_id": {"target-456"}}
	recorder, req := makePostRequestWithToken(t, "/api/impersonate/start", "admin-token", values, true)
	ApiImpersonateStart(recorder, req, deps)

	if !cookieSet {
		t.Fatal("expected auth cookie to be set when UseCookies is true")
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"success"`) {
		t.Fatalf("expected success status, got %q", body)
	}
}
