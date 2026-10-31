package api_password_restore

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/types"
)

// sharedOnlyAuth satisfies AuthSharedInterface (via embedding) but not
// AuthPasswordInterface — used to exercise the non-password guard.
type sharedOnlyAuth struct {
	types.AuthSharedInterface
}

func (a sharedOnlyAuth) GetLogger() *slog.Logger { return slog.Default() }

func makeRestorePost(t *testing.T, values url.Values) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/restore-password", strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return httptest.NewRecorder(), req
}

// passwordAuthStub wires the minimal funcs the restore dependencies
// require on top of the shared test double.
type passwordAuthStub struct {
	types.AuthSharedInterface
}

func (s passwordAuthStub) GetFuncTemporaryKeySet() func(key string, value string, expiresSeconds int) error {
	return func(key string, value string, expiresSeconds int) error { return nil }
}

func (s passwordAuthStub) GetFuncUserFindByUsername() func(ctx context.Context, username, firstName, lastName string, options types.UserAuthOptions) (string, error) {
	return func(ctx context.Context, username, firstName, lastName string, options types.UserAuthOptions) (string, error) {
		return "u1", nil
	}
}

func (s passwordAuthStub) GetFuncEmailTemplatePasswordRestore() func(ctx context.Context, userID, link string, options types.UserAuthOptions) string {
	return func(ctx context.Context, userID, link string, options types.UserAuthOptions) string {
		return "body"
	}
}

func (s passwordAuthStub) GetFuncEmailSend() func(ctx context.Context, userID, subject, body string) error {
	return func(ctx context.Context, userID, subject, body string) error { return nil }
}

// AuthPasswordInterface markers — the embedded shared interface does not
// promote these, so they must be redeclared.
func (s passwordAuthStub) LinkPasswordRestore() string           { return "/restore" }
func (s passwordAuthStub) LinkPasswordReset(token string) string { return "/reset?token=" + token }
func (s passwordAuthStub) LinkApiPasswordRestore() string        { return "/api/restore" }
func (s passwordAuthStub) LinkApiPasswordReset() string          { return "/api/reset" }

func TestApiPasswordRestoreWithAuth(t *testing.T) {
	a := passwordAuthStub{AuthSharedInterface: testutils.NewAuthSharedForTest()}

	recorder, req := makeRestorePost(t, url.Values{
		"email": {"test@example.com"},
	})
	ApiPasswordRestoreWithAuth(recorder, req, a)

	if !strings.Contains(recorder.Body.String(), `"status":`) {
		t.Fatalf("expected JSON response, got %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestApiPasswordRestoreWithAuth_NonPasswordImplReturns500(t *testing.T) {
	recorder, req := makeRestorePost(t, url.Values{})
	ApiPasswordRestoreWithAuth(recorder, req, sharedOnlyAuth{})

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for non-password impl, got %d", recorder.Code)
	}
}
