package page_login

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracory/auth/internal/testutils"
)

func TestPageLogin_UsernameAndPassword(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	testutils.SetRegistrationForTest(a, true)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	PageLogin(recorder, req, a)

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
	a := testutils.NewAuthSharedForTest()
	testutils.SetRegistrationForTest(a, true)
	testutils.SetPasswordlessForTest(a, true)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	PageLogin(recorder, req, a)

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

func TestPageLogin_AuthKnight(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	testutils.SetPasswordlessForTest(a, false)
	testutils.SetVerificationForTest(a, false)
	testutils.SetAuthKnightForTest(a, true)

	req, err := http.NewRequest("GET", "/login", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "localhost:8084"

	recorder := httptest.NewRecorder()
	PageLogin(recorder, req, a)

	if status := recorder.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusTemporaryRedirect)
	}

	location := recorder.Header().Get("Location")
	if !strings.Contains(location, "authknight.com/app/login") {
		t.Errorf("expected redirect to authknight.com, got: %s", location)
	}
	if !strings.Contains(location, "back_url=") {
		t.Errorf("expected back_url parameter in redirect, got: %s", location)
	}
	if !strings.Contains(location, "next_url=") {
		t.Errorf("expected next_url parameter in redirect, got: %s", location)
	}
	if !strings.Contains(location, "authknight/callback") {
		t.Errorf("expected next_url to contain callback path, got: %s", location)
	}
}
