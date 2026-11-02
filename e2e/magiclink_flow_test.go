package e2e

import (
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth"
	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
)

// TestMagicLinkFlow covers the magic-link method end to end through the
// multi-method mux:
//
//  1. POST /auth/login → magic-link token stored (the "email").
//  2. GET /auth/magiclink-callback?token=… → authenticated.
func TestMagicLinkFlow(t *testing.T) {
	store := newTempKeyStore()

	cfg := testutils.NewOtpConfigForTest()
	cfg.FuncTemporaryKeyGet = store.get
	cfg.FuncTemporaryKeySet = store.set

	a, err := auth.NewOtpAuth(cfg)
	if err != nil {
		t.Fatalf("NewOtpAuth: %v", err)
	}

	method := auth.NewMagicLinkMethodFromAuth(a, nil, func(token string) string {
		return "/auth/magiclink-callback?token=" + token
	})

	mma, err := auth.NewMultiMethodAuth("/auth", []types.Method{method})
	if err != nil {
		t.Fatalf("NewMultiMethodAuth: %v", err)
	}
	mux := mma.Router()

	// 1. Request the magic link (primary method's AjaxHandler at /auth/login).
	send := postForm(mux, "/auth/login", url.Values{
		"email": {"jane@example.com"},
	})
	if !strings.Contains(send.Body.String(), `"status":"success"`) {
		t.Fatalf("expected success from magic-link send, got %q", send.Body.String())
	}

	// 2. "Email arrives": extract the stored token and hit the callback.
	token := store.tokenByPrefix(t, utils.MagicLinkKeyPrefix)

	callback := get(mux, "/auth/magiclink-callback?token="+url.QueryEscape(token))
	body := callback.Body.String()
	if callback.Code != 200 && callback.Code != 302 && callback.Code != 303 && callback.Code != 307 {
		t.Fatalf("expected success or redirect from callback, got %d %q", callback.Code, body)
	}
}
