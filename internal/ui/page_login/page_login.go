package page_login

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui"
)

// Dependencies contains the dependencies required to render the login page.
type Dependencies struct {
	Passwordless       bool
	EnableRegistration bool

	Endpoint          string
	RedirectOnSuccess string

	// Layout is the outer layout function supplied by the auth package.
	Layout func(content string) string

	Logger interface {
		Error(msg string, keyvals ...interface{})
	}
}

// PageLogin renders the login page using the provided dependencies and writes
// the result to the ResponseWriter.
func PageLogin(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	content := ""
	scripts := ""
	if deps.Passwordless {
		content = ui.LoginPasswordlessContent(deps.EnableRegistration, links.Register(deps.Endpoint))
		scripts = ui.LoginPasswordlessScripts(
			links.ApiLogin(deps.Endpoint),
			links.LoginCodeVerify(deps.Endpoint),
		)
	} else {
		content = ui.LoginContent(
			deps.EnableRegistration,
			links.Register(deps.Endpoint),
			links.PasswordRestore(deps.Endpoint),
		)
		scripts = ui.LoginScripts(
			links.ApiLogin(deps.Endpoint),
			deps.RedirectOnSuccess,
		)
	}

	html := ui.BuildPage("Login", deps.Layout, content, scripts)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(html)); err != nil {
		if deps.Logger != nil {
			deps.Logger.Error("failed to write login page response", "error", err)
		}
	}
}
