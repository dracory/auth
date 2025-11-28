package page_logout

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui"
)

// Dependencies contains the dependencies required to render the logout page.
type Dependencies struct {
	Endpoint string

	Layout func(content string) string

	Logger interface {
		Error(msg string, keyvals ...interface{})
	}
}

// PageLogout renders the logout page using the provided dependencies and
// writes the result to the ResponseWriter.
func PageLogout(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	content := ui.LogoutContent()
	scripts := ui.LogoutScripts(
		links.ApiLogout(deps.Endpoint),
		links.Login(deps.Endpoint),
	)

	html := ui.BuildPage("Logout", deps.Layout, content, scripts)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(html)); err != nil {
		if deps.Logger != nil {
			deps.Logger.Error("failed to write logout page response", "error", err)
		}
	}
}
