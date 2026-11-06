package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
)

func TestSetters_BrandingAndTheme(t *testing.T) {
	a := &authImplementation{}

	a.SetThemeCSS("theme.css")
	a.SetCustomCSS("body{color:red}")
	a.SetBrandName("Acme")
	a.SetBrandTagline("Secure sign-in")
	a.SetBrandLogoURL("https://example.com/logo.png")

	if a.GetThemeCSS() != "theme.css" {
		t.Fatalf("unexpected theme CSS %q", a.GetThemeCSS())
	}
	if a.GetCustomCSS() != "body{color:red}" {
		t.Fatalf("unexpected custom CSS %q", a.GetCustomCSS())
	}
	if a.GetBrandName() != "Acme" {
		t.Fatalf("unexpected brand name %q", a.GetBrandName())
	}
	if a.GetBrandTagline() != "Secure sign-in" {
		t.Fatalf("unexpected brand tagline %q", a.GetBrandTagline())
	}
	if a.GetBrandLogoURL() != "https://example.com/logo.png" {
		t.Fatalf("unexpected brand logo URL %q", a.GetBrandLogoURL())
	}
}

func TestSetters_LoginAlternatives(t *testing.T) {
	a := &authImplementation{}

	if len(a.GetLoginAlternatives()) != 0 {
		t.Fatal("expected no alternatives by default")
	}

	alts := []types.LoginAlternative{
		{Label: "Use a password", URL: "/auth/login-password", Icon: "bi-key"},
	}
	a.SetLoginAlternatives(alts)

	got := a.GetLoginAlternatives()
	if len(got) != 1 || got[0].Label != "Use a password" || got[0].URL != "/auth/login-password" {
		t.Fatalf("unexpected alternatives %+v", got)
	}
}

func TestSetters_EmailSendThrottleConfig(t *testing.T) {
	a := &authImplementation{}

	cfg := utils.EmailSendThrottle{MaxSends: 1, WindowSeconds: 300}
	a.SetEmailSendThrottleConfig(cfg)

	got := a.GetEmailSendThrottleConfig()
	if got.MaxSends != 1 || got.WindowSeconds != 300 {
		t.Fatalf("unexpected throttle config %+v", got)
	}
}

func TestSetters_TemporaryKeyDelete(t *testing.T) {
	a := &authImplementation{}

	called := ""
	a.SetFuncTemporaryKeyDelete(func(key string) error {
		called = key
		return nil
	})

	if err := a.GetFuncTemporaryKeyDelete()("k1"); err != nil {
		t.Fatal(err)
	}
	if called != "k1" {
		t.Fatalf("expected delete fn to be invoked, got %q", called)
	}
}

func TestSetters_UserExists(t *testing.T) {
	a := &authImplementation{}

	a.SetFuncUserExists(func(ctx context.Context, email string, options types.UserAuthOptions) (bool, error) {
		return email == "exists@example.com", nil
	})

	ok, err := a.GetFuncUserExists()(context.Background(), "exists@example.com", types.UserAuthOptions{})
	if err != nil || !ok {
		t.Fatalf("expected user exists, got %v %v", ok, err)
	}
}

func TestSetters_EmailEnqueue(t *testing.T) {
	a := &authImplementation{}

	called := ""
	a.SetFuncEmailEnqueue(func(ctx context.Context, deliveryKey string) error {
		called = deliveryKey
		return nil
	})

	if err := a.GetFuncEmailEnqueue()(context.Background(), "dk1"); err != nil {
		t.Fatal(err)
	}
	if called != "dk1" {
		t.Fatalf("expected enqueue fn to be invoked, got %q", called)
	}
}

func TestSetters_RememberMe(t *testing.T) {
	a := &authImplementation{}

	if a.GetRememberMeEnabled() {
		t.Fatal("expected remember-me disabled by default")
	}
	if a.GetRememberMeDays() != 0 {
		t.Fatalf("expected 0 remember-me days by default, got %d", a.GetRememberMeDays())
	}
	if a.GetRememberCookieName() != "" {
		t.Fatalf("expected empty remember cookie name, got %q", a.GetRememberCookieName())
	}

	a.SetRememberMeEnabled(true)
	a.SetRememberMeDays(45)
	a.SetRememberCookieName("remember_me")

	if !a.GetRememberMeEnabled() {
		t.Fatal("expected remember-me enabled")
	}
	if a.GetRememberMeDays() != 45 {
		t.Fatalf("expected 45 remember-me days, got %d", a.GetRememberMeDays())
	}
	if a.GetRememberCookieName() != "remember_me" {
		t.Fatalf("unexpected remember cookie name %q", a.GetRememberCookieName())
	}
}

func TestGetSecureCookies(t *testing.T) {
	a := &authImplementation{}
	if a.GetSecureCookies() {
		t.Fatal("expected insecure cookies by default")
	}
	a.cookieConfig.Secure = "yes"
	if !a.GetSecureCookies() {
		t.Fatal("expected secure cookies")
	}
}

func TestSetters_PostAuthHooks(t *testing.T) {
	a := &authImplementation{}

	if got := a.GetPostAuthHooks(); got.EmailAllowed != nil {
		t.Fatal("expected no hooks by default")
	}

	hooks := types.PostAuthHooks{
		EmailAllowed: func(ctx context.Context, email string) (bool, string) {
			return true, ""
		},
	}
	a.SetPostAuthHooks(hooks)

	allowed, _ := a.GetPostAuthHooks().EmailAllowed(context.Background(), "a@b.com")
	if !allowed {
		t.Fatal("expected hook to allow email")
	}
}

func TestSetters_RateLimits(t *testing.T) {
	a := &authImplementation{}

	limits := map[string]int{"login": 3}
	a.SetRateLimits(limits)

	got := a.GetRateLimits()
	if got["login"] != 3 {
		t.Fatalf("expected login limit 3, got %v", got)
	}

	// Mutating the returned map must not leak into stored state.
	got["login"] = 99
	if a.GetRateLimits()["login"] != 3 {
		t.Fatal("expected returned limits to be a copy")
	}
}

func TestSetters_RateLimitsAppliesToLimiter(t *testing.T) {
	a := &authImplementation{
		rateLimiter: utils.NewInMemoryRateLimiter(10, time.Minute, time.Minute),
	}
	defer a.rateLimiter.Stop()

	a.SetRateLimits(map[string]int{"login": 2})
	if a.GetRateLimits()["login"] != 2 {
		t.Fatal("expected login limit 2")
	}

	// Third request within the window must be rejected.
	for i := 0; i < 2; i++ {
		res := a.rateLimiter.Check("1.2.3.4", "login")
		if !res.Allowed {
			t.Fatalf("expected request %d allowed, got %+v", i, res)
		}
	}
	if res := a.rateLimiter.Check("1.2.3.4", "login"); res.Allowed {
		t.Fatal("expected third request to be rate limited")
	}
}

func TestSetters_AuthKnight(t *testing.T) {
	a := &authImplementation{}

	a.SetAuthKnightUserFindByEmail(func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
		return "u1", nil
	})
	a.SetAuthKnightUserRegister(func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (string, error) {
		return "u2", nil
	})
	a.SetAuthKnightRedirectURL(func(ctx context.Context, userID string) string {
		return "/welcome"
	})

	if a.GetAuthKnightHTTPTimeout() != 10*time.Second {
		t.Fatalf("expected default 10s timeout, got %v", a.GetAuthKnightHTTPTimeout())
	}
	a.SetAuthKnightHTTPTimeout(3 * time.Second)
	if a.GetAuthKnightHTTPTimeout() != 3*time.Second {
		t.Fatalf("expected 3s timeout, got %v", a.GetAuthKnightHTTPTimeout())
	}

	if id, _ := a.GetAuthKnightUserFindByEmail()(context.Background(), "a@b.com", types.UserAuthOptions{}); id != "u1" {
		t.Fatalf("unexpected user id %q", id)
	}
	if id, _ := a.GetAuthKnightUserRegister()(context.Background(), "a@b.com", "A", "B", types.UserAuthOptions{}); id != "u2" {
		t.Fatalf("unexpected user id %q", id)
	}
	if u := a.GetAuthKnightRedirectURL()(context.Background(), "u1"); u != "/welcome" {
		t.Fatalf("unexpected redirect %q", u)
	}
}

func TestAuthenticateViaUsername_PostAuthDeny(t *testing.T) {
	a := &authImplementation{}
	a.SetPostAuthHooks(types.PostAuthHooks{
		EmailAllowed: func(ctx context.Context, email string) (bool, string) {
			return false, "not allowed"
		},
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	a.AuthenticateViaUsername(recorder, req, "a@b.com", "A", "B")

	if !strings.Contains(recorder.Body.String(), "not allowed") && recorder.Code == http.StatusOK {
		t.Fatalf("expected denial, got %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestAuthenticateViaUsername_PostAuthSuccess(t *testing.T) {
	a := &authImplementation{}
	a.SetPostAuthHooks(types.PostAuthHooks{
		UserFindByEmailOrCreate: func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
			return "u1", nil
		},
		SessionCreate: func(ctx context.Context, userID string, options types.UserAuthOptions) (types.SessionInterface, error) {
			s := utils.NewMemSession()
			s.SetUserID(userID)
			s.SetKey("session-key-1")
			return s, nil
		},
	})

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	a.AuthenticateViaUsername(recorder, req, "a@b.com", "A", "B")

	body := recorder.Body.String()
	if !strings.Contains(body, "session-key-1") {
		t.Fatalf("expected session token in response, got %d %q", recorder.Code, body)
	}
}

func TestMiddlewares(t *testing.T) {
	a := &authImplementation{endpoint: "/auth"}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)

	rec := httptest.NewRecorder()
	a.WebAuthOrRedirectMiddleware(next).ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatal("expected unauthenticated request to be redirected")
	}

	rec = httptest.NewRecorder()
	a.ApiAuthOrErrorMiddleware(next).ServeHTTP(rec, req)
	if !strings.Contains(rec.Body.String(), "auth token is required") {
		t.Fatalf("expected unauthenticated error payload, got %q", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	a.WebAppendUserIdIfExistsMiddleware(next).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected pass-through middleware, got %d", rec.Code)
	}
}

func TestHandler_ServesAuthRoutes(t *testing.T) {
	a := &authImplementation{endpoint: "/auth"}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/auth/nonexistent", nil)
	a.Handler().ServeHTTP(rec, req)
	if rec.Code == 0 || rec.Code == http.StatusInternalServerError {
		t.Fatalf("expected a response, got %d", rec.Code)
	}
}

func TestSetters_UserEmailGet(t *testing.T) {
	a := &authImplementation{}
	if a.GetFuncUserEmailGet() != nil {
		t.Fatal("expected nil by default")
	}

	a.SetFuncUserEmailGet(func(ctx context.Context, userID string, options types.UserAuthOptions) (string, error) {
		return userID + "@example.com", nil
	})
	got, err := a.GetFuncUserEmailGet()(context.Background(), "u1", types.UserAuthOptions{})
	if err != nil || got != "u1@example.com" {
		t.Fatalf("unexpected email get: %q %v", got, err)
	}
}

func TestSetters_UserProfileUpdate(t *testing.T) {
	a := &authImplementation{}

	called := ""
	a.SetFuncUserProfileUpdate(func(ctx context.Context, userID string, firstName string, lastName string, options types.UserAuthOptions) error {
		called = userID + ":" + firstName + ":" + lastName
		return nil
	})
	if err := a.GetFuncUserProfileUpdate()(context.Background(), "u1", "A", "B", types.UserAuthOptions{}); err != nil {
		t.Fatal(err)
	}
	if called != "u1:A:B" {
		t.Fatalf("expected profile update invoked, got %q", called)
	}
}

func TestSetters_UserCreationPolicy(t *testing.T) {
	a := &authImplementation{}
	if a.GetUserCreationPolicy() != "" {
		t.Fatalf("expected empty default, got %q", a.GetUserCreationPolicy())
	}
	a.SetUserCreationPolicy(types.CreateUserOnRegister)
	if a.GetUserCreationPolicy() != types.CreateUserOnRegister {
		t.Fatalf("unexpected policy %q", a.GetUserCreationPolicy())
	}
}

func TestSetters_EmailCaptured(t *testing.T) {
	a := &authImplementation{}

	captured := false
	a.SetFuncEmailCaptured(func(ctx context.Context, email string, verified bool, options types.UserAuthOptions) error {
		captured = verified
		return nil
	})
	if err := a.GetFuncEmailCaptured()(context.Background(), "a@b.com", true, types.UserAuthOptions{}); err != nil {
		t.Fatal(err)
	}
	if !captured {
		t.Fatal("expected capture hook invoked with verified=true")
	}
}

func TestSetters_PageOptions(t *testing.T) {
	a := &authImplementation{}

	a.SetPageOptions(types.PageOptions{Login: types.LoginPageOptions{Title: "Welcome"}})
	if a.GetPageOptions().Login.Title != "Welcome" {
		t.Fatal("expected SetPageOptions to store the aggregate")
	}

	a.SetLoginPageOptions(types.LoginPageOptions{Title: "Sign in please"})
	a.SetLogoutPageOptions(types.LogoutPageOptions{SubmitLabel: "Bye"})
	if a.GetPageOptions().Login.Title != "Sign in please" || a.GetPageOptions().Logout.SubmitLabel != "Bye" {
		t.Fatal("expected per-page setters to update the aggregate")
	}
}

func TestSetters_LegalLinks(t *testing.T) {
	a := &authImplementation{}

	a.SetLegalLinks(types.LegalLinks{
		Terms:   types.LegalLink{URL: "/terms", Target: types.LinkTargetNewTab},
		Privacy: types.LegalLink{URL: "/privacy"},
	})
	if a.GetLegalLinks().Terms.URL != "/terms" || a.GetLegalLinks().Privacy.URL != "/privacy" {
		t.Fatal("expected legal links round-trip")
	}
}
