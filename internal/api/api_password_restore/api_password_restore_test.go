package api_password_restore

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth/internal/testutils"
)

func TestApiPasswordRestoreRequiresEmail(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	deps, err := NewDependencies(
		func(ctx context.Context, email, firstName, lastName string) (string, error) {
			return "", nil
		},
		func(key string, value string, expiresSeconds int) error {
			return nil
		},
		3600,
		func(ctx context.Context, userID, token string) string {
			return ""
		},
		func(ctx context.Context, userID, subject, body string) error {
			return nil
		},
		logger,
	)
	if err != nil {
		t.Fatalf("NewDependencies() error = %v", err)
	}

	recorder, req := testutils.MakePostRequest(t, "/api/password-restore", url.Values{})
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
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	deps, err := NewDependencies(
		func(ctx context.Context, email, firstName, lastName string) (string, error) {
			return "", nil
		},
		func(key string, value string, expiresSeconds int) error {
			return nil
		},
		3600,
		func(ctx context.Context, userID, token string) string {
			return ""
		},
		func(ctx context.Context, userID, subject, body string) error {
			return nil
		},
		logger,
	)
	if err != nil {
		t.Fatalf("NewDependencies() error = %v", err)
	}

	values := url.Values{
		"email": {"test@test.com"},
	}
	recorder, req := testutils.MakePostRequest(t, "/api/password-restore", values)
	ApiPasswordRestore(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"First name is required field\"") {
		t.Fatalf("expected first name required message, got %q", body)
	}
}

func TestApiPasswordRestoreRequiresLastName(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	deps, err := NewDependencies(
		func(ctx context.Context, email, firstName, lastName string) (string, error) {
			return "", nil
		},
		func(key string, value string, expiresSeconds int) error {
			return nil
		},
		3600,
		func(ctx context.Context, userID, token string) string {
			return ""
		},
		func(ctx context.Context, userID, subject, body string) error {
			return nil
		},
		logger,
	)
	if err != nil {
		t.Fatalf("NewDependencies() error = %v", err)
	}
	values := url.Values{
		"email":      {"test@test.com"},
		"first_name": {"John"},
	}
	recorder, req := testutils.MakePostRequest(t, "/api/password-restore", values)
	ApiPasswordRestore(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Last name is required field\"") {
		t.Fatalf("expected last name required message, got %q", body)
	}
}

func TestApiPasswordRestoreUserNotFound(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	deps, err := NewDependencies(
		func(ctx context.Context, email, firstName, lastName string) (string, error) {
			return "", nil
		},
		func(key string, value string, expiresSeconds int) error {
			return nil
		},
		3600,
		func(ctx context.Context, userID, token string) string {
			return ""
		},
		func(ctx context.Context, userID, subject, body string) error {
			return nil
		},
		logger,
	)
	if err != nil {
		t.Fatalf("NewDependencies() error = %v", err)
	}
	values := url.Values{
		"email":      {"test@test.com"},
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}
	recorder, req := testutils.MakePostRequest(t, "/api/password-restore", values)
	ApiPasswordRestore(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"User not found\"") {
		t.Fatalf("expected user not found message, got %q", body)
	}
}

func TestApiPasswordRestoreInternalError(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))

	deps, err := NewDependencies(
		func(ctx context.Context, email, firstName, lastName string) (string, error) {
			return "", errors.New("db error")
		},
		func(key string, value string, expiresSeconds int) error {
			return nil
		},
		3600,
		func(ctx context.Context, userID, token string) string {
			return ""
		},
		func(ctx context.Context, userID, subject, body string) error {
			return nil
		},
		logger,
	)
	if err != nil {
		t.Fatalf("NewDependencies() error = %v", err)
	}

	values := url.Values{
		"email":      {"test@test.com"},
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}

	recorder, req := testutils.MakePostRequest(t, "/api/password-restore", values)
	ApiPasswordRestore(recorder, req, deps)

	body := recorder.Body.String()

	expected := `"message":"Internal server error. Please try again later"`
	if !strings.Contains(body, expected) {
		t.Fatalf("expected: %q, got: %q", expected, body)
	}

	t.Log(logBuf.String())
	if !strings.Contains(logBuf.String(), "db error") {
		t.Fatalf("expected logger to contain underlying error 'db error', got: %q", logBuf.String())
	}
}

func TestApiPasswordRestoreSuccess(t *testing.T) {
	userFound := false
	tempKeySetCalled := false
	emailSent := false

	deps, err := NewDependencies(
		func(ctx context.Context, email, firstName, lastName string) (string, error) {
			userFound = true
			return "user123", nil
		},
		func(key string, value string, expiresSeconds int) error {
			tempKeySetCalled = true
			return nil
		},
		3600,
		func(ctx context.Context, userID, token string) string {
			return "email-body"
		},
		func(ctx context.Context, userID, subject, body string) error {
			emailSent = true
			return nil
		},
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		t.Fatalf("NewDependencies() error = %v", err)
	}

	values := url.Values{
		"email":      {"test@test.com"},
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}
	recorder, req := testutils.MakePostRequest(t, "/api/password-restore", values)
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
