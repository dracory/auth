package api_login

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
)

// recordingHooks captures ObservabilityHooks calls for assertions.
type recordingHooks struct {
	types.NoopObservabilityHooks
	attempts        []bool
	sessionsCreated []string
}

func (h *recordingHooks) RecordLoginAttempt(method string, success bool, err error) {
	h.attempts = append(h.attempts, success)
}

func (h *recordingHooks) RecordSessionCreated(userID string) {
	h.sessionsCreated = append(h.sessionsCreated, userID)
}

func postFormValues() url.Values {
	return url.Values{
		"email":    {"user@example.com"},
		"password": {"secret123"},
	}
}

// TestApiLoginWithAuth_PostAuth_SessionCookieAndRedirect verifies that a
// hooks-configured password login funnels through PostAuth: session from
// the hook, cookie from the hook, redirect from the hook — and the
// host's find-or-create is bypassed (credential login never creates).
func TestApiLoginWithAuth_PostAuth_SessionCookieAndRedirect(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	testutils.SetPasswordlessForTest(a, false)
	testutils.SetUseCookiesForTest(a, true)

	cookieToken := ""
	var findOrCreateCalled bool

	session := utils.NewMemSession()
	session.SetKey("sess-key-123")
	session.SetUserID("u1")

	testutils.SetFuncUserLoginForTest(a, func(ctx context.Context, username, password string, options types.UserAuthOptions) (string, error) {
		return "u1", nil
	})
	testutils.SetPostAuthHooksForTest(a, types.PostAuthHooks{
		UserFindByEmailOrCreate: func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
			findOrCreateCalled = true
			return "other-user", nil
		},
		SessionCreate: func(ctx context.Context, userID string, options types.UserAuthOptions) (types.SessionInterface, error) {
			return session, nil
		},
		SetAuthCookie: func(w http.ResponseWriter, r *http.Request, token string) {
			cookieToken = token
		},
		RedirectURL: func(ctx context.Context, userID string) string {
			return "/dash"
		},
	})

	obs := &recordingHooks{}
	testutils.SetObservabilityHooksForTest(a, obs)

	recorder, req := makePostRequest(t, "/auth/login", postFormValues())
	ApiLoginWithAuth(recorder, req, a)

	body := recorder.Body.String()
	if !strings.Contains(body, `"token":"sess-key-123"`) {
		t.Fatalf("expected session key in response, got %q", body)
	}
	if !strings.Contains(body, `"redirect":"/dash"`) {
		t.Fatalf("expected pipeline redirect in response, got %q", body)
	}
	if cookieToken != "sess-key-123" {
		t.Fatalf("expected auth cookie via hook, got %q", cookieToken)
	}
	if findOrCreateCalled {
		t.Fatal("host find-or-create must not run on credential login")
	}
	if len(obs.attempts) != 1 || !obs.attempts[0] {
		t.Fatalf("expected one successful login attempt, got %+v", obs.attempts)
	}
	if len(obs.sessionsCreated) != 1 || obs.sessionsCreated[0] != "u1" {
		t.Fatalf("expected session created for u1, got %+v", obs.sessionsCreated)
	}
}

// TestApiLoginWithAuth_PostAuth_NoCookieWhenUseCookiesFalse verifies
// UseCookies=false suppresses both the hook cookie and the fallback —
// token-only API hosts must not receive auth cookies.
func TestApiLoginWithAuth_PostAuth_NoCookieWhenUseCookiesFalse(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	testutils.SetPasswordlessForTest(a, false)
	testutils.SetUseCookiesForTest(a, false)

	cookieSet := false
	session := utils.NewMemSession()
	session.SetKey("tok-only")
	session.SetUserID("u1")

	testutils.SetFuncUserLoginForTest(a, func(ctx context.Context, username, password string, options types.UserAuthOptions) (string, error) {
		return "u1", nil
	})
	testutils.SetPostAuthHooksForTest(a, types.PostAuthHooks{
		SessionCreate: func(ctx context.Context, userID string, options types.UserAuthOptions) (types.SessionInterface, error) {
			return session, nil
		},
		SetAuthCookie: func(w http.ResponseWriter, r *http.Request, token string) {
			cookieSet = true
		},
	})

	recorder, req := makePostRequest(t, "/auth/login", postFormValues())
	ApiLoginWithAuth(recorder, req, a)

	if cookieSet {
		t.Fatal("auth cookie hook must not run when UseCookies=false")
	}
	if got := recorder.Header().Get("Set-Cookie"); got != "" {
		t.Fatalf("expected no Set-Cookie header, got %q", got)
	}
	if !strings.Contains(recorder.Body.String(), `"token":"tok-only"`) {
		t.Fatalf("expected token in body, got %q", recorder.Body.String())
	}
}

// TestApiLoginWithAuth_PostAuth_EmailAllowedGatesBeforeCredentials
// verifies the allowlist runs before password verification — a denied
// email never reaches FuncUserLogin.
func TestApiLoginWithAuth_PostAuth_EmailAllowedGatesBeforeCredentials(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	testutils.SetPasswordlessForTest(a, false)

	loginCalled := false
	testutils.SetFuncUserLoginForTest(a, func(ctx context.Context, username, password string, options types.UserAuthOptions) (string, error) {
		loginCalled = true
		return "u1", nil
	})
	testutils.SetPostAuthHooksForTest(a, types.PostAuthHooks{
		EmailAllowed: func(ctx context.Context, email string) (bool, string) {
			return false, "denied"
		},
	})

	recorder, req := makePostRequest(t, "/auth/login", postFormValues())
	ApiLoginWithAuth(recorder, req, a)

	if loginCalled {
		t.Fatal("FuncUserLogin must not run for a denied email")
	}
	if !strings.Contains(recorder.Body.String(), "denied") {
		t.Fatalf("expected deny message, got %q", recorder.Body.String())
	}
}

// TestApiLoginWithAuth_PostAuth_UserInactive verifies the pipeline's
// active check rejects the login and records a failed attempt.
func TestApiLoginWithAuth_PostAuth_UserInactive(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	testutils.SetPasswordlessForTest(a, false)

	testutils.SetFuncUserLoginForTest(a, func(ctx context.Context, username, password string, options types.UserAuthOptions) (string, error) {
		return "u1", nil
	})
	testutils.SetPostAuthHooksForTest(a, types.PostAuthHooks{
		UserIsActive: func(ctx context.Context, userID string, options types.UserAuthOptions) bool {
			return false
		},
	})

	obs := &recordingHooks{}
	testutils.SetObservabilityHooksForTest(a, obs)

	recorder, req := makePostRequest(t, "/auth/login", postFormValues())
	ApiLoginWithAuth(recorder, req, a)

	if !strings.Contains(recorder.Body.String(), types.MsgUserInactive) {
		t.Fatalf("expected inactive-user error, got %q", recorder.Body.String())
	}
	if len(obs.attempts) != 1 || obs.attempts[0] {
		t.Fatalf("expected one failed login attempt, got %+v", obs.attempts)
	}
	if len(obs.sessionsCreated) != 0 {
		t.Fatalf("expected no session created, got %+v", obs.sessionsCreated)
	}
}
