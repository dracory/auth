package page_login

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/types"
)

func TestPageLogin_UsernameAndPassword(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	testutils.SetRegistrationForTest(a, true)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	PageLogin(recorder, req, a)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()

	expected := []string{
		`<span>Sign in</span>`,
		`var urlApiLogin = "http://localhost/auth/api/login";`,
		`var urlOnSuccess = "http://localhost/dashboard";`,
	}

	for _, v := range expected {
		if !strings.Contains(body, v) {
			t.Errorf("Handler returned unexpected result.\nEXPECTED: %s\nFOUND: %s", v, body)
		}
	}
}

func TestPageLogin_Otp(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	testutils.SetRegistrationForTest(a, true)
	testutils.SetOtpForTest(a, true)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	PageLogin(recorder, req, a)

	if status := recorder.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	body := recorder.Body.String()

	expected := []string{
		`<span>Send me a login code</span>`,
		`var urlApiLogin = "http://localhost/auth/api/login";`,
		`var urlOnSuccess = "http://localhost/auth/login-code-verify";`,
	}

	for _, v := range expected {
		if !strings.Contains(body, v) {
			t.Errorf("Handler returned unexpected result.\nEXPECTED: %s\nFOUND: %s", v, body)
		}
	}
}

func TestPageLogin_AuthKnight(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	testutils.SetOtpForTest(a, false)
	testutils.SetVerificationForTest(a, false)
	testutils.SetAuthKnightForTest(a, true)

	req, err := http.NewRequest("GET", "/login", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Host = "localhost:8084"

	recorder := httptest.NewRecorder()
	PageLogin(recorder, req, a)

	if status := recorder.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusTemporaryRedirect)
	}

	location := recorder.Header().Get("Location")
	if !strings.Contains(location, "authknight.com/app/login") {
		t.Errorf("expected redirect to authknight.com, got: %s", location)
	}
	if !strings.Contains(location, "back_url=") {
		t.Errorf("expected back_url parameter in redirect, got: %s", location)
	}
	if !strings.Contains(location, "next_url=") {
		t.Errorf("expected next_url parameter in redirect, got: %s", location)
	}
	if !strings.Contains(location, "authknight/callback") {
		t.Errorf("expected next_url to contain callback path, got: %s", location)
	}
}

func TestPageLogin_PageOptionsAndLegalLinks(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	testutils.SetRegistrationForTest(a, true)
	testutils.SetPageOptionsForTest(a, types.PageOptions{
		Login: types.LoginPageOptions{
			Title:       "Welcome back",
			SubmitLabel: "Let me in",
		},
	})
	testutils.SetLegalLinksForTest(a, types.LegalLinks{
		Terms:   types.LegalLink{URL: "/terms"},
		Privacy: types.LegalLink{URL: "/privacy"},
	})

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	PageLogin(recorder, req, a)

	body := recorder.Body.String()
	for _, v := range []string{"Welcome back", "Let me in", "By continuing, you agree to", `/terms`, `/privacy`} {
		if !strings.Contains(body, v) {
			t.Errorf("expected %q in login page body", v)
		}
	}
	// Untouched fields still render the built-in copy.
	if !strings.Contains(body, "Enter your email and password to sign in.") {
		t.Error("expected default subtitle")
	}
}

func TestLoginContentX_OptionsOverride(t *testing.T) {
	html := LoginContentX(true, "/auth/register", "/auth/password-restore", LoginExtras{}, types.LoginPageOptions{
		EmailLabel:          "E-mail address",
		ForgotPasswordLabel: "Lost it?",
		FooterPrompt:        "New here?",
	}, types.LegalLinks{})
	for _, v := range []string{"E-mail address", "Lost it?", "New here?"} {
		if !strings.Contains(html, v) {
			t.Fatalf("expected %q in content %q", v, html)
		}
	}
}

func TestLoginOtpContent_Wrapper(t *testing.T) {
	html := LoginOtpContent(true, "/auth/register")
	if !strings.Contains(html, "Send me a login code") || !strings.Contains(html, "Register") {
		t.Fatalf("unexpected content %q", html)
	}
}

func TestLoginContent_Wrapper(t *testing.T) {
	html := LoginContent(true, "/auth/register", "/auth/password-restore")
	if !strings.Contains(html, "Register") {
		t.Fatalf("unexpected content %q", html)
	}
}
