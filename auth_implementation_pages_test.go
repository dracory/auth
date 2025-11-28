package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPageLogin_UsernameAndPassword(t *testing.T) {
	auth, errAuth := testSetupUsernameAndPasswordAuth()
	if errAuth != nil {
		t.Fatal(errAuth)
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.pageLogin)
	handler.ServeHTTP(recorder, req)

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
	auth, errAuth := testSetupPasswordlessAuth()
	if errAuth != nil {
		t.Fatal(errAuth)
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(auth.pageLogin)
	handler.ServeHTTP(recorder, req)

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

func TestPageRegister_UsernameAndPassword(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(authInstance.pageRegister)
	handler.ServeHTTP(recorder, req)

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
	authInstance, err := testSetupPasswordlessAuth()
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(authInstance.pageRegister)
	handler.ServeHTTP(recorder, req)

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

func TestPageRegisterCodeVerify(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(authInstance.pageRegisterCodeVerify)
	handler.ServeHTTP(recorder, req)

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

func TestPageLogout(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(authInstance.pageLogout)
	handler.ServeHTTP(recorder, req)

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

func TestPagePasswordRestore(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(authInstance.pagePasswordRestore)
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()

	expected := []string{
		"Restore password",
		"name=\"first_name\"",
		"name=\"last_name\"",
		"name=\"email\"",
		"var urlApiPasswordRestore = \"http://localhost/auth/api/restore-password\";",
		"var urlOnSuccess = \"http://localhost/auth/login\";",
	}

	for _, v := range expected {
		if !strings.Contains(body, v) {
			t.Errorf("Handler returned unexpected result.\nEXPECTED: %s\nFOUND: %s", v, body)
		}
	}
}

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

func TestPageLoginCodeVerify(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(authInstance.pageLoginCodeVerify)
	handler.ServeHTTP(recorder, req)

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
