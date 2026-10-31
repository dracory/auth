package api_authknight_callback

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/types"
)

// akAuthStub adapts the shared test double to AuthAuthKnightInterface.
type akAuthStub struct {
	types.AuthSharedInterface
	hooks types.PostAuthHooks
}

func (s akAuthStub) IsAuthKnight() bool                            { return true }
func (s akAuthStub) LinkAuthKnightLogin(b, n string) string        { return "" }
func (s akAuthStub) LinkAuthKnightRedirect(r *http.Request) string { return "" }
func (s akAuthStub) LinkAuthKnightCallback() string                { return "" }
func (s akAuthStub) GetAuthKnightHTTPTimeout() time.Duration       { return time.Second }
func (s akAuthStub) GetPostAuthHooks() types.PostAuthHooks         { return s.hooks }

func (s akAuthStub) GetAuthKnightUserFindByEmail() func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
	return func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
		return "u1", nil
	}
}

func (s akAuthStub) GetAuthKnightUserRegister() func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (string, error) {
	return func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (string, error) {
		return "u2", nil
	}
}

func (s akAuthStub) GetAuthKnightRedirectURL() func(ctx context.Context, userID string) string {
	return func(ctx context.Context, userID string) string { return "/" }
}

func TestApiAuthKnightCallbackWithAuth_MissingOnce(t *testing.T) {
	a := akAuthStub{AuthSharedInterface: testutils.NewAuthSharedForTest()}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/authknight-callback", nil)
	ApiAuthKnightCallbackWithAuth(rec, req, a)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected redirect on missing once, got %d %q", rec.Code, rec.Body.String())
	}
}

func TestApiAuthKnightCallbackWithAuth_PostAuthPath(t *testing.T) {
	a := akAuthStub{
		AuthSharedInterface: testutils.NewAuthSharedForTest(),
		hooks: types.PostAuthHooks{
			EmailAllowed: func(ctx context.Context, email string) (bool, string) {
				return true, ""
			},
		},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/authknight-callback", nil)
	ApiAuthKnightCallbackWithAuth(rec, req, a)

	// Post-auth path also redirects to the fallback on a missing once token.
	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected redirect on missing once, got %d %q", rec.Code, rec.Body.String())
	}
}
