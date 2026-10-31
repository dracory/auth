package api_logout

import (
	"strings"
	"testing"

	"github.com/dracory/auth/internal/testutils"
)

func TestApiLogoutWithAuth(t *testing.T) {
	a := testutils.NewAuthSharedForTest()

	recorder, req := makePostRequest(t, "/api/logout")
	ApiLogoutWithAuth(recorder, req, a)

	if !strings.Contains(recorder.Body.String(), `"status":`) {
		t.Fatalf("expected JSON response, got %d %q", recorder.Code, recorder.Body.String())
	}
}
