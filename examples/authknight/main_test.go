package main

import (
	"context"
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
