package page_register

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracory/auth/internal/testutils"
)

func TestPageRegister_UsernameAndPassword(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	// Username/password branch with verification disabled.
	testutils.SetPasswordlessForTest(a, false)
	testutils.SetVerificationForTest(a, false)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	PageRegister(recorder, req, a)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()

	expected := []string{
		"Register",
		"name=\"first_name\"",
		"name=\"last_name\"",
		"name=\"email\"",
		"name=\"password\"",
		"var urlApiRegister = \"http://localhost/auth/api/register\";",
		"var urlOnSuccess = \"http://localhost/auth/login\";",
	}

	for _, v := range expected {
		if !strings.Contains(body, v) {
			t.Errorf("Handler returned unexpected result.\nEXPECTED: %s\nFOUND: %s", v, body)
		}
	}
}

func TestPageRegister_Passwordless(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	testutils.SetPasswordlessForTest(a, true)
	testutils.SetVerificationForTest(a, true)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	PageRegister(recorder, req, a)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()

	expected := []string{
		"Register",
		"name=\"first_name\"",
		"name=\"last_name\"",
		"name=\"email\"",
		"var urlApiRegister = \"http://localhost/auth/api/register\";",
		"var urlOnSuccess = \"http://localhost/auth/register-code-verify\";",
	}

	for _, v := range expected {
		if !strings.Contains(body, v) {
			t.Errorf("Handler returned unexpected result.\nEXPECTED: %s\nFOUND: %s", v, body)
		}
	}
}

func TestPageRegister_AuthKnight(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	testutils.SetPasswordlessForTest(a, false)
	testutils.SetVerificationForTest(a, false)
	testutils.SetAuthKnightForTest(a, true)
	testutils.SetTemporaryKeyGetForTest(a, func(key string) (string, error) {
		if key == "test-ak-key" {
			return "verified@example.com", nil
		}
		return "", nil
	})

	req, err := http.NewRequest("GET", "/register?ak_key=test-ak-key", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	PageRegister(recorder, req, a)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()

	expected := []string{
		"Register",
		"name=\"first_name\"",
		"name=\"last_name\"",
		"name=\"email\"",
		"verified@example.com",
		"name=\"ak_key\"",
		"form-control-plaintext",
		"var urlApiRegister = \"http://localhost/auth/api/register\";",
	}

	for _, v := range expected {
		if !strings.Contains(body, v) {
			t.Errorf("Handler returned unexpected result.\nEXPECTED: %s\nFOUND: %s", v, body)
		}
	}
}

func TestPageRegister_AuthKnightNoAkKey(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	testutils.SetPasswordlessForTest(a, false)
	testutils.SetVerificationForTest(a, false)
	testutils.SetAuthKnightForTest(a, true)

	req, err := http.NewRequest("GET", "/register", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "localhost:8084"

	recorder := httptest.NewRecorder()
	PageRegister(recorder, req, a)

	if status := recorder.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusTemporaryRedirect)
	}

	location := recorder.Header().Get("Location")
	if !strings.Contains(location, "authknight.com/app/login") {
		t.Errorf("expected redirect to authknight.com, got: %s", location)
	}
}
