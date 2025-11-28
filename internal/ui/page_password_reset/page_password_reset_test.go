package page_password_reset

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPagePasswordReset_ValidTokenShowsForm(t *testing.T) {
	deps := Dependencies{
		Endpoint:           "http://localhost/auth",
		EnableRegistration: true,
		Token:              "valid-token",
		ErrorMessage:       "",
		Layout:             func(content string) string { return content },
		Logger:             slog.Default(),
	}

	req, err := http.NewRequest("GET", "/?t=valid-token", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	PagePasswordReset(recorder, req, deps)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()

	expected := []string{
		"Reset Password",
		"name=\"password\"",
		"name=\"password_confirm\"",
		"var urlApiPasswordReset = \"http://localhost/auth/api/reset-password\";",
		"var urlOnSuccess = \"http://localhost/auth/login\";",
	}

	for _, v := range expected {
		if !strings.Contains(body, v) {
			t.Errorf("Handler returned unexpected result.\nEXPECTED: %s\nFOUND: %s", v, body)
		}
	}
}

func TestPagePasswordReset_MissingTokenShowsError(t *testing.T) {
	deps := Dependencies{
		Endpoint:           "http://localhost/auth",
		EnableRegistration: true,
		Token:              "",
		ErrorMessage:       "Link is invalid",
		Layout:             func(content string) string { return content },
		Logger:             slog.Default(),
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	PagePasswordReset(recorder, req, deps)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()

	if !strings.Contains(body, "Link is invalid") {
		t.Errorf("expected error message %q in body, got %s", "Link is invalid", body)
	}
}
