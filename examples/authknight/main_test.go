package main

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	authtypes "github.com/dracory/auth/types"
)

func resetAuthKnightStore() {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.usersByEmail = make(map[string]*authKnightUser)
	store.sessions = make(map[string]string)
	store.tempKeys = make(map[string]string)
}

func TestAuthKnightRegisterAndFindUser(t *testing.T) {
	resetAuthKnightStore()

	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	email := "alice@example.com"
	userID, err := store.registerUser(ctx, email, "Alice", "Smith", opts)
	if err != nil {
		t.Fatalf("registerUser unexpected error: %v", err)
	}
	if userID == "" {
		t.Fatalf("expected non-empty user ID")
	}

	foundID, err := store.findUserByEmail(ctx, email, opts)
	if err != nil {
		t.Fatalf("findUserByEmail unexpected error: %v", err)
	}
	if foundID != userID {
		t.Fatalf("expected userID %q, got %q", userID, foundID)
	}

	// Duplicate registration should fail
	if _, err := store.registerUser(ctx, email, "Alice", "Smith", opts); err == nil {
		t.Fatalf("expected error when registering duplicate email, got nil")
	}
}

func TestAuthKnightAuthTokenLifecycle(t *testing.T) {
	resetAuthKnightStore()

	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	userID := "user-1"
	token := "token-ak-123"

	if err := store.storeAuthToken(ctx, token, userID, opts); err != nil {
		t.Fatalf("storeAuthToken unexpected error: %v", err)
	}

	gotID, err := store.findByAuthToken(ctx, token, opts)
	if err != nil {
		t.Fatalf("findByAuthToken unexpected error: %v", err)
	}
	if gotID != userID {
		t.Fatalf("expected userID %q, got %q", userID, gotID)
	}

	if err := store.logout(ctx, userID, opts); err != nil {
		t.Fatalf("logout unexpected error: %v", err)
	}

	if _, err := store.findByAuthToken(ctx, token, opts); err == nil {
		t.Fatalf("expected error after logout, got nil")
	}
}

func TestAuthKnightTempKeys(t *testing.T) {
	resetAuthKnightStore()

	key := "ak_once_123"
	val := "alice@example.com"

	if err := store.tempKeySet(key, val, 300); err != nil {
		t.Fatalf("tempKeySet unexpected error: %v", err)
	}

	got, err := store.tempKeyGet(key)
	if err != nil {
		t.Fatalf("tempKeyGet unexpected error: %v", err)
	}
	if got != val {
		t.Fatalf("expected value %q, got %q", val, got)
	}

	if _, err := store.tempKeyGet("missing"); err == nil {
		t.Fatalf("expected error for missing temp key, got nil")
	}
}

func TestAuthKnightDisplayName(t *testing.T) {
	resetAuthKnightStore()

	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	email := "bob@example.com"
	userID, err := store.registerUser(ctx, email, "Bob", "Jones", opts)
	if err != nil {
		t.Fatalf("registerUser unexpected error: %v", err)
	}

	displayName := store.displayName(userID)
	if !strings.Contains(displayName, "Bob") || !strings.Contains(displayName, "Jones") {
		t.Fatalf("expected display name to contain Bob Jones, got %q", displayName)
	}
}

func TestAuthKnightStoreHelpers(t *testing.T) {
	resetAuthKnightStore()
	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	store.usersByEmail["noname@x.com"] = &authKnightUser{ID: "u-noname", Email: "noname@x.com"}

	if got := store.displayName("u-noname"); got != "noname@x.com" {
		t.Fatalf("expected email fallback, got %q", got)
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
}

func TestAuthKnightMuxServesPages(t *testing.T) {
	resetAuthKnightStore()

	authInstance, err := newAuth()
	if err != nil {
		t.Fatal(err)
	}
	mux := newMux(authInstance)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "AuthKnight") {
		t.Fatalf("expected home page, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/debug?x=1", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "debug") {
		t.Fatalf("expected debug page, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/dashboard", nil))
	if rec.Code == 200 {
		t.Fatal("expected redirect for unauthenticated dashboard")
	}
}
