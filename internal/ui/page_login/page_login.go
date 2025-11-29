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
	content := ""
	scripts := ""
	if a.IsPasswordless() {
		content = LoginPasswordlessContent(a.IsRegistrationEnabled(), links.Register(a.GetEndpoint()))
		scripts = LoginPasswordlessScripts(
			links.ApiLogin(a.GetEndpoint()),
			links.LoginCodeVerify(a.GetEndpoint()),
		)
	} else {
		content = LoginContent(
			a.IsRegistrationEnabled(),
			links.Register(a.GetEndpoint()),
			links.PasswordRestore(a.GetEndpoint()),
		)
		scripts = LoginScripts(
			links.ApiLogin(a.GetEndpoint()),
			a.LinkRedirectOnSuccess(),
		)
	}

	shared.PageRender(w, shared.PageOptions{
		Title:      "Login",
		Layout:     a.GetLayout(),
		Content:    content,
		Scripts:    scripts,
		Logger:     a.GetLogger(),
		LogMessage: "failed to write login page response",
	})
}
