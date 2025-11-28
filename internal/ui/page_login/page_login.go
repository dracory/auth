package page_login

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
)

// PageLogin renders the login page using the provided dependencies and writes
// the result to the ResponseWriter.
func PageLogin(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	content := ""
	scripts := ""
	if deps.Passwordless {
		content = LoginPasswordlessContent(deps.EnableRegistration, links.Register(deps.Endpoint))
		scripts = LoginPasswordlessScripts(
			links.ApiLogin(deps.Endpoint),
			links.LoginCodeVerify(deps.Endpoint),
		)
	} else {
		content = LoginContent(
			deps.EnableRegistration,
			links.Register(deps.Endpoint),
			links.PasswordRestore(deps.Endpoint),
		)
		scripts = LoginScripts(
			links.ApiLogin(deps.Endpoint),
			deps.RedirectOnSuccess,
		)
	}

	shared.PageRender(w, shared.PageOptions{
		Title:      "Login",
		Layout:     deps.Layout,
		Content:    content,
		Scripts:    scripts,
		Logger:     deps.Logger,
		LogMessage: "failed to write login page response",
	})
}
