package main

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	authtypes "github.com/dracory/auth/types"
)

func TestMultiMethodStoreUserLifecycle(t *testing.T) {
	store := newMemoryStore()
	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	username := "user@example.com"
	pass := "secret123"

	// Register
	if err := store.userRegister(ctx, username, pass, "Jane", "Doe", opts); err != nil {
		t.Fatalf("userRegister unexpected error: %v", err)
	}

	// Login
	userID, err := store.userLogin(ctx, username, pass, opts)
	if err != nil {
		t.Fatalf("userLogin unexpected error: %v", err)
	}
	if userID == "" {
		t.Fatalf("expected non-empty user ID")
	}

	// Find by username & find by email
	foundID1, err := store.userFindByUsername(ctx, username, "Jane", "Doe", opts)
	if err != nil || foundID1 != userID {
		t.Fatalf("userFindByUsername failed, got ID %q (err: %v)", foundID1, err)
	}

	foundID2, err := store.userFindByEmail(ctx, username, opts)
	if err != nil || foundID2 != userID {
		t.Fatalf("userFindByEmail failed, got ID %q (err: %v)", foundID2, err)
	}

	// Password change
	newPass := "newsecret456"
	if err := store.userPasswordChange(ctx, username, newPass, opts); err != nil {
		t.Fatalf("userPasswordChange unexpected error: %v", err)
	}

	if _, err := store.userLogin(ctx, username, pass, opts); err == nil {
		t.Fatalf("expected error logging in with old password")
	}

	if _, err := store.userLogin(ctx, username, newPass, opts); err != nil {
		t.Fatalf("userLogin with new password failed: %v", err)
	}
}

func TestMultiMethodStoreAuthTokenAndLogout(t *testing.T) {
	store := newMemoryStore()
	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	userID := "1"
	token := "multi-token-123"

	if err := store.storeAuthToken(ctx, token, userID, opts); err != nil {
		t.Fatalf("storeAuthToken unexpected error: %v", err)
	}

	gotID, err := store.findByAuthToken(ctx, token, opts)
	if err != nil || gotID != userID {
		t.Fatalf("findByAuthToken failed, got ID %q (err: %v)", gotID, err)
	}

	if err := store.logout(ctx, userID, opts); err != nil {
		t.Fatalf("logout unexpected error: %v", err)
	}

	gotIDAfter, _ := store.findByAuthToken(ctx, token, opts)
	if gotIDAfter != "" {
		t.Fatalf("expected empty token after logout, got %q", gotIDAfter)
	}
}

func TestMultiMethodTempKeyStorage(t *testing.T) {
	store := newMemoryStore()

	key := "otp-key"
	val := "999888"

	if err := store.tempKeySet(key, val, 60); err != nil {
		t.Fatalf("tempKeySet unexpected error: %v", err)
	}

	got, err := store.tempKeyGet(key)
	if err != nil || got != val {
		t.Fatalf("tempKeyGet failed, got %q (err: %v)", got, err)
	}

	if _, err := store.tempKeyGet("missing"); err == nil {
		t.Fatalf("expected error for missing key")
	}
}

func TestMultiMethodStoreDisplayNameAndEmailSend(t *testing.T) {
	s := newMemoryStore()
	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	if err := s.userRegister(ctx, "n@x.com", "", "Ni", "Mo", opts); err != nil {
		t.Fatal(err)
	}
	if got := s.displayName("1"); got != "Ni Mo" {
		t.Fatalf("expected full name, got %q", got)
	}
	if got := s.displayName("missing"); got != "User" {
		t.Fatalf("expected fallback name, got %q", got)
	}
	if err := emailSend(ctx, "a@b.com", "s", "b"); err != nil {
		t.Fatal(err)
	}
}

func TestSiteLayout(t *testing.T) {
	html := siteLayout(nil, nil, "Title", "<p>body</p>", authtypes.LayoutOptions{
		StyleURLs:  []string{"/a.css"},
		Styles:     []string{"body{}"},
		ScriptURLs: []string{"/a.js"},
		Scripts:    []string{"console.log(1)"},
	})
	for _, v := range []string{"Title", "<p>body</p>", "/a.css", "body{}", "/a.js", "console.log(1)", "Acme Corp"} {
		if !strings.Contains(html, v) {
			t.Fatalf("expected %q in layout, got:\n%s", v, html)
		}
	}
}

func TestMultiMethodMuxServesPages(t *testing.T) {
	mux, err := newMux()
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "Home") {
		t.Fatalf("expected home page, got %d", rec.Code)
	}

	for _, path := range []string{"/auth/login", "/auth/login-otp", "/auth/login-magiclink"} {
		rec = httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != 200 {
			t.Fatalf("expected %s to render, got %d", path, rec.Code)
		}
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/dashboard", nil))
	if rec.Code == 200 {
		t.Fatal("expected redirect for unauthenticated dashboard")
	}
}
