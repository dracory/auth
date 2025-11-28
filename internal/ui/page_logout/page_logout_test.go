package page_logout

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPageLogout(t *testing.T) {
	deps := Dependencies{
		Endpoint: "http://localhost/auth",
		Layout:   func(content string) string { return content },
		Logger:   slog.Default(),
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	PageLogout(recorder, req, deps)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()

	expected := []string{
		"Sign out",
		"Logout",
		"var urlApiLogout = \"http://localhost/auth/api/logout\";",
		"var urlOnSuccess = \"http://localhost/auth/login\";",
	}

	for _, v := range expected {
		if !strings.Contains(body, v) {
			t.Errorf("Handler returned unexpected result.\nEXPECTED: %s\nFOUND: %s", v, body)
		}
	}
}
