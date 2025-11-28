package page_login

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
)

// PageLogin renders the login page using the provided dependencies and writes
// the result to the ResponseWriter.
func PageLogin(deps Dependencies, w http.ResponseWriter, r *http.Request) {
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

	html := shared.BuildPage("Login", deps.Layout, content, scripts)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(html)); err != nil {
		if deps.Logger != nil {
			deps.Logger.Error("failed to write login page response", "error", err)
		}
	}
}
