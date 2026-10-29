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
			return s, nil
		},
		AuthCookieSet: func(w http.ResponseWriter, r *http.Request, token string) {
			http.SetCookie(w, &http.Cookie{Name: "auth", Value: token})
		},
	}
}

func run(cfg RememberMeConfig, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", "/x", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	RememberMeMiddleware(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)
	return rec
}

func TestRememberMe_ExchangeAndRotate(t *testing.T) {
	store := utils.NewMemSessionStore()
	markedSession(store, "u1", "remember-key")
	rec := run(baseCfg(store), &http.Cookie{Name: "remember", Value: "remember-key"})

	// Old remember key consumed; rotated cookie + auth cookie set.
	if s, _ := store.SessionFindByKey(context.Background(), "remember-key"); s != nil {
		t.Fatal("expected old remember session deleted")
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
}

func TestRememberMe_NoCookiesPassesThrough(t *testing.T) {
	store := utils.NewMemSessionStore()
	rec := run(baseCfg(store))
	if rec.Code != http.StatusOK || len(rec.Result().Cookies()) != 0 {
		t.Fatalf("expected pass-through, got %d %v", rec.Code, rec.Result().Cookies())
	}
}

func TestRememberMe_UnmarkedRememberRejected(t *testing.T) {
	store := utils.NewMemSessionStore()
	s := utils.NewMemSession()
	s.SetUserID("u1")
	s.SetKey("plain-key")
	_ = store.SessionCreate(context.Background(), s)
	rec := run(baseCfg(store), &http.Cookie{Name: "remember", Value: "plain-key"})
	for _, c := range rec.Result().Cookies() {
		if c.Name == "remember" && c.MaxAge >= 0 {
			t.Fatal("expected remember cookie expired")
		}
	}
}

func TestRememberMe_RememberKeyInAuthCookieDeleted(t *testing.T) {
	store := utils.NewMemSessionStore()
	markedSession(store, "u1", "remember-key")
	run(baseCfg(store), &http.Cookie{Name: "auth", Value: "remember-key"})
	if s, _ := store.SessionFindByKey(context.Background(), "remember-key"); s != nil {
		t.Fatal("misused remember key must be deleted")
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
