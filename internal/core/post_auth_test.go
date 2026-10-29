package core_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/dracory/auth/internal/core"
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
)

type ruleStub struct {
	pass bool
	msg  string
}

func (r ruleStub) Passes() bool        { return r.pass }
func (r ruleStub) FailMessage() string { return r.msg }

func postAuth(t *testing.T, hooks types.PostAuthHooks, backURL string) (*core.PostAuthResult, error) {
	req := httptest.NewRequest("POST", "/x", nil)
	rec := httptest.NewRecorder()
	return core.PostAuth(context.Background(), rec, req, "u@x.com", backURL, core.PostAuthDeps{Hooks: hooks})
}

func TestPostAuth_FullSuccess(t *testing.T) {
	session := utils.NewMemSession()
	session.SetKey("sess-key")
	res, err := postAuth(t, types.PostAuthHooks{
		EmailAllowed:            func(ctx context.Context, e string) (bool, string) { return true, "" },
		UserFindByEmailOrCreate: func(ctx context.Context, e string, o types.UserAuthOptions) (string, error) { return "u1", nil },
		UserIsActive:            func(ctx context.Context, u string, o types.UserAuthOptions) bool { return true },
		SessionCreate: func(ctx context.Context, u string, o types.UserAuthOptions) (types.SessionInterface, error) {
			return session, nil
		},
		RedirectURL: func(ctx context.Context, u string) string { return "/home" },
	}, "")
	if err != nil || res.Token != "sess-key" || res.Redirect != "/home" {
		t.Fatalf("got %v %+v", err, res)
	}
}

func TestPostAuth_EmailDeniedAbortsBeforeUserCreate(t *testing.T) {
	createCalled := false
	_, err := postAuth(t, types.PostAuthHooks{
		EmailAllowed: func(ctx context.Context, e string) (bool, string) { return false, "denied" },
		UserFindByEmailOrCreate: func(ctx context.Context, e string, o types.UserAuthOptions) (string, error) {
			createCalled = true
			return "", nil
		},
	}, "")
	if err == nil || err.Error() != "denied" {
		t.Fatalf("expected deny, got %v", err)
	}
	if createCalled {
		t.Fatal("user create must not run after allowlist deny")
	}
}

func TestPostAuth_InactiveUserAborts(t *testing.T) {
	sessionCreated := false
	_, err := postAuth(t, types.PostAuthHooks{
		UserFindByEmailOrCreate: func(ctx context.Context, e string, o types.UserAuthOptions) (string, error) { return "u1", nil },
		UserIsActive:            func(ctx context.Context, u string, o types.UserAuthOptions) bool { return false },
		SessionCreate: func(ctx context.Context, u string, o types.UserAuthOptions) (types.SessionInterface, error) {
			sessionCreated = true
			return utils.NewMemSession(), nil
		},
	}, "")
	if err == nil || sessionCreated {
		t.Fatalf("expected inactive abort, got %v", err)
	}
}

func TestPostAuth_RegisterRuleShortCircuits(t *testing.T) {
	second := false
	_, err := postAuth(t, types.PostAuthHooks{
		EmailAllowed:            func(ctx context.Context, e string) (bool, string) { return true, "" },
		UserFindByEmailOrCreate: func(ctx context.Context, e string, o types.UserAuthOptions) (string, error) { return "u1", nil },
		RegisterRules: func(ctx context.Context, e string) []types.PostAuthRule {
			return []types.PostAuthRule{ruleStub{false, "first failed"}, ruleStub{false, "second failed"}}
		},
	}, "")
	if err == nil || err.Error() != "first failed" {
		t.Fatalf("expected first rule short-circuit, got %v", err)
	}
	_ = second
}

func TestPostAuth_OpenRedirectRejected(t *testing.T) {
	res, err := postAuth(t, types.PostAuthHooks{
		RedirectURL: func(ctx context.Context, u string) string { return "/home" },
	}, "//evil.com")
	if err != nil || res.Redirect != "/home" {
		t.Fatalf("expected external backUrl ignored, got %v %+v", err, res)
	}
}

func TestPostAuth_RelativeBackURLHonored(t *testing.T) {
	res, err := postAuth(t, types.PostAuthHooks{}, "/dashboard")
	if err != nil || res.Redirect != "/dashboard" {
		t.Fatalf("expected backUrl honored, got %v %+v", err, res)
	}
}

func TestIsSafeBackURL(t *testing.T) {
	cases := map[string]bool{
		"/dash": true, "//evil.com": false, "https://evil.com": false,
		"": false, "/login": true,
	}
	for u, want := range cases {
		if utils.IsSafeBackURL(u, "app.test") != want {
			t.Fatalf("IsSafeBackURL(%q) != %v", u, want)
		}
	}
	if !utils.IsSafeBackURL("https://app.test/x", "app.test") {
		t.Fatal("same-origin absolute URL should be safe")
	}
}
