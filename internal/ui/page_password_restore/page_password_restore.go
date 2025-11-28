package page_password_restore

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
)

// PagePasswordRestore renders the password restore page using the provided
// dependencies and writes the result to the ResponseWriter.
func PagePasswordRestore(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	content := PasswordRestoreContent(
		deps.EnableRegistration,
		links.Login(deps.Endpoint),
		links.Register(deps.Endpoint),
	)
	scripts := PasswordRestoreScripts(
		links.ApiPasswordRestore(deps.Endpoint),
		links.Login(deps.Endpoint),
	)

	shared.PageRender(w, shared.PageOptions{
		Title:      "Restore Password",
		Layout:     deps.Layout,
		Content:    content,
		Scripts:    scripts,
		Logger:     deps.Logger,
		LogMessage: "failed to write password restore page response",
	})
}
