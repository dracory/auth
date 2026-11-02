package api_login

import (
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

func TestApiLoginWithAuth_UsernamePassword(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	testutils.SetOtpForTest(a, false)

	recorder, req := makePostRequest(t, "/api/login", url.Values{})
	ApiLoginWithAuth(recorder, req, a)

	if !strings.Contains(recorder.Body.String(), `"status":`) {
		t.Fatalf("expected JSON response, got %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestApiLoginWithAuth_Otp(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	testutils.SetOtpForTest(a, true)

	recorder, req := makePostRequest(t, "/api/login", url.Values{
		"email": {"test@example.com"},
	})
	ApiLoginWithAuth(recorder, req, a)

	if !strings.Contains(recorder.Body.String(), `"status":`) {
		t.Fatalf("expected JSON response, got %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestApiLoginWithAuth_NonPasswordImplReturns500(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/login", nil)
	ApiLoginWithAuth(recorder, req, sharedOnlyAuth{})

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for non-password impl, got %d", recorder.Code)
	}
}
