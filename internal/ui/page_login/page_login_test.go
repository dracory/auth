package page_login

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPageLogin_UsernameAndPassword(t *testing.T) {
	deps := Dependencies{
		Passwordless:       false,
		EnableRegistration: true,
		Endpoint:           "http://localhost/auth",
		RedirectOnSuccess:  "http://localhost/dashboard",
		Layout: func(content string) string {
			return content
		},
		Logger: slog.Default(),
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	PageLogin(recorder, req, deps)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()

	expected := []string{
		`<span>Log in</span>`,
		`var urlApiLogin = "http://localhost/auth/api/login";`,
		`var urlOnSuccess = "http://localhost/dashboard";`,
	}

	for _, v := range expected {
		if !strings.Contains(body, v) {
			t.Errorf("Handler returned unexpected result.\nEXPECTED: %s\nFOUND: %s", v, body)
		}
	}
}

func TestPageLogin_Passwordless(t *testing.T) {
	deps := Dependencies{
		Passwordless:       true,
		EnableRegistration: true,
		Endpoint:           "http://localhost/auth",
		RedirectOnSuccess:  "http://localhost/dashboard", // unused in passwordless branch
		Layout: func(content string) string {
			return content
		},
		Logger: slog.Default(),
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	PageLogin(recorder, req, deps)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()

	expected := []string{
		`<span>Send me a login code</span>`,
		`var urlApiLogin = "http://localhost/auth/api/login";`,
		`var urlOnSuccess = "http://localhost/auth/login-code-verify";`,
	}

	for _, v := range expected {
		if !strings.Contains(body, v) {
			t.Errorf("Handler returned unexpected result.\nEXPECTED: %s\nFOUND: %s", v, body)
		}
	}
}
