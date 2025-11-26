package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPagePasswordReset_ValidTokenShowsForm(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatal(err)
	}

	// Mock token lookup to succeed
	authInstance.funcTemporaryKeyGet = func(key string) (value string, err error) {
		if key == "valid-token" {
			return "user123", nil
		}
		return "", nil
	}

	req, err := http.NewRequest("GET", "/?t=valid-token", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(authInstance.pagePasswordReset)
	handler.ServeHTTP(recorder, req)

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
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatal(err)
	}

	// No token query parameter
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(authInstance.pagePasswordReset)
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()

	if !strings.Contains(body, "Link is invalid") {
		t.Errorf("expected error message %q in body, got %s", "Link is invalid", body)
	}
}
