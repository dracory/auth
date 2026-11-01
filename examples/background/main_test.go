package main

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	authtypes "github.com/dracory/auth/types"
)

func resetStore() {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.users = make(map[string]*bgUser)
	store.sessions = make(map[string]string)
	store.tempKeys = make(map[string]string)
	store.nextIndex = 1
}

func TestBackgroundStoreHelpers(t *testing.T) {
	resetStore()
	ctx := context.Background()
	opts := authtypes.UserAuthOptions{}

	store.users["a@b.com"] = &bgUser{ID: "u1", Username: "a@b.com", Password: "pw", FirstName: "A", LastName: "B"}
	store.users["noname@x.com"] = &bgUser{ID: "u2", Username: "noname@x.com", Password: "pw"}

	if id, err := store.userLogin(ctx, "a@b.com", "pw", opts); err != nil || id != "u1" {
		t.Fatalf("userLogin failed: %v %q", err, id)
	}
	if _, err := store.userLogin(ctx, "a@b.com", "bad", opts); err == nil {
		t.Fatal("expected error for bad password")
	}
	if _, err := store.userLogin(ctx, "missing", "pw", opts); err == nil {
		t.Fatal("expected error for unknown user")
	}

	if id, err := store.userFindByUsername(ctx, "u2", "", "", opts); err == nil {
		if id != "" {
			t.Fatalf("expected not found, got %q", id)
		}
	} else {
		// acceptable: either error or empty id
	}
	if id, err := store.userFindByUsername(ctx, "a@b.com", "", "", opts); err != nil || id != "u1" {
		t.Fatalf("userFindByUsername failed: %v %q", err, id)
	}

	if got := store.displayName("u1"); got != "A B" {
		t.Fatalf("expected full name, got %q", got)
	}
	if got := store.displayName("u2"); got != "noname@x.com" {
		t.Fatalf("expected username fallback, got %q", got)
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
	if id, err := store.findByAuthToken(ctx, "tok", opts); err != nil || id != "u1" {
		t.Fatalf("findByAuthToken failed: %v %q", err, id)
	}
	if err := store.logout(ctx, "u1", opts); err != nil {
		t.Fatal(err)
	}
	if _, err := store.findByAuthToken(ctx, "tok", opts); err == nil {
		t.Fatal("expected revoked token")
	}

	if err := emailSend(ctx, "u1", "s", "b"); err != nil {
		t.Fatal(err)
	}
}

func TestBackgroundMuxServesPages(t *testing.T) {
	resetStore()

	authInstance, err := newAuth()
	if err != nil {
		t.Fatal(err)
	}
	mux := newMux(authInstance)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "Background") {
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
