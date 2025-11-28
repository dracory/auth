package api_password_restore

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func makePostRequest(t *testing.T, path string, values url.Values) (*httptest.ResponseRecorder, *http.Request) {
	body := strings.NewReader(values.Encode())
	req, err := http.NewRequest("POST", path, body)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	return recorder, req
}

func TestApiPasswordRestoreRequiresEmail(t *testing.T) {
	deps := Dependencies{}

	recorder, req := makePostRequest(t, "/api/password-restore", url.Values{})
	ApiPasswordRestore(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"error\"") {
		t.Fatalf("expected status error, got %q", body)
	}
	if !strings.Contains(body, "\"message\":\"Email is required field\"") {
		t.Fatalf("expected email required message, got %q", body)
	}
}

func TestApiPasswordRestoreRequiresFirstName(t *testing.T) {
	deps := Dependencies{}

	values := url.Values{
		"email": {"test@test.com"},
	}
	recorder, req := makePostRequest(t, "/api/password-restore", values)
	ApiPasswordRestore(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"First name is required field\"") {
		t.Fatalf("expected first name required message, got %q", body)
	}
}

func TestApiPasswordRestoreRequiresLastName(t *testing.T) {
	deps := Dependencies{}

	values := url.Values{
		"email":      {"test@test.com"},
		"first_name": {"John"},
	}
	recorder, req := makePostRequest(t, "/api/password-restore", values)
	ApiPasswordRestore(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Last name is required field\"") {
		t.Fatalf("expected last name required message, got %q", body)
	}
}

func TestApiPasswordRestoreUserNotFound(t *testing.T) {
	deps := Dependencies{
		UserFindByUsername: func(ctx context.Context, email, firstName, lastName string) (string, error) {
			return "", nil
		},
	}

	values := url.Values{
		"email":      {"test@test.com"},
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}
	recorder, req := makePostRequest(t, "/api/password-restore", values)
	ApiPasswordRestore(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"User not found\"") {
		t.Fatalf("expected user not found message, got %q", body)
	}
}

func TestApiPasswordRestoreInternalError(t *testing.T) {
	deps := Dependencies{
		UserFindByUsername: func(ctx context.Context, email, firstName, lastName string) (string, error) {
			return "", errors.New("db error")
		},
	}

	values := url.Values{
		"email":      {"test@test.com"},
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}
	recorder, req := makePostRequest(t, "/api/password-restore", values)
	ApiPasswordRestore(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Internal server error\"") {
		t.Fatalf("expected internal server error message, got %q", body)
	}
}

func TestApiPasswordRestoreSuccess(t *testing.T) {
	userFound := false
	tempKeySetCalled := false
	emailSent := false

	deps := Dependencies{
		UserFindByUsername: func(ctx context.Context, email, firstName, lastName string) (string, error) {
			userFound = true
			return "user123", nil
		},
		TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
			tempKeySetCalled = true
			return nil
		},
		ExpiresSeconds: 3600,
		EmailTemplate: func(ctx context.Context, userID, token string) string {
			return "email-body"
		},
		EmailSend: func(ctx context.Context, userID, subject, body string) error {
			emailSent = true
			return nil
		},
	}

	values := url.Values{
		"email":      {"test@test.com"},
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}
	recorder, req := makePostRequest(t, "/api/password-restore", values)
	ApiPasswordRestore(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"success\"") {
		t.Fatalf("expected success status, got %q", body)
	}
	if !strings.Contains(body, "\"message\":\"Password reset link was sent to your e-mail\"") {
		t.Fatalf("expected success message, got %q", body)
	}

	if !userFound {
		t.Fatalf("UserFindByUsername should be called")
	}
	if !tempKeySetCalled {
		t.Fatalf("TemporaryKeySet should be called")
	}
	if !emailSent {
		t.Fatalf("EmailSend should be called")
	}
}
