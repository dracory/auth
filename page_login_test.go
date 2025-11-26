package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPageLogin(t *testing.T) {
	auth, errAuth := testSetupUsernameAndPasswordAuth()

	if errAuth != nil {
		t.Fatal(errAuth)
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.pageLogin)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusSeeOther)
	}

	expected := []string{
		`<span>Log in</span>`,
		`var urlApiLogin = "http://localhost/auth/api/login";`,
		`var urlOnSuccess = "http://localhost/dashboard";`,
	}

	for _, v := range expected {
		if !strings.Contains(rr.Body.String(), v) {
			t.Error("Handler returned unexpected result.\nEXPECTED:", v, "\nFOUND:", rr.Body.String())
		}
	}
}

func TestPageLoginPasswordless(t *testing.T) {
	auth, errAuth := testSetupPasswordlessAuth()

	if errAuth != nil {
		t.Fatal(errAuth)
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.pageLogin)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusSeeOther)
	}

	expected := []string{
		`<span>Send me a login code</span>`,
		`var urlApiLogin = "http://localhost/auth/api/login";`,
		`var urlOnSuccess = "http://localhost/auth/login-code-verify";`,
	}

	for _, v := range expected {
		if !strings.Contains(rr.Body.String(), v) {
			t.Error("Handler returned unexpected result.\nEXPECTED:", v, "\nFOUND:", rr.Body.String())
		}
	}
}
