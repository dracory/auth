package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth/internal/testutils"
)

func TestApiLogin_UsernameAndPassword_Smoke(t *testing.T) {
	config := testutils.NewUsernameAndPasswordConfigForTest()
	authShared, err := NewUsernameAndPasswordAuth(config)
	if err != nil {
		t.Fatal(err)
	}

	values := url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}

	recorder, req := testutils.MakePostRequest(t, authShared.LinkApiLogin(), values)
	authShared.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":") {
		t.Fatalf("expected JSON status in response, got %q", body)
	}
}

func TestApiLogin_Passwordless_Smoke(t *testing.T) {
	config := testutils.NewPasswordlessConfigForTest()
	authShared, err := NewPasswordlessAuth(config)
	if err != nil {
		t.Fatal(err)
	}

	values := url.Values{
		"email": {"test@test.com"},
	}

	recorder, req := testutils.MakePostRequest(t, authShared.LinkApiLogin(), values)
	authShared.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":") {
		t.Fatalf("expected JSON status in response, got %q", body)
	}
}

func TestApiRegister_UsernameAndPassword_Smoke(t *testing.T) {
	config := testutils.NewUsernameAndPasswordConfigForTest()
	authShared, err := NewUsernameAndPasswordAuth(config)
	if err != nil {
		t.Fatal(err)
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
		"password":   {"1234"},
	}

	recorder, req := testutils.MakePostRequest(t, authShared.LinkApiRegister(), values)
	authShared.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":") {
		t.Fatalf("expected JSON status in response, got %q", body)
	}
}

func TestApiRegister_Passwordless_Smoke(t *testing.T) {
	config := testutils.NewPasswordlessConfigForTest()
	authShared, err := NewPasswordlessAuth(config)
	if err != nil {
		t.Fatal(err)
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}

	recorder, req := testutils.MakePostRequest(t, authShared.LinkApiRegister(), values)
	authShared.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":") {
		t.Fatalf("expected JSON status in response, got %q", body)
	}
}

func TestApiLogout_Smoke(t *testing.T) {
	config := testutils.NewUsernameAndPasswordConfigForTest()
	authShared, err := NewUsernameAndPasswordAuth(config)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, authShared.LinkApiLogout(), nil)

	authShared.Router().ServeHTTP(recorder, req)

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
	config := testutils.NewUsernameAndPasswordConfigForTest()
	authShared, err := NewUsernameAndPasswordAuth(config)
	if err != nil {
		t.Fatal(err)
	}

	values := url.Values{
		"email":      {"test@test.com"},
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}

	recorder, req := testutils.MakePostRequest(t, authShared.LinkApiPasswordRestore(), values)
	authShared.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":") {
		t.Fatalf("expected JSON status in response, got %q", body)
	}
}

func TestApiPasswordReset_Smoke(t *testing.T) {
	config := testutils.NewUsernameAndPasswordConfigForTest()
	authShared, err := NewUsernameAndPasswordAuth(config)
	if err != nil {
		t.Fatal(err)
	}

	values := url.Values{
		"token":            {"valid-token"},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}

	recorder, req := testutils.MakePostRequest(t, authShared.LinkApiPasswordReset(), values)
	authShared.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":") {
		t.Fatalf("expected JSON status in response, got %q", body)
	}
}

func TestApiLoginCodeVerify_Smoke(t *testing.T) {
	config := testutils.NewPasswordlessConfigForTest()
	authShared, err := NewPasswordlessAuth(config)
	if err != nil {
		t.Fatal(err)
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	recorder, req := testutils.MakePostRequest(t, authShared.LinkApiLoginCodeVerify(), values)
	authShared.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":") {
		t.Fatalf("expected JSON status in response, got %q", body)
	}
}

func TestApiRegisterCodeVerify_Smoke(t *testing.T) {
	config := testutils.NewPasswordlessConfigForTest()
	authShared, err := NewPasswordlessAuth(config)
	if err != nil {
		t.Fatal(err)
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}

	recorder, req := testutils.MakePostRequest(t, authShared.LinkApiRegisterCodeVerify(), values)
	authShared.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":") {
		t.Fatalf("expected JSON status in response, got %q", body)
	}
}
