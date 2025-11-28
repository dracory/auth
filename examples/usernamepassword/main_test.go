package main

import (
	"context"
	"testing"

	auth "github.com/dracory/auth"
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
	opts := auth.UserAuthOptions{}

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
	opts := auth.UserAuthOptions{}

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
	opts := auth.UserAuthOptions{}

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
