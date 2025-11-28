package page_register

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
)

// PageRegister renders the register page using the provided dependencies and
// writes the result to the ResponseWriter.
func PageRegister(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	content := ""
	scripts := ""

	if deps.Passwordless {
		content = RegisterPasswordlessContent(links.Login(deps.Endpoint))
		scripts = RegisterPasswordlessScripts(
			links.ApiRegister(deps.Endpoint),
			links.RegisterCodeVerify(deps.Endpoint),
		)
	} else {
		content = RegisterUsernameAndPasswordContent(
			links.Login(deps.Endpoint),
			links.PasswordRestore(deps.Endpoint),
		)
		urlSuccess := links.Login(deps.Endpoint)
		if deps.EnableVerification {
			urlSuccess = links.RegisterCodeVerify(deps.Endpoint)
		}
		scripts = RegisterUsernameAndPasswordScripts(
			links.ApiRegister(deps.Endpoint),
			urlSuccess,
		)
	}

	shared.PageRender(w, shared.PageOptions{
		Title:      "Register",
		Layout:     deps.Layout,
		Content:    content,
		Scripts:    scripts,
		Logger:     deps.Logger,
		LogMessage: "failed to write register page response",
	})
}
