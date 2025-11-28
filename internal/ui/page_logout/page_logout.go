package page_logout

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
)

// PageLogout renders the logout page using the provided dependencies and
// writes the result to the ResponseWriter.
func PageLogout(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	content := LogoutContent()
	scripts := LogoutScripts(
		links.ApiLogout(deps.Endpoint),
		links.Login(deps.Endpoint),
	)

	html := shared.BuildPage("Logout", deps.Layout, content, scripts)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(html)); err != nil {
		if deps.Logger != nil {
			deps.Logger.Error("failed to write logout page response", "error", err)
		}
	}
}
