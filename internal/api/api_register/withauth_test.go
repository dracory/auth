package api_register

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth/internal/testutils"
)

func TestApiRegisterWithAuth(t *testing.T) {
	a := testutils.NewAuthSharedForTest()

	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(url.Values{}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()

	ApiRegisterWithAuth(recorder, req, a)

	if !strings.Contains(recorder.Body.String(), `"status":`) {
		t.Fatalf("expected JSON response, got %d %q", recorder.Code, recorder.Body.String())
	}
}
