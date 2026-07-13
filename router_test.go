package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracory/auth/internal/testutils"
)

func TestRouter_UnknownPathRedirectsToLogin(t *testing.T) {
	config := testutils.NewUsernameAndPasswordConfigForTest()
	authShared, err := NewUsernameAndPasswordAuth(config)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	recorder := httptest.NewRecorder()

	authShared.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, status)
	}

	location := recorder.Header().Get("Location")
	expectedLocation := authShared.LinkLogin()
	if location != expectedLocation {
		t.Fatalf("expected redirect to %q, got %q", expectedLocation, location)
	}
}

func TestRouter_SuffixDoesNotFalselyMatchLoginPath(t *testing.T) {
	config := testutils.NewUsernameAndPasswordConfigForTest()
	authShared, err := NewUsernameAndPasswordAuth(config)
	if err != nil {
		t.Fatal(err)
	}

	// "/some/prefix/mylogin" ends with "login" but is NOT the login path.
	// It must NOT be routed to the login page.
	req := httptest.NewRequest(http.MethodGet, "/some/prefix/mylogin", nil)
	recorder := httptest.NewRecorder()

	authShared.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d (not found redirect), got %d", http.StatusTemporaryRedirect, status)
	}
}

func TestRouter_LoginPathServesLoginPage(t *testing.T) {
	config := testutils.NewUsernameAndPasswordConfigForTest()
	authShared, err := NewUsernameAndPasswordAuth(config)
	if err != nil {
		t.Fatal(err)
	}

	// Use the full login URL, as done in other tests
	loginURL := authShared.LinkLogin()
	req := httptest.NewRequest(http.MethodGet, loginURL, nil)
	recorder := httptest.NewRecorder()

	authShared.Router().ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, status)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "<span>Log in</span>") {
		t.Fatalf("expected login page HTML to contain %q, got %s", "<span>Log in</span>", body)
	}
}
