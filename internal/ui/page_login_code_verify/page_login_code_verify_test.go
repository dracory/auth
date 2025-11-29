package page_login_code_verify

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracory/auth/internal/testutils"
)

func TestPageLoginCodeVerify(t *testing.T) {
	a := testutils.NewAuthSharedForTest()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	PageLoginCodeVerify(recorder, req, a)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()

	expected := []string{
		"Login Code Verification",
		"We sent you a login code to your email. Please check your mailbox",
		"var urlApiLoginCodeVerify = \"http://localhost/auth/api/login-code-verify\";",
		"var urlOnSuccess = \"http://localhost/dashboard\";",
	}

	for _, v := range expected {
		if !strings.Contains(body, v) {
			t.Errorf("Handler returned unexpected result.\nEXPECTED: %s\nFOUND: %s", v, body)
		}
	}
}
