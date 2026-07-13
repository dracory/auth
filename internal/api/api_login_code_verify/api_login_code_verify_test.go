package api_login_code_verify

import (
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

func TestApiLoginCodeVerifyRequiresVerificationCode(t *testing.T) {
	deps := Dependencies{}

	recorder, req := makePostRequest(t, "/api/login-code-verify", url.Values{})
	ApiLoginCodeVerify(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"error\"") {
		t.Fatalf("expected status error, got %q", body)
	}
	if !strings.Contains(body, "\"message\":\"Verification code is required field\"") {
		t.Fatalf("expected verification code required message, got %q", body)
	}
}

func TestApiLoginCodeVerifyInvalidLength(t *testing.T) {
	deps := Dependencies{}

	values := url.Values{
		"verification_code": {"123456"},
	}
	recorder, req := makePostRequest(t, "/api/login-code-verify", values)
	ApiLoginCodeVerify(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Verification code is invalid length\"") {
		t.Fatalf("expected invalid length message, got %q", body)
	}
}

func TestApiLoginCodeVerifyInvalidCharacters(t *testing.T) {
	deps := Dependencies{}

	values := url.Values{
		"verification_code": {"12345678"},
	}
	recorder, req := makePostRequest(t, "/api/login-code-verify", values)
	ApiLoginCodeVerify(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Verification code contains invalid characters\"") {
		t.Fatalf("expected invalid characters message, got %q", body)
	}
}

func TestApiLoginCodeVerifyExpiredCode(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			return "", errors.New("expired")
		},
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}
	recorder, req := makePostRequest(t, "/api/login-code-verify", values)
	ApiLoginCodeVerify(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Verification code has expired\"") {
		t.Fatalf("expected expired code message, got %q", body)
	}
}

func TestApiLoginCodeVerifyAuthenticationError(t *testing.T) {
	called := false
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			return "user@example.com", nil
		},
		AuthenticateViaUsername: func(w http.ResponseWriter, r *http.Request, email string) {
			called = true
			// Simulate authentication error response
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"error","message":"Invalid credentials"}`))
		},
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}
	recorder, req := makePostRequest(t, "/api/login-code-verify", values)
	ApiLoginCodeVerify(recorder, req, deps)

	if !called {
		t.Fatalf("AuthenticateViaUsername should be called")
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Invalid credentials\"") {
		t.Fatalf("expected invalid credentials message, got %q", body)
	}
}

func TestApiLoginCodeVerifyUserNotFound(t *testing.T) {
	called := false
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			return "user@example.com", nil
		},
		AuthenticateViaUsername: func(w http.ResponseWriter, r *http.Request, email string) {
			called = true
			// Simulate user not found response (same message as authentication error)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"error","message":"Invalid credentials"}`))
		},
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}
	recorder, req := makePostRequest(t, "/api/login-code-verify", values)
	ApiLoginCodeVerify(recorder, req, deps)

	if !called {
		t.Fatalf("AuthenticateViaUsername should be called")
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Invalid credentials\"") {
		t.Fatalf("expected invalid credentials message, got %q", body)
	}
}

func TestApiLoginCodeVerifyTokenStoreError(t *testing.T) {
	called := false
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			return "user@example.com", nil
		},
		AuthenticateViaUsername: func(w http.ResponseWriter, r *http.Request, email string) {
			called = true
			// Simulate token store error response
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"error","message":"Failed to process request. Please try again later"}`))
		},
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}
	recorder, req := makePostRequest(t, "/api/login-code-verify", values)
	ApiLoginCodeVerify(recorder, req, deps)

	if !called {
		t.Fatalf("AuthenticateViaUsername should be called")
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Failed to process request. Please try again later\"") {
		t.Fatalf("expected token store error message, got %q", body)
	}
}

func TestApiLoginCodeVerifySuccess(t *testing.T) {
	called := false
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			return "user@example.com", nil
		},
		AuthenticateViaUsername: func(w http.ResponseWriter, r *http.Request, email string) {
			called = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"success","message":"login success","token":"token-123"}`))
		},
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}
	recorder, req := makePostRequest(t, "/api/login-code-verify", values)
	ApiLoginCodeVerify(recorder, req, deps)

	if !called {
		t.Fatalf("AuthenticateViaUsername should be called")
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"success\"") {
		t.Fatalf("expected success status, got %q", body)
	}
	if !strings.Contains(body, "\"message\":\"login success\"") {
		t.Fatalf("expected login success message, got %q", body)
	}
	if !strings.Contains(body, "\"token\":\"") {
		t.Fatalf("expected token in response, got %q", body)
	}
}

func TestLoginCodeVerifyError_Error(t *testing.T) {
	var e *LoginCodeVerifyError
	if e.Error() != "" {
		t.Fatalf("expected empty string for nil receiver, got %q", e.Error())
	}

	e = &LoginCodeVerifyError{Message: "custom message"}
	if e.Error() != "custom message" {
		t.Fatalf("expected 'custom message', got %q", e.Error())
	}

	e = &LoginCodeVerifyError{Err: errors.New("inner error")}
	if e.Error() != "inner error" {
		t.Fatalf("expected 'inner error', got %q", e.Error())
	}

	e = &LoginCodeVerifyError{Code: "some_code"}
	if e.Error() != "some_code" {
		t.Fatalf("expected 'some_code', got %q", e.Error())
	}
}

func TestApiLoginCodeVerifyNilTemporaryKeyGet(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet: nil,
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}
	recorder, req := makePostRequest(t, "/api/login-code-verify", values)
	ApiLoginCodeVerify(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "Verification code has expired") {
		t.Fatalf("expected expired message, got %q", body)
	}
}

func TestApiLoginCodeVerifyNilAuthenticateViaUsername(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			return "user@example.com", nil
		},
		AuthenticateViaUsername: nil,
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
	}
	recorder, req := makePostRequest(t, "/api/login-code-verify", values)
	ApiLoginCodeVerify(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "Failed to process request") {
		t.Fatalf("expected failed to process, got %q", body)
	}
}
