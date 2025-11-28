package page_register_code_verify

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPageRegisterCodeVerify(t *testing.T) {
	deps := Dependencies{
		Endpoint:          "http://localhost/auth",
		RedirectOnSuccess: "http://localhost/dashboard",
		Layout:            func(content string) string { return content },
		Logger:            slog.Default(),
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	PageRegisterCodeVerify(recorder, req, deps)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()

	expected := []string{
		"Registration Code Verification",
		"Verification code",
		"var urlApiRegisterCodeVerify = \"http://localhost/auth/api/register-code-verify\";",
		"var urlOnSuccess = \"http://localhost/dashboard\";",
	}

	for _, v := range expected {
		if !strings.Contains(body, v) {
			t.Errorf("Handler returned unexpected result.\nEXPECTED: %s\nFOUND: %s", v, body)
		}
	}
}
