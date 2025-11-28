package page_password_restore

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui"
)

// Dependencies contains the dependencies required to render the password restore page.
type Dependencies struct {
	EnableRegistration bool

	Endpoint string

	Layout func(content string) string

	Logger interface {
		Error(msg string, keyvals ...interface{})
	}
}

// PagePasswordRestore renders the password restore page using the provided
// dependencies and writes the result to the ResponseWriter.
func PagePasswordRestore(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	content := ui.PasswordRestoreContent(
		deps.EnableRegistration,
		links.Login(deps.Endpoint),
		links.Register(deps.Endpoint),
	)
	scripts := ui.PasswordRestoreScripts(
		links.ApiPasswordRestore(deps.Endpoint),
		links.Login(deps.Endpoint),
	)

	html := ui.BuildPage("Restore Password", deps.Layout, content, scripts)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(html)); err != nil {
		if deps.Logger != nil {
			deps.Logger.Error("failed to write password restore page response", "error", err)
		}
	}
}
