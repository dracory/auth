package page_magic_link

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/types"
)

// altAuth wraps the shared test double with sibling login alternatives so
// the alternatives block path is exercised.
type altAuth struct {
	types.AuthSharedInterface
	alts []types.LoginAlternative
}

func (a altAuth) GetLoginAlternatives() []types.LoginAlternative { return a.alts }

func render(t *testing.T, a types.AuthSharedInterface) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	recorder := httptest.NewRecorder()
	PageMagicLink(recorder, req, a)
	return recorder
}

func TestPageMagicLink_Renders(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	testutils.SetRegistrationForTest(a, true)

	recorder := render(t, a)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	expected := []string{
		`Send me a quick login link`,
		`loginFormValidate()`,
		`Register`,
	}
	for _, v := range expected {
		if !strings.Contains(body, v) {
			t.Errorf("expected %q in page body, got:\n%s", v, body)
		}
	}
}

func TestPageMagicLink_NoRegistrationNoFooter(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	testutils.SetRegistrationForTest(a, false)

	body := render(t, a).Body.String()
	if strings.Contains(body, "Don't have an account?") {
		t.Error("expected no register footer when registration is disabled")
	}
}

func TestPageMagicLink_AlternativesBlock(t *testing.T) {
	a := altAuth{
		AuthSharedInterface: testutils.NewAuthSharedForTest(),
		alts: []types.LoginAlternative{
			{Label: "Use a password", URL: "/auth/login-password", Icon: "bi-key"},
		},
	}

	body := render(t, a).Body.String()
	if !strings.Contains(body, "/auth/login-password") || !strings.Contains(body, "Use a password") {
		t.Error("expected alternatives block with sibling link")
	}
}

func TestAlternativesBlock_NoAlternatives(t *testing.T) {
	a := testutils.NewAuthSharedForTest()
	if got := alternativesBlock(a); got != "" {
		t.Fatalf("expected empty alternatives block, got %q", got)
	}
}
