package api_impersonate_stop

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracory/auth/types"
)

type mockObservabilityHooks struct {
	impersonationStops []struct{ admin, target string }
}

func (m *mockObservabilityHooks) RecordLoginAttempt(string, bool, error)               {}
func (m *mockObservabilityHooks) RecordRegistrationAttempt(bool, error)                 {}
func (m *mockObservabilityHooks) RecordRateLimitHit(string, string)                     {}
func (m *mockObservabilityHooks) RecordSessionCreated(string)                           {}
func (m *mockObservabilityHooks) RecordPasswordReset(bool, error)                       {}
func (m *mockObservabilityHooks) RecordImpersonationStart(string, string)               {}
func (m *mockObservabilityHooks) RecordImpersonationStop(admin, target string)          { m.impersonationStops = append(m.impersonationStops, struct{ admin, target string }{admin, target}) }

func makeStopRequestWithToken(t *testing.T, authToken string, useCookies bool) (*httptest.ResponseRecorder, *http.Request) {
	req, err := http.NewRequest("POST", "/api/impersonate/stop", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	if useCookies {
		req.AddCookie(&http.Cookie{Name: types.CookieName, Value: authToken})
	} else {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	return httptest.NewRecorder(), req
}

func TestApiImpersonateStopEmptyToken(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet:     func(key string) (string, error) { return "", nil },
		TemporaryKeySet:     func(key string, value string, expires int) error { return nil },
		UserFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) { return "", nil },
	}

	recorder, req := makeStopRequestWithToken(t, "", false)
	ApiImpersonateStop(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "not currently impersonating") {
		t.Fatalf("expected not impersonating message, got %q", body)
	}
}

func TestApiImpersonateStopNotImpersonating(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet:     func(key string) (string, error) { return "", nil },
		TemporaryKeySet:     func(key string, value string, expires int) error { return nil },
		UserFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) { return "", nil },
	}

	recorder, req := makeStopRequestWithToken(t, "some-token", false)
	ApiImpersonateStop(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "not currently impersonating") {
		t.Fatalf("expected not impersonating message, got %q", body)
	}
}

func TestApiImpersonateStopTemporaryKeyGetFails(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet:     func(key string) (string, error) { return "", errors.New("redis down") },
		TemporaryKeySet:     func(key string, value string, expires int) error { return nil },
		UserFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) { return "", nil },
	}

	recorder, req := makeStopRequestWithToken(t, "some-token", false)
	ApiImpersonateStop(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"error"`) {
		t.Fatalf("expected error status, got %q", body)
	}
}

func TestApiImpersonateStopUserFindByAuthTokenFails(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			if strings.HasPrefix(key, "imp:") {
				return "original-admin-token", nil
			}
			return "", nil
		},
		TemporaryKeySet:     func(key string, value string, expires int) error { return nil },
		UserFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) { return "", errors.New("db error") },
	}

	recorder, req := makeStopRequestWithToken(t, "impersonation-token", false)
	ApiImpersonateStop(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"error"`) {
		t.Fatalf("expected error status, got %q", body)
	}
}

func TestApiImpersonateStopHappyPath(t *testing.T) {
	hooks := &mockObservabilityHooks{}
	var deletedKey string
	var deletedValue string
	var deletedExpires int
	var cookieSet bool
	var loggedOutUserID string

	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			if key == "imp:impersonation-token" {
				return "original-admin-token", nil
			}
			return "", nil
		},
		TemporaryKeySet: func(key string, value string, expires int) error {
			deletedKey = key
			deletedValue = value
			deletedExpires = expires
			return nil
		},
		UserFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
			if token == "impersonation-token" {
				return "target-456", nil
			}
			if token == "original-admin-token" {
				return "admin-123", nil
			}
			return "", nil
		},
		UserLogout: func(ctx context.Context, userID string, opts types.UserAuthOptions) error {
			loggedOutUserID = userID
			return nil
		},
		UseCookies: true,
		SetAuthCookie: func(w http.ResponseWriter, r *http.Request, token string) {
			cookieSet = true
			if token != "original-admin-token" {
				t.Errorf("expected cookie to be set to original-admin-token, got %s", token)
			}
		},
		ObservabilityHooks: hooks,
		ImpersonationStop: func(ctx context.Context, admin, target string) error {
			if admin != "admin-123" || target != "target-456" {
				t.Errorf("unexpected admin/target: %s/%s", admin, target)
			}
			return nil
		},
	}

	recorder, req := makeStopRequestWithToken(t, "impersonation-token", true)
	ApiImpersonateStop(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"success"`) {
		t.Fatalf("expected success status, got %q", body)
	}
	if !strings.Contains(body, "impersonation stopped") {
		t.Fatalf("expected stopped message, got %q", body)
	}
	if !cookieSet {
		t.Fatal("expected auth cookie to be restored")
	}
	if loggedOutUserID != "target-456" {
		t.Fatalf("expected logout for target-456, got %q", loggedOutUserID)
	}
	if deletedKey != "imp:impersonation-token" {
		t.Fatalf("expected deleted key to be imp:impersonation-token, got %q", deletedKey)
	}
	if deletedValue != "" {
		t.Fatalf("expected deleted value to be empty, got %q", deletedValue)
	}
	if deletedExpires != 1 {
		t.Fatalf("expected expires to be 1, got %d", deletedExpires)
	}
	if len(hooks.impersonationStops) != 1 || hooks.impersonationStops[0].admin != "admin-123" || hooks.impersonationStops[0].target != "target-456" {
		t.Fatalf("expected 1 stop hook with admin-123/target-456, got %+v", hooks.impersonationStops)
	}
}
