package main

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	authtypes "github.com/dracory/auth/types"
)

// resetPasswordStore is a helper to reset the global store between tests.
func resetPasswordStore() {
	passwordStore = &passwordMemoryStore{
		usersByName:   make(map[string]*passwordUser),
		sessions:      make(map[string]string),
		tempKeys:      make(map[string]string),
		nextUserIndex: 1,
	}
}

func TestUsernamePasswordRegisterAndLogin(t *testing.T) {
	resetPasswordStore()

	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	username := "user@example.com"
	password := "P@ssw0rd!"

	if err := passwordStore.userRegister(ctx, username, password, "John", "Doe", opts); err != nil {
		t.Fatalf("userRegister unexpected error: %v", err)
	}

	userID, err := passwordStore.userLogin(ctx, username, password, opts)
	if err != nil {
		t.Fatalf("userLogin unexpected error: %v", err)
	}
	if userID == "" {
		t.Fatalf("expected non-empty user ID")
	}

	// Wrong password should fail
	if _, err := passwordStore.userLogin(ctx, username, "wrong", opts); err == nil {
		t.Fatalf("expected error when logging in with wrong password, got nil")
	}

	// Duplicate registration should fail
	if err := passwordStore.userRegister(ctx, username, password, "John", "Doe", opts); err == nil {
		t.Fatalf("expected error when registering duplicate username, got nil")
	}
}

func TestUsernamePasswordFindByUsernameAndChangePassword(t *testing.T) {
	resetPasswordStore()

	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	username := "user@example.com"
	password := "OldPass1!"
	firstName := "John"
	lastName := "Doe"

	if err := passwordStore.userRegister(ctx, username, password, firstName, lastName, opts); err != nil {
		t.Fatalf("userRegister unexpected error: %v", err)
	}

	userID, err := passwordStore.userFindByUsername(ctx, username, firstName, lastName, opts)
	if err != nil {
		t.Fatalf("userFindByUsername unexpected error: %v", err)
	}
	if userID == "" {
		t.Fatalf("expected non-empty user ID from userFindByUsername")
	}

	// Change password
	newPassword := "NewPass2!"
	if err := passwordStore.userPasswordChange(ctx, username, newPassword, opts); err != nil {
		t.Fatalf("userPasswordChange unexpected error: %v", err)
	}

	// Old password should now fail
	if _, err := passwordStore.userLogin(ctx, username, password, opts); err == nil {
		t.Fatalf("expected error when logging in with old password after change, got nil")
	}

	// New password should succeed
	if _, err := passwordStore.userLogin(ctx, username, newPassword, opts); err != nil {
		t.Fatalf("expected login success with new password, got error: %v", err)
	}
}

func TestUsernamePasswordAuthTokenAndLogout(t *testing.T) {
	resetPasswordStore()

	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	userID1 := "user-1"
	userID2 := "user-2"

	_ = passwordStore.storeAuthToken(ctx, "t1", userID1, opts)
	_ = passwordStore.storeAuthToken(ctx, "t2", userID1, opts)
	_ = passwordStore.storeAuthToken(ctx, "t3", userID2, opts)

	if err := passwordStore.logout(ctx, userID1, opts); err != nil {
		t.Fatalf("logout unexpected error: %v", err)
	}

	if _, err := passwordStore.findByAuthToken(ctx, "t1", opts); err == nil {
		t.Fatalf("expected error for token t1 after logout, got nil")
	}
	if _, err := passwordStore.findByAuthToken(ctx, "t2", opts); err == nil {
		t.Fatalf("expected error for token t2 after logout, got nil")
	}

	if _, err := passwordStore.findByAuthToken(ctx, "t3", opts); err != nil {
		t.Fatalf("expected token t3 to remain valid, got error: %v", err)
	}
}

func TestUsernamePasswordTempKeys(t *testing.T) {
	resetPasswordStore()

	key := "reset-123"
	val := "user@example.com"

	if err := passwordStore.tempKeySet(key, val, 60); err != nil {
		t.Fatalf("tempKeySet unexpected error: %v", err)
	}

	got, err := passwordStore.tempKeyGet(key)
	if err != nil {
		t.Fatalf("tempKeyGet unexpected error: %v", err)
	}
	if got != val {
		t.Fatalf("expected value %q, got %q", val, got)
	}

	if _, err := passwordStore.tempKeyGet("missing"); err == nil {
		t.Fatalf("expected error for missing key, got nil")
	}
}

func TestPasswordStoreDisplayNameUsesRegisteredName(t *testing.T) {
	resetPasswordStore()

	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	username := "user@example.com"
	password := "P@ssw0rd!"
	firstName := "Alice"
	lastName := "Smith"

	if err := passwordStore.userRegister(ctx, username, password, firstName, lastName, opts); err != nil {
		t.Fatalf("userRegister unexpected error: %v", err)
	}

	userID, err := passwordStore.userLogin(ctx, username, password, opts)
	if err != nil {
		t.Fatalf("userLogin unexpected error: %v", err)
	}

	display := passwordStore.displayName(userID)
	if !strings.Contains(display, firstName) || !strings.Contains(display, lastName) {
		t.Fatalf("expected display name to contain first and last name, got %q", display)
	}
}

func TestDisplayNameVariants(t *testing.T) {
	resetPasswordStore()

	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	if err := passwordStore.userRegister(ctx, "named@example.com", "pass", "John", "Doe", opts); err != nil {
		t.Fatal(err)
	}
	if err := passwordStore.userRegister(ctx, "unnamed@example.com", "pass", "", "", opts); err != nil {
		t.Fatal(err)
	}

	namedID, err := passwordStore.userLogin(ctx, "named@example.com", "pass", opts)
	if err != nil {
		t.Fatal(err)
	}
	unnamedID, err := passwordStore.userLogin(ctx, "unnamed@example.com", "pass", opts)
	if err != nil {
		t.Fatal(err)
	}

	if got := passwordStore.displayName(namedID); got != "John Doe" {
		t.Fatalf("expected full name, got %q", got)
	}
	if got := passwordStore.displayName(unnamedID); got != "unnamed@example.com" {
		t.Fatalf("expected username fallback, got %q", got)
	}
	if got := passwordStore.displayName("missing"); got != "missing" {
		t.Fatalf("expected id fallback, got %q", got)
	}
}

func TestTempKeyRoundTripAndLogout(t *testing.T) {
	resetPasswordStore()
	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	if err := passwordStore.tempKeySet("k1", "v1", 60); err != nil {
		t.Fatal(err)
	}
	if v, err := passwordStore.tempKeyGet("k1"); err != nil || v != "v1" {
		t.Fatalf("expected v1, got %q %v", v, err)
	}
	if _, err := passwordStore.tempKeyGet("missing"); err == nil {
		t.Fatal("expected error for missing key")
	}

	if err := passwordStore.userRegister(ctx, "u@x.com", "pass", "A", "B", opts); err != nil {
		t.Fatal(err)
	}
	userID, _ := passwordStore.userLogin(ctx, "u@x.com", "pass", opts)
	if err := passwordStore.storeAuthToken(ctx, "tok1", userID, opts); err != nil {
		t.Fatal(err)
	}
	if err := passwordStore.logout(ctx, userID, opts); err != nil {
		t.Fatal(err)
	}
	if _, err := passwordStore.findByAuthToken(ctx, "tok1", opts); err == nil {
		t.Fatal("expected token revoked after logout")
	}
}

func TestExampleEmailSend(t *testing.T) {
	if err := exampleEmailSend(context.Background(), "a@b.com", "subj", "body"); err != nil {
		t.Fatal(err)
	}
}

func TestNewMuxServesPages(t *testing.T) {
	resetPasswordStore()

	authInstance, err := newAuth()
	if err != nil {
		t.Fatal(err)
	}
	mux := newMux(authInstance)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "Username") {
		t.Fatalf("expected home page, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/auth/login", nil))
	if rec.Code != 200 {
		t.Fatalf("expected login page, got %d", rec.Code)
	}

	// Dashboard requires auth — expect a redirect to login.
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/dashboard", nil))
	if rec.Code == 200 {
		t.Fatal("expected redirect for unauthenticated dashboard")
	}
}
