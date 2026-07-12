package auth

import (
	"context"
	"net/http"
	"testing"

	"github.com/dracory/auth/types"
)

func TestIsImpersonatingFalse(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	if IsImpersonating(req) {
		t.Fatal("expected IsImpersonating to return false when no context value")
	}
}

func TestIsImpersonatingTrue(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), types.IsImpersonating{}, true)
	req = req.WithContext(ctx)
	if !IsImpersonating(req) {
		t.Fatal("expected IsImpersonating to return true")
	}
}

func TestGetImpersonatorUserIDEmpty(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	if id := GetImpersonatorUserID(req); id != "" {
		t.Fatalf("expected empty impersonator ID, got %q", id)
	}
}

func TestGetImpersonatorUserIDSet(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), types.ImpersonatorUserID{}, "admin-123")
	req = req.WithContext(ctx)
	if id := GetImpersonatorUserID(req); id != "admin-123" {
		t.Fatalf("expected admin-123, got %q", id)
	}
}

func TestIsImpersonatingWrongType(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), types.IsImpersonating{}, "not-a-bool")
	req = req.WithContext(ctx)
	if IsImpersonating(req) {
		t.Fatal("expected false when context value is not bool")
	}
}

func TestGetImpersonatorUserIDWrongType(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), types.ImpersonatorUserID{}, 12345)
	req = req.WithContext(ctx)
	if id := GetImpersonatorUserID(req); id != "" {
		t.Fatalf("expected empty string for wrong type, got %q", id)
	}
}
