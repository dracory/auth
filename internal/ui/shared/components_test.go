package shared

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracory/auth/types"
	"github.com/dracory/hb"
)

func TestPageHeader(t *testing.T) {
	html := PageHeader("Welcome", "Sign in below").ToHTML()
	if !strings.Contains(html, "Welcome") || !strings.Contains(html, "Sign in below") {
		t.Fatalf("unexpected header %q", html)
	}
}

func TestAlertGroup(t *testing.T) {
	html := AlertGroup().ToHTML()
	if !strings.Contains(html, "alert-success") || !strings.Contains(html, "alert-danger") {
		t.Fatalf("unexpected alert group %q", html)
	}
}

func TestCard(t *testing.T) {
	html := Card(AlertGroup()).ToHTML()
	if !strings.Contains(html, "auth-panel") || !strings.Contains(html, "card-body") {
		t.Fatalf("unexpected card %q", html)
	}
}

func TestFieldGroup(t *testing.T) {
	html := FieldGroup("Email", hb.NewInput()).ToHTML()
	if !strings.Contains(html, "Email") {
		t.Fatalf("unexpected field group %q", html)
	}
}

func TestNameRow(t *testing.T) {
	html := NameRow(true).ToHTML()
	if !strings.Contains(html, `name="first_name"`) || !strings.Contains(html, `name="last_name"`) {
		t.Fatalf("unexpected name row %q", html)
	}
	if !strings.Contains(html, "required") {
		t.Fatalf("expected required attributes %q", html)
	}
	html = NameRow(false).ToHTML()
	if strings.Contains(html, "required") {
		t.Fatalf("expected no required attributes %q", html)
	}
}

func TestPrimaryButton(t *testing.T) {
	html := PrimaryButton("ButtonLogin", "bi-envelope", "Log in", "doLogin()").ToHTML()
	for _, v := range []string{"ButtonLogin", "bi-envelope", "Log in", "doLogin()"} {
		if !strings.Contains(html, v) {
			t.Fatalf("expected %q in button html %q", v, html)
		}
	}
}

func TestOrDivider(t *testing.T) {
	if html := OrDivider().ToHTML(); !strings.Contains(html, "or") {
		t.Fatalf("unexpected divider %q", html)
	}
}

func TestAltLink(t *testing.T) {
	html := AltLink("/auth/login-otp", "bi-key", "Use a code").ToHTML()
	for _, v := range []string{`/auth/login-otp`, "bi-key", "Use a code"} {
		if !strings.Contains(html, v) {
			t.Fatalf("expected %q in alt link %q", v, html)
		}
	}
	// No icon → no icon element.
	html = AltLink("/x", "", "Go").ToHTML()
	if strings.Contains(html, "bi ") {
		t.Fatalf("expected no icon class %q", html)
	}
}

func TestFooterText(t *testing.T) {
	html := FooterText("No account?", "/auth/register", "Register").ToHTML()
	for _, v := range []string{"No account?", "/auth/register", "Register"} {
		if !strings.Contains(html, v) {
			t.Fatalf("expected %q in footer %q", v, html)
		}
	}
}

// knightStub satisfies the unexported authKnightChecker shape so
// RedirectAuthKnightIfActive's type assertion passes.
type knightStub struct {
	types.AuthSharedInterface
	knight bool
}

func (s knightStub) IsAuthKnight() bool { return s.knight }

func (s knightStub) LinkAuthKnightRedirect(r *http.Request) string {
	return "https://authknight.com/app/login"
}

func TestRedirectAuthKnightIfActive(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://localhost/auth/login", nil)

	if RedirectAuthKnightIfActive(rec, req, knightStub{knight: false}) {
		t.Fatal("expected no redirect when authknight disabled")
	}

	rec = httptest.NewRecorder()
	if !RedirectAuthKnightIfActive(rec, req, knightStub{knight: true}) {
		t.Fatal("expected redirect when authknight enabled")
	}
	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected temporary redirect, got %d", rec.Code)
	}
}
