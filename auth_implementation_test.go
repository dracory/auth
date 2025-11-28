package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dracory/auth/internal/links"
)

func TestGetCurrentUserID_EmptyContextReturnsEmptyString(t *testing.T) {
	auth := &authImplementation{}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	userID := auth.GetCurrentUserID(req)

	if userID != "" {
		t.Fatalf("expected empty user ID, got %q", userID)
	}
}

func TestGetCurrentUserID_ReturnsUserIDFromContext(t *testing.T) {
	auth := &authImplementation{}

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

	got := links.Join(endpoint, uri)
	want := endpoint + "/" + uri

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestLinkKeepsTrailingSlash(t *testing.T) {
	endpoint := "http://example.com/auth/"
	uri := PathApiLogin

	got := links.Join(endpoint, uri)
	want := endpoint + uri

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestAuthLinkHelpersRespectEndpointTrailingSlash(t *testing.T) {
	authWithSlash := &authImplementation{endpoint: "http://localhost/auth/"}

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

func TestAuthLinkHelpers_NoTrailingSlash(t *testing.T) {
	authNoSlash := &authImplementation{endpoint: "http://localhost/auth"}

	if got, want := authNoSlash.LinkApiLoginCodeVerify(), "http://localhost/auth/"+PathApiLoginCodeVerify; got != want {
		t.Fatalf("LinkApiLoginCodeVerify without trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authNoSlash.LinkApiLogout(), "http://localhost/auth/"+PathApiLogout; got != want {
		t.Fatalf("LinkApiLogout without trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authNoSlash.LinkApiRegister(), "http://localhost/auth/"+PathApiRegister; got != want {
		t.Fatalf("LinkApiRegister without trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authNoSlash.LinkApiRegisterCodeVerify(), "http://localhost/auth/"+PathApiRegisterCodeVerify; got != want {
		t.Fatalf("LinkApiRegisterCodeVerify without trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authNoSlash.LinkApiPasswordRestore(), "http://localhost/auth/"+PathApiRestorePassword; got != want {
		t.Fatalf("LinkApiPasswordRestore without trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authNoSlash.LinkApiPasswordReset(), "http://localhost/auth/"+PathApiResetPassword; got != want {
		t.Fatalf("LinkApiPasswordReset without trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authNoSlash.LinkLoginCodeVerify(), "http://localhost/auth/"+PathLoginCodeVerify; got != want {
		t.Fatalf("LinkLoginCodeVerify without trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authNoSlash.LinkLogout(), "http://localhost/auth/"+PathLogout; got != want {
		t.Fatalf("LinkLogout without trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authNoSlash.LinkPasswordRestore(), "http://localhost/auth/"+PathPasswordRestore; got != want {
		t.Fatalf("LinkPasswordRestore without trailing slash: expected %q, got %q", want, got)
	}

	token := "abc123"
	if got, want := authNoSlash.LinkPasswordReset(token), "http://localhost/auth/"+PathPasswordReset+"?t="+token; got != want {
		t.Fatalf("LinkPasswordReset without trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authNoSlash.LinkRegisterCodeVerify(), "http://localhost/auth/"+PathRegisterCodeVerify; got != want {
		t.Fatalf("LinkRegisterCodeVerify without trailing slash: expected %q, got %q", want, got)
	}
}

func TestAuthLinkRedirectOnSuccess(t *testing.T) {
	authInstance := &authImplementation{urlRedirectOnSuccess: "/dashboard"}

	if got, want := authInstance.LinkRedirectOnSuccess(), "/dashboard"; got != want {
		t.Fatalf("LinkRedirectOnSuccess: expected %q, got %q", want, got)
	}
}

func TestRegistrationEnableDisableToggle(t *testing.T) {
	var authInstance authImplementation

	if authInstance.enableRegistration {
		t.Fatalf("expected enableRegistration to be false by default")
	}

	authInstance.RegistrationEnable()
	if !authInstance.enableRegistration {
		t.Fatalf("expected enableRegistration to be true after RegistrationEnable")
	}

	authInstance.RegistrationDisable()
	if authInstance.enableRegistration {
		t.Fatalf("expected enableRegistration to be false after RegistrationDisable")
	}
}
