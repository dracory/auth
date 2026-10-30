package main

import (
	"context"
	"testing"

	authtypes "github.com/dracory/auth/types"
)

func resetImpersonationStore() {
	store.mu.Lock()
	defer store.mu.Unlock()

	adminUser := &userAccount{ID: "admin-1", Username: "admin@example.com", Password: "adminpassword", FirstName: "Admin", LastName: "User", IsAdmin: true}
	targetUser := &userAccount{ID: "user-2", Username: "john@example.com", Password: "userpassword", FirstName: "John", LastName: "Doe", IsAdmin: false}

	store.users = map[string]*userAccount{
		adminUser.Username:  adminUser,
		targetUser.Username: targetUser,
	}
	store.usersByID = map[string]*userAccount{
		adminUser.ID:  adminUser,
		targetUser.ID: targetUser,
	}
	store.sessions = make(map[string]string)
	store.impersonating = make(map[string]string)
}

func TestImpersonationPermissions(t *testing.T) {
	resetImpersonationStore()
	ctx := context.Background()

	// Admin should be able to impersonate user-2
	can, err := store.canImpersonate(ctx, "admin-1", "user-2")
	if err != nil || !can {
		t.Fatalf("expected admin-1 can impersonate user-2, got can=%v, err=%v", can, err)
	}

	// Normal user should NOT be able to impersonate admin-1
	can, err = store.canImpersonate(ctx, "user-2", "admin-1")
	if err != nil || can {
		t.Fatalf("expected user-2 cannot impersonate admin-1, got can=%v, err=%v", can, err)
	}
}

func TestImpersonationStartAndStop(t *testing.T) {
	resetImpersonationStore()
	ctx := context.Background()

	if err := store.impersonationStart(ctx, "admin-1", "user-2"); err != nil {
		t.Fatalf("impersonationStart failed: %v", err)
	}

	if target := store.impersonating["admin-1"]; target != "user-2" {
		t.Fatalf("expected impersonating target user-2, got %q", target)
	}

	if err := store.impersonationStop(ctx, "admin-1", "user-2"); err != nil {
		t.Fatalf("impersonationStop failed: %v", err)
	}

	if _, ok := store.impersonating["admin-1"]; ok {
		t.Fatalf("expected impersonating record removed after stop")
	}
}

func TestImpersonationStoreAuthTokenAndLogout(t *testing.T) {
	resetImpersonationStore()
	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	token := "imp-token-123"
	if err := store.storeAuthToken(ctx, token, "admin-1", opts); err != nil {
		t.Fatalf("storeAuthToken failed: %v", err)
	}

	id, err := store.findByAuthToken(ctx, token, opts)
	if err != nil || id != "admin-1" {
		t.Fatalf("findByAuthToken failed, got %q (err: %v)", id, err)
	}

	if err := store.logout(ctx, "admin-1", opts); err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	if _, err := store.findByAuthToken(ctx, token, opts); err == nil {
		t.Fatalf("expected error finding token after logout")
	}
}
