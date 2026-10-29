package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
