package rules

import (
	"net/http"
	"testing"
)

func TestProfileCompletionRule_PassesForAuthenticatedRequest(t *testing.T) {
	req, err := http.NewRequest("GET", "/auth/register", nil)
	if err != nil {
		t.Fatal(err)
	}

	rule := NewProfileCompletionRule(func(r *http.Request) string { return "user-1" }, req)
	if rule.Fails() {
		t.Fatal("expected rule to pass for an authenticated request")
	}
	if rule.UserID() != "user-1" {
		t.Fatalf("expected user-1, got %q", rule.UserID())
	}
}

func TestProfileCompletionRule_FailsForGuest(t *testing.T) {
	req, err := http.NewRequest("GET", "/auth/register", nil)
	if err != nil {
		t.Fatal(err)
	}

	rule := NewProfileCompletionRule(func(r *http.Request) string { return "" }, req)
	if !rule.Fails() {
		t.Fatal("expected rule to fail for an unauthenticated request")
	}
	if rule.UserID() != "" {
		t.Fatalf("expected empty userID, got %q", rule.UserID())
	}
}

func TestProfileCompletionRule_FailsWithoutResolver(t *testing.T) {
	req, err := http.NewRequest("GET", "/auth/register", nil)
	if err != nil {
		t.Fatal(err)
	}

	if rule := NewProfileCompletionRule(nil, req); !rule.Fails() {
		t.Fatal("expected rule to fail when no resolver is configured")
	}
}
