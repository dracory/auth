package page_logout

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
)

// PageLogout renders the logout page using the provided dependencies and
// writes the result to the ResponseWriter.
func PageLogout(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	content := LogoutContent()
	scripts := LogoutScripts(
		links.ApiLogout(deps.Endpoint),
		links.Login(deps.Endpoint),
	)

	shared.PageRender(w, shared.PageOptions{
		Title:      "Logout",
		Layout:     deps.Layout,
		Content:    content,
		Scripts:    scripts,
		Logger:     deps.Logger,
		LogMessage: "failed to write logout page response",
	})
}
