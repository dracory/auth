package page_password_restore

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/auth/types"
)

// PagePasswordRestore renders the password restore page using the provided
// dependencies and writes the result to the ResponseWriter.
func PagePasswordRestore(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
	content := PasswordRestoreContent(
		a.IsRegistrationEnabled(),
		links.Login(a.GetEndpoint()),
		links.Register(a.GetEndpoint()),
	)
	scripts := PasswordRestoreScripts(
		links.ApiPasswordRestore(a.GetEndpoint()),
		links.Login(a.GetEndpoint()),
	)

	shared.PageRender(w, shared.PageOptions{
		Title:      "Restore Password",
		Layout:     a.GetLayout(),
		Content:    content,
		Scripts:    scripts,
		Logger:     a.GetLogger(),
		LogMessage: "failed to write password restore page response",
	})
}
