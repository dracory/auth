package main

import (
	"context"
	"strings"
	"testing"

	authtypes "github.com/dracory/auth/types"
)

func resetMagicLinkStore() {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.usersByEmail = make(map[string]*magicLinkUser)
	store.sessions = make(map[string]string)
	store.tempKeys = make(map[string]string)
}

func TestMagicLinkRegisterAndFindUser(t *testing.T) {
	resetMagicLinkStore()

	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	email := "magic@example.com"
	if err := store.registerUser(ctx, email, "Magic", "Link", opts); err != nil {
		t.Fatalf("registerUser unexpected error: %v", err)
	}

	userID, err := store.findUserByEmail(ctx, email, opts)
	if err != nil {
		t.Fatalf("findUserByEmail unexpected error: %v", err)
	}
	if userID == "" {
		t.Fatalf("expected non-empty user ID")
	}

	display := store.displayName(userID)
	if !strings.Contains(display, "Magic") || !strings.Contains(display, "Link") {
		t.Fatalf("expected display name to contain Magic Link, got %q", display)
	}
}

func TestMagicLinkAuthTokenLifecycle(t *testing.T) {
	resetMagicLinkStore()

	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	userID := "user-100"
	token := "token-magic-123"

	if err := store.storeAuthToken(ctx, token, userID, opts); err != nil {
		t.Fatalf("storeAuthToken unexpected error: %v", err)
	}

	gotID, err := store.findByAuthToken(ctx, token, opts)
	if err != nil || gotID != userID {
		t.Fatalf("findByAuthToken failed, got %q (err: %v)", gotID, err)
	}

	if err := store.logout(ctx, userID, opts); err != nil {
		t.Fatalf("logout unexpected error: %v", err)
	}

	if _, err := store.findByAuthToken(ctx, token, opts); err == nil {
		t.Fatalf("expected error after logout")
	}
}

func TestMagicLinkTempKeys(t *testing.T) {
	resetMagicLinkStore()

	key := "magic-key-1"
	val := "magic@example.com"

	if err := store.tempKeySet(key, val, 300); err != nil {
		t.Fatalf("tempKeySet unexpected error: %v", err)
	}

	got, err := store.tempKeyGet(key)
	if err != nil || got != val {
		t.Fatalf("tempKeyGet failed, got %q (err: %v)", got, err)
	}
}
