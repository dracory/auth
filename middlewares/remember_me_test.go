package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
)

func markedSession(store *utils.MemSessionStore, userID, key string) *utils.MemSession {
	s := utils.NewMemSession()
	s.SetUserID(userID)
	s.SetKey(key)
	s.SetExpiresAt(time.Now().Add(24 * time.Hour))
	s.SetValue(utils.RememberSessionMarker, "1")
	_ = store.SessionCreate(context.Background(), s)
	return s
}

func baseCfg(store *utils.MemSessionStore) RememberMeConfig {
	return RememberMeConfig{
		Enabled:            true,
		SessionStore:       store,
		AuthCookieName:     "auth",
		RememberCookieName: "remember",
		IssueRememberSession: func(ctx context.Context, userID string, o types.UserAuthOptions) (types.SessionInterface, error) {
			s := utils.NewMemSession()
			s.SetUserID(userID)
			s.SetKey("rotated-key")
			s.SetValue(utils.RememberSessionMarker, "1")
			return s, nil
		},
		IssueAuthSession: func(ctx context.Context, userID string, o types.UserAuthOptions) (types.SessionInterface, error) {
			s := utils.NewMemSession()
			s.SetUserID(userID)
			s.SetKey("auth-key")
			s.SetExpiresAt(time.Now().Add(2 * time.Hour))
			return s, nil
		},
		AuthCookieSet: func(w http.ResponseWriter, r *http.Request, token string) {
			http.SetCookie(w, &http.Cookie{Name: "auth", Value: token})
		},
	}
}

func run(cfg RememberMeConfig, cookies ...*http.Cookie) (*httptest.ResponseRecorder, *http.Request) {
	req := httptest.NewRequest("GET", "/x", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	var seen *http.Request
	RememberMeMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)
	return rec, seen
}

func TestRememberMe_ExchangeAndRotate(t *testing.T) {
	store := utils.NewMemSessionStore()
	markedSession(store, "u1", "remember-key")
	rec, seen := run(baseCfg(store), &http.Cookie{Name: "remember", Value: "remember-key"})

	// Old remember session shortened to the grace window (not deleted).
	old, err := store.SessionFindByKey(context.Background(), "remember-key")
	if err != nil || old == nil {
		t.Fatal("expected old remember session kept within grace window")
	}
	if old.GetExpiresAt().After(time.Now().Add(2 * time.Minute)) {
		t.Fatal("expected old remember session shortened to grace window")
	}
	foundAuth, foundRemember := false, false
	for _, c := range rec.Result().Cookies() {
		if c.Name == "auth" && c.Value == "auth-key" {
			foundAuth = true
		}
		if c.Name == "remember" && c.Value == "rotated-key" {
			foundRemember = true
		}
	}
	if !foundAuth || !foundRemember {
		t.Fatalf("expected auth+rotated cookies, got %v", rec.Result().Cookies())
	}
	// Request cookie rewritten so downstream sees the minted session.
	if c, err := seen.Cookie("auth"); err != nil || c.Value != "auth-key" {
		t.Fatal("expected rewritten request cookie with new auth key")
	}
}

func TestRememberMe_NoCookiesPassesThrough(t *testing.T) {
	store := utils.NewMemSessionStore()
	rec, _ := run(baseCfg(store))
	if rec.Code != http.StatusOK || len(rec.Result().Cookies()) != 0 {
		t.Fatalf("expected pass-through, got %d %v", rec.Code, rec.Result().Cookies())
	}
}

func TestRememberMe_UnmarkedRememberRejected(t *testing.T) {
	store := utils.NewMemSessionStore()
	s := utils.NewMemSession()
	s.SetUserID("u1")
	s.SetKey("plain-key")
	s.SetExpiresAt(time.Now().Add(24 * time.Hour))
	_ = store.SessionCreate(context.Background(), s)
	rec, _ := run(baseCfg(store), &http.Cookie{Name: "remember", Value: "plain-key"})
	for _, c := range rec.Result().Cookies() {
		if c.Name == "remember" && c.MaxAge >= 0 {
			t.Fatal("expected remember cookie expired")
		}
	}
}

func TestRememberMe_RememberKeyInAuthCookieStripped(t *testing.T) {
	store := utils.NewMemSessionStore()
	markedSession(store, "u1", "remember-key")
	rec, seen := run(baseCfg(store), &http.Cookie{Name: "auth", Value: "remember-key"})

	// The cookie is stripped from the request and expired — the session
	// itself is left for the rotation flow (it is still a valid remember
	// token held in the remember cookie).
	if _, err := seen.Cookie("auth"); err == nil {
		t.Fatal("expected auth cookie stripped from request")
	}
	foundExpire := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == "auth" && c.MaxAge < 0 {
			foundExpire = true
		}
	}
	if !foundExpire {
		t.Fatal("expected expiring auth cookie")
	}
	if s, _ := store.SessionFindByKey(context.Background(), "remember-key"); s == nil {
		t.Fatal("remember session must not be deleted on misuse")
	}
}

func TestRememberMe_InactiveUserRevoked(t *testing.T) {
	store := utils.NewMemSessionStore()
	markedSession(store, "u1", "remember-key")
	cfg := baseCfg(store)
	cfg.UserIsActive = func(ctx context.Context, userID string) bool { return false }
	rec, _ := run(cfg, &http.Cookie{Name: "remember", Value: "remember-key"})

	if s, _ := store.SessionFindByKey(context.Background(), "remember-key"); s != nil {
		t.Fatal("expected remember session revoked for inactive user")
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == "remember" && c.MaxAge >= 0 {
			t.Fatal("expected remember cookie expired")
		}
	}
}

func TestRememberMe_DisabledStillStripsMisuse(t *testing.T) {
	store := utils.NewMemSessionStore()
	markedSession(store, "u1", "remember-key")
	cfg := baseCfg(store)
	cfg.Enabled = false
	rec, _ := run(cfg,
		&http.Cookie{Name: "auth", Value: "remember-key"},
		&http.Cookie{Name: "remember", Value: "remember-key"})

	foundExpire := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == "auth" && c.MaxAge < 0 {
			foundExpire = true
		}
	}
	if !foundExpire {
		t.Fatal("expected expiring auth cookie even when disabled")
	}
}

func TestRememberMe_ValidAuthSessionPassesThrough(t *testing.T) {
	store := utils.NewMemSessionStore()
	s := utils.NewMemSession()
	s.SetUserID("u1")
	s.SetKey("auth-key")
	s.SetExpiresAt(time.Now().Add(2 * time.Hour))
	_ = store.SessionCreate(context.Background(), s)
	rec, _ := run(baseCfg(store), &http.Cookie{Name: "auth", Value: "auth-key"})
	if len(rec.Result().Cookies()) != 0 {
		t.Fatalf("expected no cookies touched, got %v", rec.Result().Cookies())
	}
}

func TestRememberMeCheckbox_Fragment(t *testing.T) {
	if utils.RememberMeCheckbox(false) != "" {
		t.Fatal("expected empty when disabled")
	}
	if utils.RememberMeCheckbox(true) == "" {
		t.Fatal("expected checkbox when enabled")
	}
}
