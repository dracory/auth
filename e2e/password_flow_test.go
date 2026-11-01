package e2e

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth"
	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
)

// TestPasswordFlow covers the username(email)+password method end to end:
//
//  1. POST /auth/api/register → verification code stored →
//     POST /auth/api/register-code-verify → user created + authenticated.
//  2. POST /auth/api/login with the new credentials → authenticated.
//  3. POST /auth/api/restore-password → reset token stored →
//     POST /auth/api/reset-password → password changed.
func TestPasswordFlow(t *testing.T) {
	store := newTempKeyStore()
	var gotEmail, gotPassword, changedPassword string

	cfg := testutils.NewUsernameAndPasswordConfigForTest()
	cfg.EnableVerification = true
	cfg.FuncTemporaryKeyGet = store.get
	cfg.FuncTemporaryKeySet = store.set
	cfg.FuncEmailTemplateRegisterCode = func(ctx context.Context, email, link string, options types.UserAuthOptions) string {
		return "Your code: " + link
	}
	cfg.FuncEmailTemplatePasswordRestore = func(ctx context.Context, userID, link string, options types.UserAuthOptions) string {
		return "Reset: " + link
	}
	cfg.FuncUserRegister = func(ctx context.Context, username, password, firstName, lastName string, options types.UserAuthOptions) error {
		gotEmail, gotPassword = username, password
		return nil
	}
	// After registration the flow authenticates the user — the username
	// lookup must resolve the just-created account.
	cfg.FuncUserFindByUsername = func(ctx context.Context, username, firstName, lastName string, options types.UserAuthOptions) (string, error) {
		return "user-1", nil
	}
	cfg.FuncUserLogin = func(ctx context.Context, username, password string, options types.UserAuthOptions) (string, error) {
		if username == "jane@example.com" && password == "long-enough-password" {
			return "user-1", nil
		}
		return "", nil
	}
	cfg.FuncUserPasswordChange = func(ctx context.Context, username, newPassword string, options types.UserAuthOptions) error {
		changedPassword = newPassword
		return nil
	}

	a, err := auth.NewUsernameAndPasswordAuth(cfg)
	if err != nil {
		t.Fatalf("NewUsernameAndPasswordAuth: %v", err)
	}
	mux := a.Router()

	// ---- Register: code sent → verified → user created -------------------

	reg := postForm(mux, "/auth/api/register", url.Values{
		"first_name": {"Jane"},
		"last_name":  {"Doe"},
		"email":      {"jane@example.com"},
		"password":   {"long-enough-password"},
	})
	nonce := responseField(t, reg, "nonce")
	if nonce == "" {
		t.Fatalf("expected nonce in register response, got %q", reg.Body.String())
	}

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
	if gotEmail != "jane@example.com" || gotPassword != "long-enough-password" {
		t.Fatalf("FuncUserRegister got (%q, %q)", gotEmail, gotPassword)
	}

	// ---- Login -----------------------------------------------------------

	login := postForm(mux, "/auth/api/login", url.Values{
		"email":    {"jane@example.com"},
		"password": {"long-enough-password"},
	})
	if !strings.Contains(login.Body.String(), `"status":"success"`) {
		t.Fatalf("expected login success, got %q", login.Body.String())
	}

	// ---- Password reset: restore → token → reset -------------------------

	restore := postForm(mux, "/auth/api/restore-password", url.Values{
		"first_name": {"Jane"},
		"last_name":  {"Doe"},
		"email":      {"jane@example.com"},
	})
	if !strings.Contains(restore.Body.String(), `"status":"success"`) {
		t.Fatalf("expected success from restore-password, got %q", restore.Body.String())
	}

	token := store.tokenByPrefix(t, utils.PasswordResetTokenKeyPrefix)

	reset := postForm(mux, "/auth/api/reset-password", url.Values{
		"token":            {token},
		"password":         {"brand-new-password"},
		"password_confirm": {"brand-new-password"},
	})
	if !strings.Contains(reset.Body.String(), `"status":"success"`) {
		t.Fatalf("expected success from reset-password, got %q", reset.Body.String())
	}
	if changedPassword != "brand-new-password" {
		t.Fatalf("FuncUserPasswordChange got %q", changedPassword)
	}
}
