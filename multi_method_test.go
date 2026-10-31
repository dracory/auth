package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracory/auth/internal/api/api_magic_link_callback"
	"github.com/dracory/auth/internal/api/api_magic_link_send"
	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/types"
)

// stubMethod is a minimal Method implementation for route-table tests.
type stubMethod struct {
	name     string
	callback bool
	cbPath   string
	extra    map[string]http.HandlerFunc
}

func (m stubMethod) Name() string             { return m.name }
func (m stubMethod) AlternativeLabel() string { return m.name }
func (m stubMethod) PageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("page:" + m.name)) }
}
func (m stubMethod) AjaxHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ajax:" + m.name)) }
}
func (m stubMethod) CallbackRoute() (string, http.HandlerFunc, bool) {
	if !m.callback {
		return "", nil, false
	}
	return m.cbPath, func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("cb:" + m.name)) }, true
}
func (m stubMethod) ExtraRoutes() map[string]http.HandlerFunc { return m.extra }

func serve(t *testing.T, mux *http.ServeMux, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestMultiMethod_SingleMethodPrimaryOnly(t *testing.T) {
	mma, err := NewMultiMethodAuth("/auth", []types.Method{stubMethod{name: "otp"}})
	if err != nil {
		t.Fatal(err)
	}
	if rec := serve(t, mma.Router(), "GET", "/auth/login"); rec.Code != 200 || !strings.Contains(rec.Body.String(), "page:otp") {
		t.Fatalf("expected primary page at /auth/login, got %d", rec.Code)
	}
	if rec := serve(t, mma.Router(), "POST", "/auth/login"); !strings.Contains(rec.Body.String(), "ajax:otp") {
		t.Fatalf("expected primary ajax, got %d", rec.Code)
	}
	// No alternative routes for a single method.
	if rec := serve(t, mma.Router(), "GET", "/auth/login-otp"); rec.Code != 404 {
		t.Fatalf("expected no alternative route, got %d", rec.Code)
	}
}

func TestMultiMethod_SecondaryMountsAtNamePath(t *testing.T) {
	mma, err := NewMultiMethodAuth("/auth", []types.Method{
		stubMethod{name: "otp"},
		stubMethod{name: "password"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if rec := serve(t, mma.Router(), "GET", "/auth/login-password"); !strings.Contains(rec.Body.String(), "page:password") {
		t.Fatalf("expected secondary page, got %d %q", rec.Code, rec.Body.String())
	}
	if rec := serve(t, mma.Router(), "GET", "/auth/login"); !strings.Contains(rec.Body.String(), "page:otp") {
		t.Fatalf("expected primary still at /auth/login, got %d", rec.Code)
	}
}

func TestMultiMethod_CallbackMountedOnce(t *testing.T) {
	mma, err := NewMultiMethodAuth("/auth", []types.Method{
		stubMethod{name: "magiclink", callback: true, cbPath: "/magiclink-callback"},
		stubMethod{name: "otp"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if rec := serve(t, mma.Router(), "GET", "/auth/magiclink-callback"); !strings.Contains(rec.Body.String(), "cb:magiclink") {
		t.Fatalf("expected callback mounted, got %d", rec.Code)
	}
}

func TestMultiMethod_ExtraRoutesOnlyWhenPresent(t *testing.T) {
	withExtra := stubMethod{name: "password", extra: map[string]http.HandlerFunc{
		"GET /password-restore": func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("restore")) },
	}}

	mma, _ := NewMultiMethodAuth("/auth", []types.Method{stubMethod{name: "otp"}, withExtra})
	if rec := serve(t, mma.Router(), "GET", "/auth/password-restore"); rec.Code != 200 {
		t.Fatalf("expected recovery route when password present, got %d", rec.Code)
	}

	mma2, _ := NewMultiMethodAuth("/auth", []types.Method{stubMethod{name: "otp"}})
	if rec := serve(t, mma2.Router(), "GET", "/auth/password-restore"); rec.Code != 404 {
		t.Fatalf("expected no recovery route without password, got %d", rec.Code)
	}
}

func TestMultiMethod_DuplicateNameError(t *testing.T) {
	_, err := NewMultiMethodAuth("/auth", []types.Method{
		stubMethod{name: "otp"},
		stubMethod{name: "otp"},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate name error, got %v", err)
	}
}

func TestMultiMethod_EmptyMethodsError(t *testing.T) {
	if _, err := NewMultiMethodAuth("/auth", nil); err == nil {
		t.Fatal("expected error for empty methods")
	}
}

func TestMultiMethod_BuiltInOTPAdapts(t *testing.T) {
	shared := testutils.NewAuthSharedForTest()
	m := NewOTPMethod(shared.(types.AuthSharedInterface))
	if m.Name() != "otp" || m.PageHandler() == nil || m.AjaxHandler() == nil {
		t.Fatal("expected otp method to expose name, page, ajax")
	}
}

func TestMultiMethod_InvalidNameSlug(t *testing.T) {
	for _, name := range []string{"my method", "x/y", "UPPER", "-lead"} {
		_, err := NewMultiMethodAuth("/auth", []types.Method{stubMethod{name: name}})
		if err == nil || !strings.Contains(err.Error(), "slug") {
			t.Fatalf("expected slug error for %q, got %v", name, err)
		}
	}
}

func TestMultiMethod_DuplicateCallbackPathError(t *testing.T) {
	_, err := NewMultiMethodAuth("/auth", []types.Method{
		stubMethod{name: "a", callback: true, cbPath: "/cb"},
		stubMethod{name: "b", callback: true, cbPath: "/cb"},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate route") {
		t.Fatalf("expected duplicate route error, got %v", err)
	}
}

func TestMultiMethod_PostCallbackPath(t *testing.T) {
	mma, err := NewMultiMethodAuth("/auth", []types.Method{
		stubMethod{name: "oauth", callback: true, cbPath: "POST /oauth-callback"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if rec := serve(t, mma.Router(), "POST", "/auth/oauth-callback"); !strings.Contains(rec.Body.String(), "cb:oauth") {
		t.Fatalf("expected POST callback, got %d %q", rec.Code, rec.Body.String())
	}
	if rec := serve(t, mma.Router(), "GET", "/auth/oauth-callback"); rec.Code != 405 {
		t.Fatalf("expected GET rejected, got %d", rec.Code)
	}
}

func TestMultiMethod_HandlerAndMethods(t *testing.T) {
	methods := []types.Method{stubMethod{name: "otp"}, stubMethod{name: "password"}}
	mma, err := NewMultiMethodAuth("/auth", methods)
	if err != nil {
		t.Fatal(err)
	}

	if mma.Handler() == nil {
		t.Fatal("expected non-nil handler")
	}
	if len(mma.Methods()) != 2 || mma.Methods()[0].Name() != "otp" {
		t.Fatalf("unexpected methods %+v", mma.Methods())
	}
}

func TestMultiMethod_LoginAlternativesHTML(t *testing.T) {
	mma, err := NewMultiMethodAuth("/auth", []types.Method{
		stubMethod{name: "otp"},
		stubMethod{name: "password"},
	})
	if err != nil {
		t.Fatal(err)
	}

	html := mma.LoginAlternativesHTML("otp")
	if !strings.Contains(html, "/auth/login-password") {
		t.Fatalf("expected sibling link in alternatives html, got %q", html)
	}
}

func TestNewOTPMethodWithPage_CustomPage(t *testing.T) {
	shared := testutils.NewAuthSharedForTest()
	custom := func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("custom")) }
	m := NewOTPMethodWithPage(shared.(types.AuthSharedInterface), custom)

	mma, err := NewMultiMethodAuth("/auth", []types.Method{m})
	if err != nil {
		t.Fatal(err)
	}
	if rec := serve(t, mma.Router(), "GET", "/auth/login"); !strings.Contains(rec.Body.String(), "custom") {
		t.Fatalf("expected custom page, got %d %q", rec.Code, rec.Body.String())
	}
}

func TestOTPMethod_ExtraRoutesMountCodeVerify(t *testing.T) {
	shared := testutils.NewAuthSharedForTest()
	m := NewOTPMethod(shared)

	mma, err := NewMultiMethodAuth("/auth", []types.Method{m})
	if err != nil {
		t.Fatal(err)
	}
	if rec := serve(t, mma.Router(), "GET", "/auth/login-code-verify"); rec.Code != 200 {
		t.Fatalf("expected code-verify page mounted, got %d", rec.Code)
	}
	if path, h, ok := m.CallbackRoute(); ok || path != "" || h != nil {
		t.Fatal("expected no callback route for otp")
	}
}

func TestNewMagicLinkMethod_PanicsOnNilPage(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for nil page")
		}
	}()
	NewMagicLinkMethod(nil, api_magic_link_send.Dependencies{}, api_magic_link_callback.Dependencies{})
}

func TestMagicLinkMethod_Adaptation(t *testing.T) {
	page := func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ml-page")) }
	m := NewMagicLinkMethod(page,
		api_magic_link_send.Dependencies{},
		api_magic_link_callback.Dependencies{},
	)

	if m.Name() != "magiclink" {
		t.Fatalf("unexpected name %q", m.Name())
	}
	if m.AlternativeLabel() == "" {
		t.Fatal("expected alternative label")
	}
	if m.PageHandler() == nil || m.AjaxHandler() == nil {
		t.Fatal("expected page and ajax handlers")
	}
	path, h, ok := m.CallbackRoute()
	if !ok || path != "/magiclink-callback" || h == nil {
		t.Fatalf("unexpected callback route %q %v", path, ok)
	}
}

func TestNewMagicLinkMethodFromAuth_BuildsDeps(t *testing.T) {
	shared := testutils.NewAuthSharedForTest()
	m := NewMagicLinkMethodFromAuth(shared, nil, func(token string) string {
		return "/auth/magiclink-callback?token=" + token
	})
	if m == nil || m.Name() != "magiclink" {
		t.Fatal("expected magiclink method")
	}
	// nil page falls back to the package's generic magic-link page.
	if m.PageHandler() == nil {
		t.Fatal("expected default package page handler")
	}
}

func TestPasswordMethod_AdaptationAndRoutes(t *testing.T) {
	shared := testutils.NewAuthSharedForTest()
	testutils.SetRegistrationForTest(shared, true)

	m := NewPasswordMethod(shared, nil)
	if m.Name() != "password" || m.AlternativeLabel() == "" {
		t.Fatal("expected password method adaptation")
	}
	if path, _, ok := m.CallbackRoute(); ok || path != "" {
		t.Fatal("expected no callback route for password")
	}
	if len(m.(types.RouteProvider).ExtraRoutes()) == 0 {
		t.Fatal("expected recovery extra routes")
	}

	mma, err := NewMultiMethodAuth("/auth", []types.Method{m})
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/auth/password-restore", "/auth/password-reset", "/auth/register", "/auth/logout"} {
		if rec := serve(t, mma.Router(), "GET", path); rec.Code != 200 {
			t.Fatalf("expected %s mounted, got %d", path, rec.Code)
		}
	}
}

func TestMultiMethod_NoRegistrationNoRegisterRoutes(t *testing.T) {
	shared := testutils.NewAuthSharedForTest()
	testutils.SetRegistrationForTest(shared, false)

	mma, err := NewMultiMethodAuth("/auth", []types.Method{NewPasswordMethod(shared, nil)})
	if err != nil {
		t.Fatal(err)
	}
	if rec := serve(t, mma.Router(), "GET", "/auth/register"); rec.Code != 404 {
		t.Fatalf("expected no register route, got %d", rec.Code)
	}
}

func TestMultiMethod_WiresSiblingAlternatives(t *testing.T) {
	shared := testutils.NewAuthSharedForTest()
	mma, err := NewMultiMethodAuth("/auth", []types.Method{
		NewOTPMethod(shared),
		NewPasswordMethod(shared, nil),
	})
	if err != nil {
		t.Fatal(err)
	}

	impl := mma.Methods()[0].(implSwapper).authImpl()
	getter, ok := impl.(interface {
		GetLoginAlternatives() []types.LoginAlternative
	})
	if !ok {
		t.Fatal("expected impl to be decorated with alternatives")
	}
	alts := getter.GetLoginAlternatives()
	if len(alts) != 1 || alts[0].URL != "/auth/login-password" {
		t.Fatalf("unexpected alternatives %+v", alts)
	}

	// The decorator must unwrap to the original impl for capability checks.
	unwrapper, ok := impl.(interface {
		Unwrap() types.AuthSharedInterface
	})
	if !ok || unwrapper.Unwrap() == nil {
		t.Fatal("expected decorator to expose Unwrap")
	}
}

func TestAuthKnightMethod_Adaptation(t *testing.T) {
	a, err := NewAuthKnightAuth(newValidAuthKnightConfig())
	if err != nil {
		t.Fatal(err)
	}
	m := NewAuthKnightMethod(a)

	if m.Name() != "authknight" || m.AlternativeLabel() == "" {
		t.Fatal("expected authknight method adaptation")
	}
	if m.AjaxHandler() != nil {
		t.Fatal("expected no ajax handler for external flow")
	}
	if m.PageHandler() == nil {
		t.Fatal("expected page handler")
	}
	path, h, ok := m.CallbackRoute()
	if !ok || path != "/authknight-callback" || h == nil {
		t.Fatalf("unexpected callback route %q %v", path, ok)
	}

	mma, err := NewMultiMethodAuth("/auth", []types.Method{m})
	if err != nil {
		t.Fatal(err)
	}
	if rec := serve(t, mma.Router(), "GET", "/auth/authknight-callback"); rec.Code == 404 || rec.Code == 405 {
		t.Fatalf("expected callback route mounted, got %d", rec.Code)
	}
}

func TestMagicLinkMethod_AuthImplSwap(t *testing.T) {
	shared := testutils.NewAuthSharedForTest()
	page := func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ml")) }
	m := NewMagicLinkMethod(page,
		api_magic_link_send.Dependencies{},
		api_magic_link_callback.Dependencies{},
	)

	sw := m.(implSwapper)
	if sw.authImpl() != nil {
		t.Fatal("expected nil impl before swap")
	}
	sw.setAuthImpl(shared)
	if sw.authImpl() == nil {
		t.Fatal("expected impl after swap")
	}
}
