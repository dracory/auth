package page_password_reset

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
)

// PagePasswordReset renders the password reset page using the provided
// dependencies and writes the result to the ResponseWriter.
func PagePasswordReset(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	urlPasswordRestore := links.PasswordRestore(deps.Endpoint)
	urlLogin := links.Login(deps.Endpoint)
	urlRegister := links.Register(deps.Endpoint)

	content := PasswordResetContent(
		deps.Token,
		deps.ErrorMessage,
		urlPasswordRestore,
		urlLogin,
		urlRegister,
		deps.EnableRegistration,
	)
	scripts := PasswordResetScripts(
		links.ApiPasswordReset(deps.Endpoint),
		links.Login(deps.Endpoint),
	)

	shared.PageRender(w, shared.PageOptions{
		Title:      "Reset Password",
		Layout:     deps.Layout,
		Content:    content,
		Scripts:    scripts,
		Logger:     deps.Logger,
		LogMessage: "failed to write password reset page response",
	})
}
