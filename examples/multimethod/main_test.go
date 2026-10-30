package main

import (
	"context"
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
