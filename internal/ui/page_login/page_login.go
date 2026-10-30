package page_login

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/auth/types"
)

// PageLogin renders the login page using the provided dependencies and writes
// the result to the ResponseWriter.

func PageLogin(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
	PageLoginWithAPI(w, r, a, links.ApiLogin(a.GetEndpoint()))
}

// PageLoginWithAPI renders the login page with an explicit ajax endpoint.
// Used by the multi-method mux where each method's ajax lives at its own
// mount path (POST /auth/login[-<name>]) rather than /api/login.
func PageLoginWithAPI(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface, apiLoginURL string) {
	if shared.RedirectAuthKnightIfActive(w, r, a) {
		return
	}

	x := LoginExtras{
		BrandName:    brandName(a),
		BrandTagline: brandTagline(a),
		BrandLogoURL: brandLogoURL(a),
	}
	if getter, ok := a.(interface {
		GetLoginAlternatives() []types.LoginAlternative
	}); ok {
		x.Alternatives = getter.GetLoginAlternatives()
	}
	if g, ok := a.(interface{ GetRememberMeEnabled() bool }); ok {
		x.RememberMe = g.GetRememberMeEnabled()
	}

	content := ""
	scripts := ""
	if a.IsPasswordless() {
		content = LoginPasswordlessContentX(a.IsRegistrationEnabled(), links.Register(a.GetEndpoint()), x)
		scripts = LoginPasswordlessScripts(
			apiLoginURL,
			links.LoginCodeVerify(a.GetEndpoint()),
		)
	} else {
		content = LoginContentX(
			a.IsRegistrationEnabled(),
			links.Register(a.GetEndpoint()),
			links.PasswordRestore(a.GetEndpoint()),
			x,
		)
		scripts = LoginScripts(
			apiLoginURL,
			a.LinkRedirectOnSuccess(),
		)
	}

	shared.PageRender(w, r, shared.PageOptions{
		Title:      "Login",
		Auth:       a,
		Layout:     a.GetLayout(),
		Content:    content,
		Scripts:    scripts,
		Logger:     a.GetLogger(),
		LogMessage: "failed to write login page response",
	})
}

// brandName reads the optional brand-name getter from the auth
// implementation (nil-safe).
func brandName(a types.AuthSharedInterface) string {
	if g, ok := a.(interface{ GetBrandName() string }); ok {
		return g.GetBrandName()
	}
	return ""
}

func brandTagline(a types.AuthSharedInterface) string {
	if g, ok := a.(interface{ GetBrandTagline() string }); ok {
		return g.GetBrandTagline()
	}
	return ""
}

func brandLogoURL(a types.AuthSharedInterface) string {
	if g, ok := a.(interface{ GetBrandLogoURL() string }); ok {
		return g.GetBrandLogoURL()
	}
	return ""
}
