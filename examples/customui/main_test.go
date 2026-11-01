package main

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	authtypes "github.com/dracory/auth/types"
)

func resetCustomUIStore() {
	store.mu.Lock()
	defer store.mu.Unlock()

	demoUser := &customUIUser{ID: "usr-1", Username: "user@aura.local", Password: "aurapassword", FirstName: "Aura", LastName: "Member"}
	store.users = map[string]*customUIUser{
		demoUser.Username: demoUser,
	}
	store.sessions = make(map[string]string)
	store.tempKeys = make(map[string]string)
}

func TestCustomUILayoutFunction(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/auth/login", nil)

	title := "Login Title"
	body := "<form>Custom Form</form>"
	opts := authtypes.LayoutOptions{
		Styles: []string{".custom { color: red; }"},
	}

	html := customLayout(rec, req, title, body, opts)
	if !strings.Contains(html, title) {
		t.Fatalf("expected output HTML to contain title %q", title)
	}
	if !strings.Contains(html, body) {
		t.Fatalf("expected output HTML to contain body %q", body)
	}
	if !strings.Contains(html, "Aura Security Systems") {
		t.Fatalf("expected output HTML to contain footer text")
	}
}

func TestCustomUIStoreLoginAndAuthToken(t *testing.T) {
	resetCustomUIStore()
	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	userID, err := store.userLogin(ctx, "user@aura.local", "aurapassword", opts)
	if err != nil || userID != "usr-1" {
		t.Fatalf("userLogin failed, got %q (err: %v)", userID, err)
	}

	token := "customui-token"
	if err := store.storeAuthToken(ctx, token, userID, opts); err != nil {
		t.Fatalf("storeAuthToken failed: %v", err)
	}

	foundID, err := store.findByAuthToken(ctx, token, opts)
	if err != nil || foundID != "usr-1" {
		t.Fatalf("findByAuthToken failed: %v", err)
	}

	if err := store.logout(ctx, userID, opts); err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	if _, err := store.findByAuthToken(ctx, token, opts); err == nil {
		t.Fatalf("expected error after logout")
	}
}

func TestCustomUIStoreHelpers(t *testing.T) {
	resetCustomUIStore()
	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	// Seed a user directly.
	store.users["a@b.com"] = &customUIUser{ID: "u1", Username: "a@b.com", Password: "pw"}
	store.users["noname@x.com"] = &customUIUser{ID: "u2", Username: "noname@x.com", Password: "pw"}

	if id, err := store.userFindByUsername(ctx, "a@b.com", "", "", opts); err != nil || id != "u1" {
		t.Fatalf("userFindByUsername failed: %v %q", err, id)
	}
	if _, err := store.userFindByUsername(ctx, "missing", "", "", opts); err == nil {
		t.Fatal("expected error for unknown user")
	}
	if got := store.displayName("u2"); got != "noname@x.com" {
		t.Fatalf("expected username fallback, got %q", got)
	}
	if got := store.displayName("missing"); got != "missing" {
		t.Fatalf("expected id fallback, got %q", got)
	}

	if err := store.tempKeySet("k", "v", 60); err != nil {
		t.Fatal(err)
	}
	if v, err := store.tempKeyGet("k"); err != nil || v != "v" {
		t.Fatalf("tempKeyGet failed: %v %q", err, v)
	}
	if _, err := store.tempKeyGet("missing"); err == nil {
		t.Fatal("expected error for missing key")
	}

	if err := store.storeAuthToken(ctx, "tok", "u1", opts); err != nil {
		t.Fatal(err)
	}
	if err := store.logout(ctx, "u1", opts); err != nil {
		t.Fatal(err)
	}
	if _, err := store.findByAuthToken(ctx, "tok", opts); err == nil {
		t.Fatal("expected revoked token")
	}

	if err := emailSend(ctx, "u1", "subj", "body"); err != nil {
		t.Fatal(err)
	}
}

func TestCustomUIMuxServesPages(t *testing.T) {
	resetCustomUIStore()

	authInstance, err := newAuth()
	if err != nil {
		t.Fatal(err)
	}
	mux := newMux(authInstance)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "Aura") {
		t.Fatalf("expected home page, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/auth/login", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "AURA PLATFORM") {
		t.Fatalf("expected custom login page, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/dashboard", nil))
	if rec.Code == 200 {
		t.Fatal("expected redirect for unauthenticated dashboard")
	}
}
