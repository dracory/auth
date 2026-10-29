package utils_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
)

type altMethod struct{ name, label string }

func (m altMethod) Name() string                                    { return m.name }
func (m altMethod) AlternativeLabel() string                        { return m.label }
func (m altMethod) PageHandler() http.HandlerFunc                   { return nil }
func (m altMethod) AjaxHandler() http.HandlerFunc                   { return nil }
func (m altMethod) CallbackRoute() (string, http.HandlerFunc, bool) { return "", nil, false }

func TestLoginAlternatives_SingleMethodEmpty(t *testing.T) {
	if got := utils.LoginAlternatives("/auth", "otp", []types.Method{altMethod{"otp", "code"}}); got != "" {
		t.Fatalf("expected empty for 1 method, got %q", got)
	}
	if got := utils.LoginAlternatives("/auth", "otp", nil); got != "" {
		t.Fatalf("expected empty for 0 methods, got %q", got)
	}
}

func TestLoginAlternatives_TwoMethods(t *testing.T) {
	out := utils.LoginAlternatives("/auth", "otp", []types.Method{
		altMethod{"otp", "Email me a code"},
		altMethod{"password", "Use a password"},
	})
	if !strings.Contains(out, "Use a password") {
		t.Fatalf("expected other method label, got %q", out)
	}
	if !strings.Contains(out, `href="/auth/login-password"`) {
		t.Fatalf("expected mount path, got %q", out)
	}
	if strings.Contains(out, "Email me a code") {
		t.Fatal("current method must be excluded")
	}
	if !strings.Contains(out, "OR") {
		t.Fatal("expected OR divider")
	}
}

func TestLoginAlternatives_PrimaryLinksToLogin(t *testing.T) {
	out := utils.LoginAlternatives("/auth", "password", []types.Method{
		altMethod{"otp", "code"}, altMethod{"password", "pass"},
	})
	if !strings.Contains(out, `href="/auth/login"`) {
		t.Fatalf("expected primary path /auth/login, got %q", out)
	}
}

func TestLoginAlternatives_EscapesLabel(t *testing.T) {
	out := utils.LoginAlternatives("/auth", "a", []types.Method{
		altMethod{"a", "x"}, altMethod{"b", "<script>"},
	})
	if strings.Contains(out, "<script>") {
		t.Fatal("expected escaped label")
	}
}
