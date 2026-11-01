package e2e

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth"
	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
)

// newOTPAuth builds a passwordless auth impl backed by the given
// temp-key store and returns its mux. registerCalls, when non-nil,
// collects each FuncUserRegister invocation.
func newOTPAuth(t *testing.T, store *tempKeyStore, registerCalls *[]struct {
	Email, First, Last string
},
) *http.ServeMux {
	t.Helper()

	cfg := testutils.NewPasswordlessConfigForTest()
	cfg.FuncTemporaryKeyGet = store.get
	cfg.FuncTemporaryKeySet = store.set
	cfg.FuncUserRegister = func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) error {
		if registerCalls != nil {
			*registerCalls = append(*registerCalls, struct {
				Email, First, Last string
			}{email, firstName, lastName})
		}
		return nil
	}

	a, err := auth.NewPasswordlessAuth(cfg)
	if err != nil {
		t.Fatalf("NewPasswordlessAuth: %v", err)
	}
	return a.Router()
}

// otpLoginCode extracts the login code stored for a nonce — the
// logincode value is "email:code".
func otpLoginCode(t *testing.T, store *tempKeyStore, nonce string) string {
	t.Helper()
	raw, _ := store.get(utils.LoginCodeKeyPrefix + nonce)
	code := raw[strings.LastIndex(raw, ":")+1:]
	if code == "" || code == raw {
		t.Fatalf("no login code stored under the nonce (have %q)", raw)
	}
	return code
}

// TestOTPLogin_Guest: a guest proves email ownership via a login code,
// then is authenticated.
func TestOTPLogin_Guest(t *testing.T) {
	store := newTempKeyStore()
	mux := newOTPAuth(t, store, nil)

	// 1. POST /auth/api/login sends a login code to the email.
	login := postForm(mux, "/auth/api/login", url.Values{
		"email": {"jane@example.com"},
	})
	nonce := responseField(t, login, "nonce")
	if nonce == "" {
		t.Fatalf("expected nonce in login response, got %q", login.Body.String())
	}

	// 2. POST /auth/api/login-code-verify proves ownership → authenticated.
	verify := postForm(mux, "/auth/api/login-code-verify", url.Values{
		"email":             {"jane@example.com"},
		"nonce":             {nonce},
		"verification_code": {otpLoginCode(t, store, nonce)},
	})
	if !strings.Contains(verify.Body.String(), `"status":"success"`) {
		t.Fatalf("expected success from login-code-verify, got %q", verify.Body.String())
	}
}

// TestOTPRegister_Guest: a guest's typed email is untrusted, so
// registration goes through a code round-trip — only then is the user
// created (equivalent to AuthKnight's verified_key).
//
// A user ALREADY logged in via OTP never goes through this — their email
// was proven at login, so /auth/register serves the completion form
// instead (see register_completion_flow_test.go).
func TestOTPRegister_Guest(t *testing.T) {
	store := newTempKeyStore()
	var registerCalls []struct {
		Email, First, Last string
	}
	mux := newOTPAuth(t, store, &registerCalls)

	// 1. POST /auth/api/register sends a registration code.
	reg := postForm(mux, "/auth/api/register", url.Values{
		"first_name": {"Jane"},
		"last_name":  {"Doe"},
		"email":      {"jane@example.com"},
	})
	nonce := responseField(t, reg, "nonce")
	if nonce == "" {
		t.Fatalf("expected nonce in register response, got %q", reg.Body.String())
	}

	// 2. POST /auth/api/register-code-verify proves ownership → user created.
	code := store.payloadField(t, utils.RegisterCodeKeyPrefix+nonce, "code")
	if code == "" {
		t.Fatal("no verification code stored under the nonce")
	}
	verify := postForm(mux, "/auth/api/register-code-verify", url.Values{
		"nonce":             {nonce},
		"verification_code": {code},
	})
	if !strings.Contains(verify.Body.String(), `"status":"success"`) {
		t.Fatalf("expected success from register-code-verify, got %q", verify.Body.String())
	}
	if len(registerCalls) != 1 || registerCalls[0].Email != "jane@example.com" ||
		registerCalls[0].First != "Jane" || registerCalls[0].Last != "Doe" {
		t.Fatalf("FuncUserRegister got %+v", registerCalls)
	}
}
