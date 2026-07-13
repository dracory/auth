package api_password_reset

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth/types"
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

func TestApiPasswordResetRequiresToken(t *testing.T) {
	deps := Dependencies{}

	recorder, req := makePostRequest(t, "/api/password-reset", url.Values{})
	ApiPasswordReset(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"error\"") {
		t.Fatalf("expected status error, got %q", body)
	}
	if !strings.Contains(body, "\"message\":\"Token is required field\"") {
		t.Fatalf("expected token required message, got %q", body)
	}
}

func TestApiPasswordResetRequiresPassword(t *testing.T) {
	deps := Dependencies{}

	values := url.Values{
		"token": {"valid-token"},
	}
	recorder, req := makePostRequest(t, "/api/password-reset", values)
	ApiPasswordReset(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Password is required field\"") {
		t.Fatalf("expected password required message, got %q", body)
	}
}

func TestApiPasswordResetRequiresMatchingPasswords(t *testing.T) {
	deps := Dependencies{}

	values := url.Values{
		"token":            {"valid-token"},
		"password":         {"password123"},
		"password_confirm": {"password321"},
	}
	recorder, req := makePostRequest(t, "/api/password-reset", values)
	ApiPasswordReset(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Passwords do not match\"") {
		t.Fatalf("expected passwords do not match message, got %q", body)
	}
}

func TestApiPasswordResetInvalidToken(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			// Simulate invalid token (no user ID)
			return "", nil
		},
	}

	values := url.Values{
		"token":            {"invalid-token"},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}
	recorder, req := makePostRequest(t, "/api/password-reset", values)
	ApiPasswordReset(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Link not valid or expired\"") {
		t.Fatalf("expected invalid link message, got %q", body)
	}
}

func TestApiPasswordResetPasswordChangeError(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			return "user123", nil
		},
		UserPasswordChange: func(ctx context.Context, userID, password string) error {
			return errors.New("db error")
		},
	}

	values := url.Values{
		"token":            {"valid-token"},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}
	recorder, req := makePostRequest(t, "/api/password-reset", values)
	ApiPasswordReset(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Password reset failed. Please try again later\"") {
		t.Fatalf("expected password reset failed message, got %q", body)
	}
}

func TestApiPasswordResetLogoutError(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			return "user123", nil
		},
		UserPasswordChange: func(ctx context.Context, userID, password string) error {
			return nil
		},
		LogoutUser: func(ctx context.Context, userID string) error {
			return errors.New("logout failed")
		},
	}

	values := url.Values{
		"token":            {"valid-token"},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}
	recorder, req := makePostRequest(t, "/api/password-reset", values)
	ApiPasswordReset(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Logout failed. Please try again later\"") {
		t.Fatalf("expected logout failed message, got %q", body)
	}
}

func TestApiPasswordResetSuccess(t *testing.T) {
	logoutCalled := false

	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			return "user123", nil
		},
		UserPasswordChange: func(ctx context.Context, userID, password string) error {
			return nil
		},
		LogoutUser: func(ctx context.Context, userID string) error {
			logoutCalled = true
			return nil
		},
	}

	values := url.Values{
		"token":            {"valid-token"},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}
	recorder, req := makePostRequest(t, "/api/password-reset", values)
	ApiPasswordReset(recorder, req, deps)

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

	if !logoutCalled {
		t.Fatalf("LogoutUser should be called on successful password reset")
	}
}

func TestPasswordResetError_Error(t *testing.T) {
	var e *PasswordResetError
	if e.Error() != "" {
		t.Fatalf("expected empty string for nil receiver, got %q", e.Error())
	}

	e = &PasswordResetError{Code: PasswordResetErrorCodeValidation}
	if e.Error() != "validation" {
		t.Fatalf("expected 'validation', got %q", e.Error())
	}

	e = &PasswordResetError{Message: "custom message"}
	if e.Error() != "custom message" {
		t.Fatalf("expected 'custom message', got %q", e.Error())
	}

	e = &PasswordResetError{Err: errors.New("inner error")}
	if e.Error() != "inner error" {
		t.Fatalf("expected 'inner error', got %q", e.Error())
	}
}

func TestApiPasswordResetPasswordStrengthError(t *testing.T) {
	deps := Dependencies{
		PasswordStrength: &types.PasswordStrengthConfig{
			MinLength:        12,
			RequireUppercase: true,
			RequireLowercase: true,
			RequireDigit:     true,
			RequireSpecial:   true,
		},
		TemporaryKeyGet: func(key string) (string, error) {
			return "user123", nil
		},
		UserPasswordChange: func(ctx context.Context, userID, password string) error {
			return nil
		},
	}

	values := url.Values{
		"token":            {"valid-token"},
		"password":         {"short"},
		"password_confirm": {"short"},
	}
	recorder, req := makePostRequest(t, "/api/password-reset", values)
	ApiPasswordReset(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"error\"") {
		t.Fatalf("expected error status, got %q", body)
	}
}

func TestApiPasswordResetTemporaryKeyGetNil(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet: nil,
	}

	values := url.Values{
		"token":            {"valid-token"},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}
	recorder, req := makePostRequest(t, "/api/password-reset", values)
	ApiPasswordReset(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "Link not valid or expired") {
		t.Fatalf("expected link not valid, got %q", body)
	}
}

func TestApiPasswordResetTemporaryKeyGetError(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			return "", errors.New("db error")
		},
	}

	values := url.Values{
		"token":            {"valid-token"},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}
	recorder, req := makePostRequest(t, "/api/password-reset", values)
	ApiPasswordReset(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "Link not valid or expired") {
		t.Fatalf("expected link not valid, got %q", body)
	}
}

func TestApiPasswordResetUserPasswordChangeNil(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			return "user123", nil
		},
		UserPasswordChange: nil,
	}

	values := url.Values{
		"token":            {"valid-token"},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}
	recorder, req := makePostRequest(t, "/api/password-reset", values)
	ApiPasswordReset(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "Password reset failed") {
		t.Fatalf("expected password reset failed, got %q", body)
	}
}

func TestApiPasswordResetSuccessWithoutLogoutUser(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet: func(key string) (string, error) {
			return "user123", nil
		},
		UserPasswordChange: func(ctx context.Context, userID, password string) error {
			return nil
		},
		LogoutUser: nil,
	}

	values := url.Values{
		"token":            {"valid-token"},
		"password":         {"password123"},
		"password_confirm": {"password123"},
	}
	recorder, req := makePostRequest(t, "/api/password-reset", values)
	ApiPasswordReset(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"success\"") {
		t.Fatalf("expected success, got %q", body)
	}
}
