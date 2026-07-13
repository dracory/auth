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

func TestLoginPasswordlessError_Error(t *testing.T) {
	var e *LoginPasswordlessError
	if e.Error() != "" {
		t.Fatalf("expected empty string for nil receiver, got %q", e.Error())
	}

	e = &LoginPasswordlessError{Message: "custom message"}
	if e.Error() != "custom message" {
		t.Fatalf("expected 'custom message', got %q", e.Error())
	}

	e = &LoginPasswordlessError{Err: errors.New("inner error")}
	if e.Error() != "inner error" {
		t.Fatalf("expected 'inner error', got %q", e.Error())
	}

	e = &LoginPasswordlessError{Code: "some_code"}
	if e.Error() != "some_code" {
		t.Fatalf("expected 'some_code', got %q", e.Error())
	}
}

func TestApiLoginNilLoginWithUsernameAndPassword(t *testing.T) {
	deps := Dependencies{
		Passwordless:                 false,
		LoginWithUsernameAndPassword: nil,
	}

	values := url.Values{
		"email":    {"test@test.com"},
		"password": {"1234"},
	}
	recorder, req := makePostRequest(t, "/api/login", values)
	ApiLogin(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "Internal server error") {
		t.Fatalf("expected internal server error, got %q", body)
	}
}

func TestApiLoginSuccessWithCookies(t *testing.T) {
	cookieSet := false
	deps := Dependencies{
		Passwordless: false,
		UseCookies:   true,
		LoginWithUsernameAndPassword: func(ctx context.Context, email, password, ip, userAgent string) (string, string, string) {
			return "login success", "token-123", ""
		},
		SetAuthCookie: func(w http.ResponseWriter, r *http.Request, token string) {
			cookieSet = true
			if token != "token-123" {
				t.Fatalf("expected token 'token-123', got %q", token)
			}
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
		t.Fatalf("expected success, got %q", body)
	}
	if !cookieSet {
		t.Fatal("expected SetAuthCookie to be called")
	}
}

func TestApiLoginPasswordlessNilEmailTemplateAndSend(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		PasswordlessDependencies: LoginPasswordlessDeps{
			DisableRateLimit: false,
			TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
				return nil
			},
			ExpiresSeconds: 3600,
			EmailTemplate:  nil,
			EmailSend:      nil,
		},
	}

	values := url.Values{
		"email": {"test@test.com"},
	}
	recorder, req := makePostRequest(t, "/api/login", values)
	ApiLogin(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "Failed to send email") {
		t.Fatalf("expected failed to send email, got %q", body)
	}
}

func TestApiLoginPasswordlessNilTemporaryKeySet(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		PasswordlessDependencies: LoginPasswordlessDeps{
			DisableRateLimit: false,
			TemporaryKeySet:  nil,
			ExpiresSeconds:   3600,
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
	if !strings.Contains(body, "Failed to process request") {
		t.Fatalf("expected failed to process request, got %q", body)
	}
}

func TestApiLoginPasswordlessDefaultExpiration(t *testing.T) {
	keySetExpires := 0
	deps := Dependencies{
		Passwordless: true,
		PasswordlessDependencies: LoginPasswordlessDeps{
			DisableRateLimit: false,
			TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
				keySetExpires = expiresSeconds
				return nil
			},
			ExpiresSeconds: 0,
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

	if keySetExpires != 3600 {
		t.Fatalf("expected default expiration 3600, got %d", keySetExpires)
	}
}
