package e2e

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth"
	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/types"
)

// TestAuthKnightFlow covers the AuthKnight registration path: the user's
// email is already verified by the external service (carried by verified_key),
// shown read-only on the register page, and POST /auth/api/register
// creates the user and issues a session token.
func TestAuthKnightFlow(t *testing.T) {
	store := newTempKeyStore()
	var gotEmail, gotFirst, gotLast string

	cfg := testutils.NewAuthKnightConfigForTest()
	cfg.FuncTemporaryKeyGet = func(key string) (string, error) {
		if key == "ak-key-1" {
			return "knight@example.com", nil
		}
		return store.get(key)
	}
	cfg.FuncTemporaryKeySet = store.set
	cfg.FuncUserRegister = func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (string, error) {
		gotEmail, gotFirst, gotLast = email, firstName, lastName
		return "user-1", nil
	}

	a, err := auth.NewAuthKnightAuth(cfg)
	if err != nil {
		t.Fatalf("NewAuthKnightAuth: %v", err)
	}
	mux := a.Router()

	// 1. GET /auth/register?verified_key=… → verified email read-only.
	page := get(mux, "/auth/register?verified_key=ak-key-1")
	body := page.Body.String()
	for _, want := range []string{"knight@example.com", "form-control-plaintext", `name="first_name"`, `name="verified_key"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("authknight register page missing %q", want)
		}
	}

	// 2. POST /auth/api/register → user created, token issued, cookie set.
	reg := postForm(mux, "/auth/api/register", url.Values{
		"verified_key": {"ak-key-1"},
		"first_name":   {"Jane"},
		"last_name":    {"Doe"},
	})
	regBody := reg.Body.String()
	if !strings.Contains(regBody, `"status":"success"`) {
		t.Fatalf("expected success from authknight register, got %q", regBody)
	}
	if gotEmail != "knight@example.com" || gotFirst != "Jane" || gotLast != "Doe" {
		t.Fatalf("FuncUserRegister got (%q, %q, %q)", gotEmail, gotFirst, gotLast)
	}
	if responseField(t, reg, "token") == "" {
		t.Fatalf("expected auth token in register response, got %q", regBody)
	}
}
