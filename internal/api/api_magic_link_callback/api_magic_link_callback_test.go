package api_magic_link_callback

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
)

func tokenStore(payload utils.MagicLinkPayload) (map[string]string, func(string) (string, error), func(string, string, int) error) {
	raw, _ := json.Marshal(payload)
	store := map[string]string{utils.MagicLinkKeyPrefix + payload.Token: string(raw)}
	get := func(k string) (string, error) {
		if v, ok := store[k]; ok {
			return v, nil
		}
		return "", errors.New("not found")
	}
	set := func(k, v string, e int) error { store[k] = v; return nil }
	return store, get, set
}

func TestMagicLinkCallbackSuccessSingleUse(t *testing.T) {
	payload := utils.MagicLinkPayload{Email: "u@x.com", Token: "tok123"}
	store, get, set := tokenStore(payload)

	var authedEmail string
	deps := Dependencies{
		TemporaryKeyGet: get,
		TemporaryKeySet: set,
		AuthenticateViaUsername: func(w http.ResponseWriter, r *http.Request, email, first, last string) {
			authedEmail = email
			w.WriteHeader(http.StatusOK)
		},
	}

	req := httptest.NewRequest("GET", "/auth/magiclink-callback?token=tok123", nil)
	rec := httptest.NewRecorder()
	ApiMagicLinkCallback(rec, req, deps)
	if authedEmail != "u@x.com" {
		t.Fatalf("expected auth handoff, got %q", authedEmail)
	}
	if store[utils.MagicLinkKeyPrefix+"tok123"] != "" {
		t.Fatal("expected token consumed")
	}

	// Second use must fail.
	req2 := httptest.NewRequest("GET", "/auth/magiclink-callback?token=tok123", nil)
	rec2 := httptest.NewRecorder()
	ApiMagicLinkCallback(rec2, req2, deps)
	if rec2.Code != http.StatusBadRequest {
		t.Fatalf("expected second use rejected, got %d", rec2.Code)
	}
}

func TestMagicLinkCallbackUnknownToken(t *testing.T) {
	deps := Dependencies{
		TemporaryKeyGet: func(k string) (string, error) { return "", errors.New("nf") },
	}
	req := httptest.NewRequest("GET", "/auth/magiclink-callback?token=ghost", nil)
	rec := httptest.NewRecorder()
	ApiMagicLinkCallback(rec, req, deps)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestMagicLinkCallbackIPBinding(t *testing.T) {
	payload := utils.MagicLinkPayload{Email: "u@x.com", Token: "tok", IP: "203.0.113.9"}
	_, get, set := tokenStore(payload)
	deps := Dependencies{
		TemporaryKeyGet: get,
		TemporaryKeySet: set,
		AuthenticateViaUsername: func(w http.ResponseWriter, r *http.Request, email, first, last string) {
			t.Fatal("must not authenticate on IP mismatch")
		},
	}
	req := httptest.NewRequest("GET", "/auth/magiclink-callback?token=tok", nil)
	req.RemoteAddr = "198.51.100.1:1234" // different IP
	rec := httptest.NewRecorder()
	ApiMagicLinkCallback(rec, req, deps)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected IP mismatch rejected, got %d", rec.Code)
	}
}

func TestMagicLinkCallbackSafeReturnRedirect(t *testing.T) {
	payload := utils.MagicLinkPayload{Email: "u@x.com", Token: "tok", Return: "/dashboard"}
	_, get, set := tokenStore(payload)
	deps := Dependencies{
		TemporaryKeyGet: get,
		TemporaryKeySet: set,
	}
	req := httptest.NewRequest("GET", "/auth/magiclink-callback?token=tok", nil)
	rec := httptest.NewRecorder()
	ApiMagicLinkCallback(rec, req, deps)
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/dashboard" {
		t.Fatalf("expected redirect to /dashboard, got %d %q", rec.Code, rec.Header().Get("Location"))
	}
}

func TestMagicLinkCallbackWrongMethod(t *testing.T) {
	deps := Dependencies{}
	req := httptest.NewRequest("POST", "/auth/magiclink-callback?token=tok", nil)
	rec := httptest.NewRecorder()
	ApiMagicLinkCallback(rec, req, deps)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestMagicLinkCallbackEmailCaptured(t *testing.T) {
	payload := utils.MagicLinkPayload{Email: "u@x.com", Token: "tok-cap"}
	_, get, set := tokenStore(payload)

	captured := ""
	deps := Dependencies{
		TemporaryKeyGet: get,
		TemporaryKeySet: set,
		EmailCaptured: func(ctx context.Context, email string, verified bool, options types.UserAuthOptions) error {
			captured = email
			return nil
		},
		OnSuccess: func(w http.ResponseWriter, r *http.Request, email string, p utils.MagicLinkPayload) {
			w.WriteHeader(http.StatusTeapot)
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/auth/magiclink-callback?token=tok-cap", nil)
	ApiMagicLinkCallback(rec, req, deps)

	if captured != "u@x.com" {
		t.Fatalf("expected capture hook fired, got %q", captured)
	}
}

func TestMagicLinkCallbackRegisterFirstRedirect(t *testing.T) {
	payload := utils.MagicLinkPayload{Email: "new@x.com", Token: "tok-reg"}
	_, get, set := tokenStore(payload)

	deps := Dependencies{
		TemporaryKeyGet:      get,
		TemporaryKeySet:      set,
		CreateUserOnRegister: true,
		RegisterURL:          "/auth/register",
		UserExists: func(ctx context.Context, email string, options types.UserAuthOptions) (bool, error) {
			return false, nil
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/auth/magiclink-callback?token=tok-reg", nil)
	ApiMagicLinkCallback(rec, req, deps)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if !strings.Contains(loc, "/auth/register") || !strings.Contains(loc, "verified_key=") {
		t.Fatalf("expected register redirect with verified_key, got %q", loc)
	}
}

func TestMagicLinkCallbackRegisterFirstExistingUserAuthenticates(t *testing.T) {
	payload := utils.MagicLinkPayload{Email: "old@x.com", Token: "tok-old"}
	_, get, set := tokenStore(payload)

	authed := ""
	deps := Dependencies{
		TemporaryKeyGet:      get,
		TemporaryKeySet:      set,
		CreateUserOnRegister: true,
		UserExists: func(ctx context.Context, email string, options types.UserAuthOptions) (bool, error) {
			return true, nil
		},
		AuthenticateViaUsername: func(w http.ResponseWriter, r *http.Request, email, first, last string) {
			authed = email
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/auth/magiclink-callback?token=tok-old", nil)
	ApiMagicLinkCallback(rec, req, deps)

	if authed != "old@x.com" {
		t.Fatalf("expected existing user authenticated, got %q", authed)
	}
}

func TestMagicLinkCallbackRememberForwarded(t *testing.T) {
	payload := utils.MagicLinkPayload{Email: "u@x.com", Token: "tok-rem", Remember: true}
	_, get, set := tokenStore(payload)

	rememberForm := ""
	deps := Dependencies{
		TemporaryKeyGet: get,
		TemporaryKeySet: set,
		AuthenticateViaUsername: func(w http.ResponseWriter, r *http.Request, email, first, last string) {
			rememberForm = r.FormValue("remember")
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/auth/magiclink-callback?token=tok-rem", nil)
	ApiMagicLinkCallback(rec, req, deps)

	if rememberForm != "true" {
		t.Fatalf("expected remember forwarded, got %q", rememberForm)
	}
}

func TestMagicLinkCallbackOnSuccessAndUnsafeReturn(t *testing.T) {
	payload := utils.MagicLinkPayload{Email: "u@x.com", Token: "tok-os", Return: "//evil.com"}
	_, get, set := tokenStore(payload)

	// Unsafe return falls back to root redirect.
	deps := Dependencies{TemporaryKeyGet: get, TemporaryKeySet: set}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/auth/magiclink-callback?token=tok-os", nil)
	ApiMagicLinkCallback(rec, req, deps)
	if loc := rec.Header().Get("Location"); loc != "/" {
		t.Fatalf("expected root redirect for unsafe return, got %q", loc)
	}

	// OnSuccess override is honoured.
	payload2 := utils.MagicLinkPayload{Email: "u@x.com", Token: "tok-os2"}
	_, get2, set2 := tokenStore(payload2)
	called := false
	deps2 := Dependencies{
		TemporaryKeyGet: get2,
		TemporaryKeySet: set2,
		OnSuccess: func(w http.ResponseWriter, r *http.Request, email string, p utils.MagicLinkPayload) {
			called = true
		},
	}
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/auth/magiclink-callback?token=tok-os2", nil)
	ApiMagicLinkCallback(rec2, req2, deps2)
	if !called {
		t.Fatal("expected OnSuccess invoked")
	}
}
