package e2e

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth"
	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/types"
)

// TestLogoutFlow: POST /auth/api/logout with a valid session cookie calls
// FuncUserLogout for that user and expires the auth cookie.
func TestLogoutFlow(t *testing.T) {
	var loggedOutUser string

	cfg := testutils.NewPasswordlessConfigForTest()
	cfg.FuncUserFindByAuthToken = func(ctx context.Context, token string, options types.UserAuthOptions) (string, error) {
		if token == "session-abc" {
			return "user-1", nil
		}
		return "", nil
	}
	cfg.FuncUserLogout = func(ctx context.Context, userID string, options types.UserAuthOptions) error {
		loggedOutUser = userID
		return nil
	}

	a, err := auth.NewOtpAuth(cfg)
	if err != nil {
		t.Fatalf("NewPasswordlessAuth: %v", err)
	}
	mux := a.Router()

	rec := postForm(mux, "/auth/api/logout", url.Values{},
		&http.Cookie{Name: a.GetCookieName(), Value: "session-abc"})

	if !strings.Contains(rec.Body.String(), `"status":"success"`) {
		t.Fatalf("expected logout success, got %q", rec.Body.String())
	}
	if loggedOutUser != "user-1" {
		t.Fatalf("expected FuncUserLogout for user-1, got %q", loggedOutUser)
	}
}
