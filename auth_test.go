package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCurrentUserID_EmptyContextReturnsEmptyString(t *testing.T) {
	auth := &Auth{}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	userID := auth.GetCurrentUserID(req)

	if userID != "" {
		t.Fatalf("expected empty user ID, got %q", userID)
	}
}

func TestGetCurrentUserID_ReturnsUserIDFromContext(t *testing.T) {
	auth := &Auth{}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), AuthenticatedUserID{}, "12345")
	req = req.WithContext(ctx)

	userID := auth.GetCurrentUserID(req)

	if userID != "12345" {
		t.Fatalf("expected user ID %q, got %q", "12345", userID)
	}
}

func TestLinkAddsSlashWhenMissing(t *testing.T) {
	endpoint := "http://example.com/auth"
	uri := PathApiLogin

	got := link(endpoint, uri)
	want := endpoint + "/" + uri

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestLinkKeepsTrailingSlash(t *testing.T) {
	endpoint := "http://example.com/auth/"
	uri := PathApiLogin

	got := link(endpoint, uri)
	want := endpoint + uri

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestAuthLinkHelpersRespectEndpointTrailingSlash(t *testing.T) {
	authWithSlash := &Auth{endpoint: "http://localhost/auth/"}

	if got, want := authWithSlash.LinkApiLogin(), "http://localhost/auth/"+PathApiLogin; got != want {
		t.Fatalf("LinkApiLogin with trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authWithSlash.LinkLogin(), "http://localhost/auth/"+PathLogin; got != want {
		t.Fatalf("LinkLogin with trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authWithSlash.LinkRegister(), "http://localhost/auth/"+PathRegister; got != want {
		t.Fatalf("LinkRegister with trailing slash: expected %q, got %q", want, got)
	}
}
