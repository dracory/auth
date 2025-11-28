package api_login

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// helper to build a POST request with form values
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

// Username/password login tests

func TestApiLoginUsernameAndPasswordRequiresEmail(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		LoginWithUsernameAndPassword: func(ctx context.Context, email, password, ip, userAgent string) (string, string, string) {
			if email == "" {
				return "", "", "Email is required field"
			}
			return "", "", ""
		},
	}

	recorder, req := makePostRequest(t, "/api/login", url.Values{})
	ApiLogin(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"error"`) {
		t.Fatalf("expected status error, got %q", body)
	}
	if !strings.Contains(body, `"message":"Email is required field"`) {
		t.Fatalf("expected email required message, got %q", body)
	}
}

func TestApiLoginUsernameAndPasswordRequiresPassword(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		LoginWithUsernameAndPassword: func(ctx context.Context, email, password, ip, userAgent string) (string, string, string) {
			if password == "" {
				return "", "", "Password is required field"
			}
			return "", "", ""
		},
	}

	values := url.Values{
		"email": {"test@test.com"},
	}
	recorder, req := makePostRequest(t, "/api/login", values)
	ApiLogin(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"message":"Password is required field"`) {
		t.Fatalf("expected password required message, got %q", body)
	}
}

func TestApiLoginUsernameAndPasswordUserLoginError(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		LoginWithUsernameAndPassword: func(ctx context.Context, email, password, ip, userAgent string) (string, string, string) {
			return "", "", "Invalid credentials"
		},
	}

	values := url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}
	recorder, req := makePostRequest(t, "/api/login", values)
	ApiLogin(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"message":"Invalid credentials"`) {
		t.Fatalf("expected invalid credentials message, got %q", body)
	}
}

func TestApiLoginUsernameAndPasswordUserNotFound(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		LoginWithUsernameAndPassword: func(ctx context.Context, email, password, ip, userAgent string) (string, string, string) {
			// Simulate user not found by returning empty token and error message
			return "", "", "Invalid credentials"
		},
	}

	values := url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}
	recorder, req := makePostRequest(t, "/api/login", values)
	ApiLogin(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"message":"Invalid credentials"`) {
		t.Fatalf("expected invalid credentials message, got %q", body)
	}
}

func TestApiLoginUsernameAndPasswordTokenStoreError(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		LoginWithUsernameAndPassword: func(ctx context.Context, email, password, ip, userAgent string) (string, string, string) {
			return "", "", "Failed to process request. Please try again later"
		},
	}

	values := url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}
	recorder, req := makePostRequest(t, "/api/login", values)
	ApiLogin(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"message":"Failed to process request. Please try again later"`) {
		t.Fatalf("expected token store error message, got %q", body)
	}
}

func TestApiLoginUsernameAndPasswordSuccess(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		UseCookies:   false,
		LoginWithUsernameAndPassword: func(ctx context.Context, email, password, ip, userAgent string) (string, string, string) {
			return "login success", "token-123", ""
		},
	}

	values := url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}
	recorder, req := makePostRequest(t, "/api/login", values)
	ApiLogin(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"success"`) {
		t.Fatalf("expected success status, got %q", body)
	}
	if !strings.Contains(body, `"message":"login success"`) {
		t.Fatalf("expected success message, got %q", body)
	}
	if !strings.Contains(body, `"token":"token-123"`) {
		t.Fatalf("expected token in response, got %q", body)
	}
}

// Passwordless login tests

func TestApiLoginPasswordlessRequiresEmail(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		PasswordlessDependencies: LoginPasswordlessDeps{
			DisableRateLimit: false,
			TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
				return nil
			},
			ExpiresSeconds: 3600,
			EmailTemplate: func(ctx context.Context, email string, code string) string {
				return ""
			},
			EmailSend: func(ctx context.Context, email string, subject string, body string) error {
				return nil
			},
		},
	}

	recorder, req := makePostRequest(t, "/api/login", url.Values{})
	ApiLogin(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"message":"Email is required field"`) {
		t.Fatalf("expected email required message, got %q", body)
	}
}

func TestApiLoginPasswordlessInvalidEmail(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		PasswordlessDependencies: LoginPasswordlessDeps{
			DisableRateLimit: false,
			TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
				return nil
			},
			ExpiresSeconds: 3600,
			EmailTemplate: func(ctx context.Context, email string, code string) string {
				return ""
			},
			EmailSend: func(ctx context.Context, email string, subject string, body string) error {
				return nil
			},
		},
	}

	values := url.Values{
		"email": {"invalid-email"},
	}
	recorder, req := makePostRequest(t, "/api/login", values)
	ApiLogin(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"message":"This is not a valid email: invalid-email"`) {
		t.Fatalf("expected invalid email message, got %q", body)
	}
}

func TestApiLoginPasswordlessTokenStoreError(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		PasswordlessDependencies: LoginPasswordlessDeps{
			DisableRateLimit: false,
			TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
				return errors.New("db error")
			},
			ExpiresSeconds: 3600,
			EmailTemplate: func(ctx context.Context, email string, code string) string {
				return ""
			},
			EmailSend: func(ctx context.Context, email string, subject string, body string) error {
				return nil
			},
		},
	}

	values := url.Values{
		"email": {"test@test.com"},
	}
	recorder, req := makePostRequest(t, "/api/login", values)
	ApiLogin(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"message":"Failed to process request. Please try again later"`) {
		t.Fatalf("expected token store error message, got %q", body)
	}
}

func TestApiLoginPasswordlessEmailSendError(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		PasswordlessDependencies: LoginPasswordlessDeps{
			DisableRateLimit: false,
			TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
				return nil
			},
			ExpiresSeconds: 3600,
			EmailTemplate: func(ctx context.Context, email string, code string) string {
				return ""
			},
			EmailSend: func(ctx context.Context, email string, subject string, body string) error {
				return errors.New("smtp error")
			},
		},
	}

	values := url.Values{
		"email": {"test@test.com"},
	}
	recorder, req := makePostRequest(t, "/api/login", values)
	ApiLogin(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"message":"Failed to send email. Please try again later"`) {
		t.Fatalf("expected email send error message, got %q", body)
	}
}

func TestApiLoginPasswordlessSuccess(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		PasswordlessDependencies: LoginPasswordlessDeps{
			DisableRateLimit: false,
			TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
				return nil
			},
			ExpiresSeconds: 3600,
			EmailTemplate: func(ctx context.Context, email string, code string) string {
				return ""
			},
			EmailSend: func(ctx context.Context, email string, subject string, body string) error {
				return nil
			},
		},
	}

	values := url.Values{
		"email": {"test@test.com"},
	}
	recorder, req := makePostRequest(t, "/api/login", values)
	ApiLogin(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, `"status":"success"`) {
		t.Fatalf("expected success status, got %q", body)
	}
	if !strings.Contains(body, `"message":"Login code was sent successfully"`) {
		t.Fatalf("expected success message, got %q", body)
	}
}
