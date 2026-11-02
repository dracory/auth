package api_login_code_verify

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth/internal/testutils"
)

func TestApiLoginCodeVerifyWithAuth(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	testutils.SetOtpForTest(a, true)

	body := strings.NewReader(url.Values{}.Encode())
	req := httptest.NewRequest(http.MethodPost, "/api/login-code-verify", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()

	ApiLoginCodeVerifyWithAuth(recorder, req, a)

	if !strings.Contains(recorder.Body.String(), `"status":`) {
		t.Fatalf("expected JSON response, got %d %q", recorder.Code, recorder.Body.String())
	}
}
