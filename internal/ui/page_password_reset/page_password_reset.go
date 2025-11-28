package page_password_reset

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
)

// PagePasswordReset renders the password reset page using the provided
// dependencies and writes the result to the ResponseWriter.
func PagePasswordReset(deps Dependencies, w http.ResponseWriter, r *http.Request) {
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

	html := shared.BuildPage("Reset Password", deps.Layout, content, scripts)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(html)); err != nil {
		if deps.Logger != nil {
			deps.Logger.Error("failed to write password reset page response", "error", err)
		}
	}
}
