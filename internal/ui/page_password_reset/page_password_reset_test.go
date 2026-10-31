package page_password_reset

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/utils"
)

func TestPagePasswordReset_ValidTokenShowsForm(t *testing.T) {
	a := testutils.NewAuthSharedForTest()

	req, err := http.NewRequest("GET", "/?t=valid-token", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	PagePasswordReset(recorder, req, a)

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

func TestPagePasswordReset_LooksUpNamespacedToken(t *testing.T) {
	a := testutils.NewAuthSharedForTest()

	var lookedUpKey string
	testutils.SetTemporaryKeyGetForTest(a, func(key string) (string, error) {
		lookedUpKey = key
		return "user123", nil
	})

	req, err := http.NewRequest("GET", "/?t=valid-token", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	PagePasswordReset(recorder, req, a)

	if lookedUpKey != utils.PasswordResetTokenKeyPrefix+"valid-token" {
		t.Fatalf("expected lookup of %q, got %q", utils.PasswordResetTokenKeyPrefix+"valid-token", lookedUpKey)
	}

	body := recorder.Body.String()
	if strings.Contains(body, "Link is invalid") {
		t.Errorf("expected no error message for valid token, got %s", body)
	}
}

func TestPagePasswordReset_MissingTokenShowsError(t *testing.T) {
	a := testutils.NewAuthSharedForTest()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	PagePasswordReset(recorder, req, a)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()

	if !strings.Contains(body, "Link is invalid") {
		t.Errorf("expected error message %q in body, got %s", "Link is invalid", body)
	}
}
