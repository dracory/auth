package api_login_code_verify

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth/utils"
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
		"nonce":             {"0123456789abcdef0123456789abcdef"},
		"email":             {"user@example.com"},
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
		"nonce":             {"0123456789abcdef0123456789abcdef"},
		"email":             {"user@example.com"},
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
		"nonce":             {"0123456789abcdef0123456789abcdef"},
		"email":             {"user@example.com"},
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
			return "user@example.com:BCDFGHJK", nil
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
		"nonce":             {"0123456789abcdef0123456789abcdef"},
		"email":             {"user@example.com"},
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
			return "user@example.com:BCDFGHJK", nil
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
		"nonce":             {"0123456789abcdef0123456789abcdef"},
		"email":             {"user@example.com"},
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
			return "user@example.com:BCDFGHJK", nil
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
		"nonce":             {"0123456789abcdef0123456789abcdef"},
		"email":             {"user@example.com"},
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
			return "user@example.com:BCDFGHJK", nil
		},
		AuthenticateViaUsername: func(w http.ResponseWriter, r *http.Request, email string) {
			called = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"success","message":"login success","token":"token-123"}`))
		},
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
		"nonce":             {"0123456789abcdef0123456789abcdef"},
		"email":             {"user@example.com"},
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
		"nonce":             {"0123456789abcdef0123456789abcdef"},
		"email":             {"user@example.com"},
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
			return "user@example.com:BCDFGHJK", nil
		},
		AuthenticateViaUsername: nil,
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
		"nonce":             {"0123456789abcdef0123456789abcdef"},
		"email":             {"user@example.com"},
	}
	recorder, req := makePostRequest(t, "/api/login-code-verify", values)
	ApiLoginCodeVerify(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "Failed to process request") {
		t.Fatalf("expected failed to process, got %q", body)
	}
}

func TestApiLoginCodeVerifyRequiresNonce(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			t.Fatalf("store must not be read when nonce is missing")
			return "", nil
		},
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
		"email":             {"user@example.com"},
	}
	recorder, req := makePostRequest(t, "/api/login-code-verify", values)
	ApiLoginCodeVerify(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Nonce is required field\"") {
		t.Fatalf("expected nonce required message, got %q", body)
	}
}

func TestApiLoginCodeVerifyRequiresEmail(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			t.Fatalf("store must not be read when email is missing")
			return "", nil
		},
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
		"nonce":             {"0123456789abcdef0123456789abcdef"},
	}
	recorder, req := makePostRequest(t, "/api/login-code-verify", values)
	ApiLoginCodeVerify(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Email is required field\"") {
		t.Fatalf("expected email required message, got %q", body)
	}
}

func TestApiLoginCodeVerifyWrongCode(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			return "user@example.com:BCDFGHJK", nil
		},
	}

	// Valid nonce but wrong code -> generic expired message, no auth.
	values := url.Values{
		"verification_code": {"XXXXXXXX"},
		"nonce":             {"0123456789abcdef0123456789abcdef"},
		"email":             {"user@example.com"},
	}
	recorder, req := makePostRequest(t, "/api/login-code-verify", values)
	ApiLoginCodeVerify(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Verification code has expired\"") {
		t.Fatalf("expected generic expired message, got %q", body)
	}
}

func TestApiLoginCodeVerifyWrongEmail(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			return "user@example.com:BCDFGHJK", nil
		},
	}

	// Correct code but different email -> generic expired message.
	values := url.Values{
		"verification_code": {"BCDFGHJK"},
		"nonce":             {"0123456789abcdef0123456789abcdef"},
		"email":             {"other@example.com"},
	}
	recorder, req := makePostRequest(t, "/api/login-code-verify", values)
	ApiLoginCodeVerify(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Verification code has expired\"") {
		t.Fatalf("expected generic expired message, got %q", body)
	}
}

func TestApiLoginCodeVerifySingleUseConsume(t *testing.T) {
	writes := map[string][2]any{}
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			if key == utils.LoginCodeAttemptsKeyPrefix+"0123456789abcdef0123456789abcdef" {
				return "0", nil
			}
			if key != utils.LoginCodeKeyPrefix+"0123456789abcdef0123456789abcdef" {
				t.Fatalf("expected lookup by prefixed nonce, got %q", key)
			}
			return "user@example.com:BCDFGHJK", nil
		},
		TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
			writes[key] = [2]any{value, expiresSeconds}
			return nil
		},
		AuthenticateViaUsername: func(w http.ResponseWriter, r *http.Request, email string) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"success"}`))
		},
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"},
		"nonce":             {"0123456789abcdef0123456789abcdef"},
		"email":             {"user@example.com"},
	}
	recorder, req := makePostRequest(t, "/api/login-code-verify", values)
	ApiLoginCodeVerify(recorder, req, deps)

	codeKey := utils.LoginCodeKeyPrefix + "0123456789abcdef0123456789abcdef"
	attemptsKey := utils.LoginCodeAttemptsKeyPrefix + "0123456789abcdef0123456789abcdef"

	// Both keys must end invalidated (empty value, 0 expiry).
	if w, ok := writes[codeKey]; !ok || w[0].(string) != "" || w[1].(int) != 0 {
		t.Fatalf("expected code key invalidated (empty value, 0 expiry), got %v", writes[codeKey])
	}
	if w, ok := writes[attemptsKey]; !ok || w[0].(string) != "" || w[1].(int) != 0 {
		t.Fatalf("expected attempts key invalidated (empty value, 0 expiry), got %v", writes[attemptsKey])
	}
}

// TestApiLoginCodeVerifyAttemptLimit verifies the per-nonce attempt cap:
// at MaxCodeAttempts the nonce is rejected and both keys are invalidated.
func TestApiLoginCodeVerifyAttemptLimit(t *testing.T) {
	writes := map[string][2]any{}
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			if key == utils.LoginCodeAttemptsKeyPrefix+"0123456789abcdef0123456789abcdef" {
				return "5", nil // already at the limit
			}
			if key == utils.LoginCodeKeyPrefix+"0123456789abcdef0123456789abcdef" {
				return "user@example.com:BCDFGHJK", nil
			}
			return "", errors.New("not found")
		},
		TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
			writes[key] = [2]any{value, expiresSeconds}
			return nil
		},
		AuthenticateViaUsername: func(w http.ResponseWriter, r *http.Request, email string) {
			t.Fatalf("AuthenticateViaUsername must not be called at the attempt limit")
		},
	}

	values := url.Values{
		"verification_code": {"BCDFGHJK"}, // correct code — still rejected
		"nonce":             {"0123456789abcdef0123456789abcdef"},
		"email":             {"user@example.com"},
	}
	recorder, req := makePostRequest(t, "/api/login-code-verify", values)
	ApiLoginCodeVerify(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Too many attempts, please request a new code\"") {
		t.Fatalf("expected too many attempts message, got %q", body)
	}

	codeKey := utils.LoginCodeKeyPrefix + "0123456789abcdef0123456789abcdef"
	attemptsKey := utils.LoginCodeAttemptsKeyPrefix + "0123456789abcdef0123456789abcdef"
	if w, ok := writes[codeKey]; !ok || w[0].(string) != "" {
		t.Fatalf("expected code key invalidated at limit, got %v", writes[codeKey])
	}
	if w, ok := writes[attemptsKey]; !ok || w[0].(string) != "" {
		t.Fatalf("expected attempts key invalidated at limit, got %v", writes[attemptsKey])
	}
}

// TestApiLoginCodeVerifyAttemptsIncrement verifies each failed verify bumps
// the per-nonce counter before the compare.
func TestApiLoginCodeVerifyAttemptsIncrement(t *testing.T) {
	writes := map[string][2]any{}
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			if key == utils.LoginCodeAttemptsKeyPrefix+"0123456789abcdef0123456789abcdef" {
				return "2", nil
			}
			if key == utils.LoginCodeKeyPrefix+"0123456789abcdef0123456789abcdef" {
				return "user@example.com:BCDFGHJK", nil
			}
			return "", errors.New("not found")
		},
		TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
			writes[key] = [2]any{value, expiresSeconds}
			return nil
		},
	}

	// Wrong code on attempt 3 -> counter goes to 3, generic error.
	values := url.Values{
		"verification_code": {"XXXXXXXX"},
		"nonce":             {"0123456789abcdef0123456789abcdef"},
		"email":             {"user@example.com"},
	}
	recorder, req := makePostRequest(t, "/api/login-code-verify", values)
	ApiLoginCodeVerify(recorder, req, deps)

	attemptsKey := utils.LoginCodeAttemptsKeyPrefix + "0123456789abcdef0123456789abcdef"
	if w := writes[attemptsKey]; w[0].(string) != "3" {
		t.Fatalf("expected attempts counter incremented to 3, got %v", w[0])
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Verification code has expired\"") {
		t.Fatalf("expected generic expired message, got %q", body)
	}
}
