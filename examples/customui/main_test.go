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
