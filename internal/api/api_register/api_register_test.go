package api_register

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

// Username and password registration tests

func TestApiRegisterUsernameAndPasswordRequiresFirstName(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		RegisterWithUsernameAndPassword: func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (string, string) {
			if firstName == "" {
				return "", "First name is required field"
			}
			return "", ""
		},
	}

	recorder, req := makePostRequest(t, "/api/register", url.Values{})
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"error\"") {
		t.Fatalf("expected status error, got %q", body)
	}
	if !strings.Contains(body, "\"message\":\"First name is required field\"") {
		t.Fatalf("expected first name required message, got %q", body)
	}
}

func TestApiRegisterUsernameAndPasswordRequiresLastName(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		RegisterWithUsernameAndPassword: func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (string, string) {
			if lastName == "" {
				return "", "Last name is required field"
			}
			return "", ""
		},
	}

	values := url.Values{
		"first_name": {"John"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Last name is required field\"") {
		t.Fatalf("expected last name required message, got %q", body)
	}
}

func TestApiRegisterUsernameAndPasswordRequiresEmail(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		RegisterWithUsernameAndPassword: func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (string, string) {
			if email == "" {
				return "", "Email is required field"
			}
			return "", ""
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Email is required field\"") {
		t.Fatalf("expected email required message, got %q", body)
	}
}

func TestApiRegisterUsernameAndPasswordRequiresPassword(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		RegisterWithUsernameAndPassword: func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (string, string) {
			if password == "" {
				return "", "Password is required field"
			}
			return "", ""
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Password is required field\"") {
		t.Fatalf("expected password required message, got %q", body)
	}
}

func TestApiRegisterUsernameAndPasswordInvalidEmail(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		RegisterWithUsernameAndPassword: func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (string, string) {
			if email == "invalid-email" {
				return "", "This is not a valid email: invalid-email"
			}
			return "", ""
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"invalid-email"},
		"password":   {"1234"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"This is not a valid email: invalid-email\"") {
		t.Fatalf("expected invalid email message, got %q", body)
	}
}

func TestApiRegisterUsernameAndPasswordFuncUserRegisterNotDefined(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		RegisterWithUsernameAndPassword: func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (string, string) {
			return "", "registration failed. FuncUserRegister function not defined"
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
		"password":   {"1234"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"registration failed. FuncUserRegister function not defined\"") {
		t.Fatalf("expected func not defined message, got %q", body)
	}
}

func TestApiRegisterUsernameAndPasswordRegistrationFailed(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		RegisterWithUsernameAndPassword: func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (string, string) {
			return "", "registration failed."
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
		"password":   {"1234"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"registration failed.\"") {
		t.Fatalf("expected registration failed message, got %q", body)
	}
}

func TestApiRegisterUsernameAndPasswordSuccess(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		RegisterWithUsernameAndPassword: func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (string, string) {
			return "registration success", ""
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
		"password":   {"1234"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"success\"") {
		t.Fatalf("expected success status, got %q", body)
	}
	if !strings.Contains(body, "\"message\":\"registration success\"") {
		t.Fatalf("expected registration success message, got %q", body)
	}
}

// Passwordless registration tests

func TestApiRegisterPasswordlessRequiresFirstName(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
	}

	recorder, req := makePostRequest(t, "/api/register", url.Values{})
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"error\"") {
		t.Fatalf("expected status error, got %q", body)
	}
	if !strings.Contains(body, "\"message\":\"First name is required field\"") {
		t.Fatalf("expected first name required message, got %q", body)
	}
}

func TestApiRegisterPasswordlessRequiresLastName(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
	}

	values := url.Values{
		"first_name": {"John"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Last name is required field\"") {
		t.Fatalf("expected last name required message, got %q", body)
	}
}

func TestApiRegisterPasswordlessRequiresEmail(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Email is required field\"") {
		t.Fatalf("expected email required message, got %q", body)
	}
}

func TestApiRegisterPasswordlessTokenStoreError(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		RegisterPasswordlessInitDependencies: RegisterPasswordlessInitDependencies{
			TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
				return errors.New("db error")
			},
			ExpiresSeconds: 3600,
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Failed to process request. Please try again later\"") {
		t.Fatalf("expected token store error message, got %q", body)
	}
}

func TestApiRegisterPasswordlessEmailSendError(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		RegisterPasswordlessInitDependencies: RegisterPasswordlessInitDependencies{
			TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
				return nil
			},
			ExpiresSeconds: 3600,
			EmailTemplate: func(ctx context.Context, email string, verificationCode string) string {
				return "body"
			},
			EmailSend: func(ctx context.Context, email string, subject string, body string) error {
				return errors.New("smtp error")
			},
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Failed to send email. Please try again later\"") {
		t.Fatalf("expected email send error message, got %q", body)
	}
}

func TestApiRegisterPasswordlessSuccess(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		RegisterPasswordlessInitDependencies: RegisterPasswordlessInitDependencies{
			TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
				return nil
			},
			ExpiresSeconds: 3600,
			EmailTemplate: func(ctx context.Context, email string, verificationCode string) string {
				return "body"
			},
			EmailSend: func(ctx context.Context, email string, subject string, body string) error {
				return nil
			},
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"success\"") {
		t.Fatalf("expected success status, got %q", body)
	}
	if !strings.Contains(body, "\"message\":\"Registration code was sent successfully\"") {
		t.Fatalf("expected success message, got %q", body)
	}
}
