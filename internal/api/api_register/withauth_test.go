package api_register

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/types"
)

// sharedOnlyAuth satisfies AuthSharedInterface (via embedding) but not
// AuthPasswordInterface — used to exercise the non-password guard.
type sharedOnlyAuth struct {
	types.AuthSharedInterface
}

func (a sharedOnlyAuth) GetLogger() *slog.Logger { return slog.Default() }

// profileAuth adds the optional profile-update hook used by the
// registration-completion branch.
type profileAuth struct {
	types.AuthPasswordInterface
	updated string
}

func (a *profileAuth) GetFuncUserProfileUpdate() func(ctx context.Context, userID, firstName, lastName string, options types.UserAuthOptions) error {
	return func(ctx context.Context, userID, firstName, lastName string, options types.UserAuthOptions) error {
		a.updated = userID
		return nil
	}
}

// akAuthStub adapts the shared test double to AuthAuthKnightInterface.
type akAuthStub struct {
	types.AuthSharedInterface
}

func (s akAuthStub) IsAuthKnight() bool                            { return true }
func (s akAuthStub) LinkAuthKnightLogin(b, n string) string        { return "" }
func (s akAuthStub) LinkAuthKnightRedirect(r *http.Request) string { return "" }
func (s akAuthStub) LinkAuthKnightCallback() string                { return "" }
func (s akAuthStub) GetAuthKnightHTTPTimeout() time.Duration       { return time.Second }

func (s akAuthStub) GetAuthKnightUserFindByEmail() func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
	return func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
		return "u1", nil
	}
}

func (s akAuthStub) GetAuthKnightUserRegister() func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (string, error) {
	return func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (string, error) {
		return "u2", nil
	}
}

func (s akAuthStub) GetAuthKnightRedirectURL() func(ctx context.Context, userID string) string {
	return func(ctx context.Context, userID string) string { return "/" }
}

func makeRegisterPost(t *testing.T, values url.Values) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return httptest.NewRecorder(), req
}

func TestApiRegisterWithAuth(t *testing.T) {
	a := testutils.NewAuthSharedForTest()

	recorder, req := makeRegisterPost(t, url.Values{})
	ApiRegisterWithAuth(recorder, req, a)

	if !strings.Contains(recorder.Body.String(), `"status":`) {
		t.Fatalf("expected JSON response, got %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestApiRegisterWithAuth_Passwordless(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	testutils.SetPasswordlessForTest(a, true)

	recorder, req := makeRegisterPost(t, url.Values{})
	ApiRegisterWithAuth(recorder, req, a)

	if !strings.Contains(recorder.Body.String(), `"status":`) {
		t.Fatalf("expected JSON response, got %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestApiRegisterWithAuth_AuthKnight(t *testing.T) {
	a := akAuthStub{AuthSharedInterface: testutils.NewAuthSharedForTest()}

	recorder, req := makeRegisterPost(t, url.Values{})
	ApiRegisterWithAuth(recorder, req, a)

	if !strings.Contains(recorder.Body.String(), `"status":`) {
		t.Fatalf("expected JSON response, got %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestApiRegisterWithAuth_ProfileUpdateHook(t *testing.T) {
	a := &profileAuth{AuthPasswordInterface: testutils.NewAuthSharedForTest().(types.AuthPasswordInterface)}

	recorder, req := makeRegisterPost(t, url.Values{})
	ApiRegisterWithAuth(recorder, req, a)

	if !strings.Contains(recorder.Body.String(), `"status":`) {
		t.Fatalf("expected JSON response, got %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestApiRegisterWithAuth_NonPasswordNonPasswordless(t *testing.T) {
	// A shared-only impl that is neither password nor passwordless nor
	// authknight reaches the 500 guard.
	a := sharedOnlyAuth{AuthSharedInterface: testutils.NewAuthSharedForTest()}

	recorder, req := makeRegisterPost(t, url.Values{})
	ApiRegisterWithAuth(recorder, req, a)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d %q", recorder.Code, recorder.Body.String())
	}
}
