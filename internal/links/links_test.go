package links

import "testing"

func TestApiImpersonateStart(t *testing.T) {
	got := ApiImpersonateStart("http://localhost/auth")
	expected := "http://localhost/auth/api/impersonate/start"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestApiImpersonateStartTrailingSlash(t *testing.T) {
	got := ApiImpersonateStart("http://localhost/auth/")
	expected := "http://localhost/auth/api/impersonate/start"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestApiImpersonateStop(t *testing.T) {
	got := ApiImpersonateStop("http://localhost/auth")
	expected := "http://localhost/auth/api/impersonate/stop"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestApiImpersonateStopTrailingSlash(t *testing.T) {
	got := ApiImpersonateStop("http://localhost/auth/")
	expected := "http://localhost/auth/api/impersonate/stop"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}
