package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// helper to build a POST request with form values
func makePostRequest(t *testing.T, path string, values url.Values) (*httptest.ResponseRecorder, *http.Request) {
	body := strings.NewReader(values.Encode())
	req, err := http.NewRequest(http.MethodPost, path, body)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	return recorder, req
}

func TestApiLogin_UsernameAndPassword_Smoke(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatal(err)
	}

	values := url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}

	recorder, req := makePostRequest(t, authInstance.LinkApiLogin(), values)
	authInstance.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":") {
		t.Fatalf("expected JSON status in response, got %q", body)
	}
}

func TestApiLogin_Passwordless_Smoke(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	if err != nil {
		t.Fatal(err)
	}

	values := url.Values{
		"email": {"test@test.com"},
	}

	recorder, req := makePostRequest(t, authInstance.LinkApiLogin(), values)
	authInstance.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":") {
		t.Fatalf("expected JSON status in response, got %q", body)
	}
}

func TestApiRegister_UsernameAndPassword_Smoke(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatal(err)
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
		"password":   {"1234"},
	}

	recorder, req := makePostRequest(t, authInstance.LinkApiRegister(), values)
	authInstance.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":") {
		t.Fatalf("expected JSON status in response, got %q", body)
	}
}

func TestApiRegister_Passwordless_Smoke(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	if err != nil {
		t.Fatal(err)
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}

	recorder, req := makePostRequest(t, authInstance.LinkApiRegister(), values)
	authInstance.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":") {
		t.Fatalf("expected JSON status in response, got %q", body)
	}
}

func TestApiLogout_Smoke(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, authInstance.LinkApiLogout(), nil)

	authInstance.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"success\"") {
		t.Fatalf("expected success status in response, got %q", body)
	}
	if !strings.Contains(body, "\"message\":\"logout success\"") {
		t.Fatalf("expected logout success message in response, got %q", body)
	}
}

func TestApiPasswordRestore_Smoke(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatal(err)
	}

	values := url.Values{
		"email":      {"test@test.com"},
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}

	recorder, req := makePostRequest(t, authInstance.LinkApiPasswordRestore(), values)
	authInstance.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":") {
		t.Fatalf("expected JSON status in response, got %q", body)
	}
}

func TestApiPasswordReset_Smoke(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	if err != nil {
		t.Fatal(err)
	}

	values := url.Values{
		"token":            {"valid-token"},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}

	recorder, req := makePostRequest(t, authInstance.LinkApiPasswordReset(), values)
	authInstance.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":") {
		t.Fatalf("expected JSON status in response, got %q", body)
	}
}

func TestApiLoginCodeVerify_Smoke(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	if err != nil {
		t.Fatal(err)
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	recorder, req := makePostRequest(t, authInstance.LinkApiLoginCodeVerify(), values)
	authInstance.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":") {
		t.Fatalf("expected JSON status in response, got %q", body)
	}
}

func TestApiRegisterCodeVerify_Smoke(t *testing.T) {
	authInstance, err := testSetupPasswordlessAuth()
	if err != nil {
		t.Fatal(err)
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	recorder, req := makePostRequest(t, authInstance.LinkApiRegisterCodeVerify(), values)
	authInstance.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":") {
		t.Fatalf("expected JSON status in response, got %q", body)
	}
}
