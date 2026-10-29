package api_magic_link_callback

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

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
