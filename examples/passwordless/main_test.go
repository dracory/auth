package main

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	authtypes "github.com/dracory/auth/types"
)

// resetPasswordlessStore is a helper to reset the global store between tests.
func resetPasswordlessStore() {
	passwordlessStore = &passwordlessMemoryStore{
		usersByEmail: make(map[string]*passwordlessUser),
		sessions:     make(map[string]string),
		tempKeys:     make(map[string]string),
	}
}

func TestPasswordlessRegisterAndFindUser(t *testing.T) {
	resetPasswordlessStore()

	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	email := "user@example.com"
	if err := passwordlessStore.registerUser(ctx, email, "John", "Doe", opts); err != nil {
		t.Fatalf("registerUser unexpected error: %v", err)
	}

	id, err := passwordlessStore.findUserByEmail(ctx, email, opts)
	if err != nil {
		t.Fatalf("findUserByEmail unexpected error: %v", err)
	}
	if id == "" {
		t.Fatalf("expected non-empty user ID")
	}

	// Registering the same email again should fail
	if err := passwordlessStore.registerUser(ctx, email, "John", "Doe", opts); err == nil {
		t.Fatalf("expected error when registering duplicate email, got nil")
	}
}

func TestPasswordlessAuthTokenLifecycle(t *testing.T) {
	resetPasswordlessStore()

	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	userID := "user-1"
	token := "token-123"

	if err := passwordlessStore.storeAuthToken(ctx, token, userID, opts); err != nil {
		t.Fatalf("storeAuthToken unexpected error: %v", err)
	}

	gotID, err := passwordlessStore.findByAuthToken(ctx, token, opts)
	if err != nil {
		t.Fatalf("findByAuthToken unexpected error: %v", err)
	}
	if gotID != userID {
		t.Fatalf("expected userID %q, got %q", userID, gotID)
	}

	// Invalid token should return an error
	if _, err := passwordlessStore.findByAuthToken(ctx, "invalid", opts); err == nil {
		t.Fatalf("expected error for invalid token, got nil")
	}
}

func TestPasswordlessTempKeys(t *testing.T) {
	resetPasswordlessStore()

	key := "code-123"
	value := "user@example.com"

	if err := passwordlessStore.tempKeySet(key, value, 60); err != nil {
		t.Fatalf("tempKeySet unexpected error: %v", err)
	}

	got, err := passwordlessStore.tempKeyGet(key)
	if err != nil {
		t.Fatalf("tempKeyGet unexpected error: %v", err)
	}
	if got != value {
		t.Fatalf("expected value %q, got %q", value, got)
	}

	// Unknown key should return an error
	if _, err := passwordlessStore.tempKeyGet("missing"); err == nil {
		t.Fatalf("expected error for missing key, got nil")
	}
}

func TestPasswordlessLogoutClearsSessionsForUser(t *testing.T) {
	resetPasswordlessStore()

	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	userID1 := "user-1"
	userID2 := "user-2"

	// two tokens for user1, one for user2
	_ = passwordlessStore.storeAuthToken(ctx, "t1", userID1, opts)
	_ = passwordlessStore.storeAuthToken(ctx, "t2", userID1, opts)
	_ = passwordlessStore.storeAuthToken(ctx, "t3", userID2, opts)

	if err := passwordlessStore.logout(ctx, userID1, opts); err != nil {
		t.Fatalf("logout unexpected error: %v", err)
	}

	if _, err := passwordlessStore.findByAuthToken(ctx, "t1", opts); err == nil {
		t.Fatalf("expected error for token t1 after logout, got nil")
	}
	if _, err := passwordlessStore.findByAuthToken(ctx, "t2", opts); err == nil {
		t.Fatalf("expected error for token t2 after logout, got nil")
	}

	// Token for user2 should still be valid
	if _, err := passwordlessStore.findByAuthToken(ctx, "t3", opts); err != nil {
		t.Fatalf("expected token t3 to remain valid, got error: %v", err)
	}
}

func TestPasswordlessDisplayNameUsesRegisteredName(t *testing.T) {
	resetPasswordlessStore()

	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	email := "user@example.com"
	firstName := "Alice"
	lastName := "Smith"

	if err := passwordlessStore.registerUser(ctx, email, firstName, lastName, opts); err != nil {
		t.Fatalf("registerUser unexpected error: %v", err)
	}

	userID, err := passwordlessStore.findUserByEmail(ctx, email, opts)
	if err != nil {
		t.Fatalf("findUserByEmail unexpected error: %v", err)
	}

	display := passwordlessStore.displayName(userID)
	if !strings.Contains(display, firstName) || !strings.Contains(display, lastName) {
		t.Fatalf("expected display name to contain first and last name, got %q", display)
	}
}

func TestPasswordlessDisplayNameFallbacks(t *testing.T) {
	resetPasswordlessStore()
	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	// User with no names falls back to email.
	passwordlessStore.mu.Lock()
	passwordlessStore.usersByEmail["noname@example.com"] = &passwordlessUser{ID: "u-noname", Email: "noname@example.com"}
	passwordlessStore.mu.Unlock()

	if got := passwordlessStore.displayName("u-noname"); got != "noname@example.com" {
		t.Fatalf("expected email fallback, got %q", got)
	}
	if got := passwordlessStore.displayName("missing"); got != "missing" {
		t.Fatalf("expected id fallback, got %q", got)
	}

	if err := passwordlessStore.registerUser(ctx, "named@example.com", "Jane", "Roe", opts); err != nil {
		t.Fatal(err)
	}
	id, err := passwordlessStore.findUserByEmail(ctx, "named@example.com", opts)
	if err != nil {
		t.Fatal(err)
	}
	if got := passwordlessStore.displayName(id); got != "Jane Roe" {
		t.Fatalf("expected full name, got %q", got)
	}
}

func TestPasswordlessEmailSend(t *testing.T) {
	// No local mail server — expect a dial error, not a panic.
	err := passwordlessEmailSend(context.Background(), "a@b.com", "subj", "body")
	if err == nil {
		t.Skip("local mail server is running; send succeeded")
	}
}

func TestPasswordlessMuxServesPages(t *testing.T) {
	resetPasswordlessStore()

	authInstance, err := newAuth()
	if err != nil {
		t.Fatal(err)
	}
	mux := newMux(authInstance)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "Login") {
		t.Fatalf("expected home page, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/auth/login", nil))
	if rec.Code != 200 {
		t.Fatalf("expected login page, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/dashboard", nil))
	if rec.Code == 200 {
		t.Fatal("expected redirect for unauthenticated dashboard")
	}
}
