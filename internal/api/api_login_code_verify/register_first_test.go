package api_login_code_verify

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth/utils"
)

// newTempStore returns a map-backed temporary key store pair.
func newTempStore() (get func(string) (string, error), set func(string, string, int) error, store map[string]string) {
	store = map[string]string{}
	get = func(key string) (string, error) {
		v, ok := store[key]
		if !ok {
			return "", errors.New("not found")
		}
		return v, nil
	}
	set = func(key, value string, expires int) error {
		store[key] = value
		return nil
	}
	return get, set, store
}

func seedLoginCode(store map[string]string, nonce, email, code string) {
	store[utils.LoginCodeKeyPrefix+nonce] = email + ":" + code
	store[utils.LoginCodeAttemptsKeyPrefix+nonce] = "0"
}

func validVerifyForm(email, nonce, code string) url.Values {
	return url.Values{
		"verification_code": {code},
		"nonce":             {nonce},
		"email":             {email},
	}
}

// Register-first: a verified email with no account receives a
// verified-email token + register redirect instead of a session.
func TestApiLoginCodeVerify_RegisterFirstNewUser(t *testing.T) {
	get, set, store := newTempStore()
	seedLoginCode(store, "11111111111111111111111111111111", "new@example.com", "BCDFGHJK")

	authCalled := false
	var capturedEmail string
	var capturedVerified bool

	deps := Dependencies{
		TemporaryKeyGet:      get,
		TemporaryKeySet:      set,
		CreateUserOnRegister: true,
		RegisterURL:          "/auth/register",
		UserExists: func(ctx context.Context, email string) (bool, error) {
			return false, nil
		},
		EmailCaptured: func(ctx context.Context, email string, verified bool) error {
			capturedEmail, capturedVerified = email, verified
			return nil
		},
		AuthenticateViaUsername: func(w http.ResponseWriter, r *http.Request, email string) {
			authCalled = true
		},
	}

	recorder, req := makePostRequest(t, "/api/login-code-verify", validVerifyForm("new@example.com", "11111111111111111111111111111111", "BCDFGHJK"))
	ApiLoginCodeVerify(recorder, req, deps)

	body := recorder.Body.String()
	if authCalled {
		t.Fatalf("new user must not be authenticated under register-first")
	}
	if !strings.Contains(body, "\"status\":\"success\"") {
		t.Fatalf("expected success, got %q", body)
	}
	if !strings.Contains(body, "/auth/register?ak_key=") {
		t.Fatalf("expected register redirect with token, got %q", body)
	}

	// The issued token resolves back to the verified email in the store.
	if !strings.Contains(body, "ak_key=") {
		t.Fatalf("missing token in response: %q", body)
	}
	found := false
	for k, v := range store {
		if v == "new@example.com" && !strings.HasPrefix(k, utils.LoginCodeKeyPrefix) {
			found = true
			if !strings.Contains(body, k) {
				t.Fatalf("response token %q not found in store key %q", body, k)
			}
		}
	}
	if !found {
		t.Fatalf("no verified-email token stored: %v", store)
	}

	if capturedEmail != "new@example.com" || !capturedVerified {
		t.Fatalf("EmailCaptured got (%q,%v), want (new@example.com,true)", capturedEmail, capturedVerified)
	}
}

// Register-first: an existing account logs in normally — no token.
func TestApiLoginCodeVerify_RegisterFirstExistingUser(t *testing.T) {
	get, set, store := newTempStore()
	seedLoginCode(store, "22222222222222222222222222222222", "old@example.com", "BCDFGHJK")

	authCalled := false
	deps := Dependencies{
		TemporaryKeyGet:      get,
		TemporaryKeySet:      set,
		CreateUserOnRegister: true,
		RegisterURL:          "/auth/register",
		UserExists: func(ctx context.Context, email string) (bool, error) {
			return true, nil
		},
		AuthenticateViaUsername: func(w http.ResponseWriter, r *http.Request, email string) {
			authCalled = true
			_, _ = w.Write([]byte(`{"status":"success"}`))
		},
	}

	recorder, req := makePostRequest(t, "/api/login-code-verify", validVerifyForm("old@example.com", "22222222222222222222222222222222", "BCDFGHJK"))
	ApiLoginCodeVerify(recorder, req, deps)

	if !authCalled {
		t.Fatalf("existing user must be authenticated")
	}
	if strings.Contains(recorder.Body.String(), "ak_key=") {
		t.Fatalf("existing user must not receive a verified-email token")
	}
}

// Default policy unchanged: no UserExists configured → authenticate
// path preserved even when CreateUserOnRegister is somehow set.
func TestApiLoginCodeVerify_RegisterFirstNoUserExistsFallback(t *testing.T) {
	get, set, store := newTempStore()
	seedLoginCode(store, "33333333333333333333333333333333", "x@example.com", "BCDFGHJK")

	authCalled := false
	deps := Dependencies{
		TemporaryKeyGet:      get,
		TemporaryKeySet:      set,
		CreateUserOnRegister: true,
		AuthenticateViaUsername: func(w http.ResponseWriter, r *http.Request, email string) {
			authCalled = true
			_, _ = w.Write([]byte(`{"status":"success"}`))
		},
	}

	recorder, req := makePostRequest(t, "/api/login-code-verify", validVerifyForm("x@example.com", "33333333333333333333333333333333", "BCDFGHJK"))
	ApiLoginCodeVerify(recorder, req, deps)

	if !authCalled {
		t.Fatalf("missing UserExists must fall back to authenticate")
	}
}
