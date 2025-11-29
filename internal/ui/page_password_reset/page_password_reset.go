package page_password_reset

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/auth/types"
	"github.com/dracory/req"
)

// PagePasswordReset renders the password reset page using the provided
// auth instance and computes the user-facing message internally.
func PagePasswordReset(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
	urlPasswordRestore := links.PasswordRestore(a.GetEndpoint())
	urlLogin := links.Login(a.GetEndpoint())
	urlRegister := links.Register(a.GetEndpoint())

	token := req.GetString(r, "t")

	message := ""
	if token == "" {
		message = "Link is invalid"
	} else {
		if fn := a.GetFuncTemporaryKeyGet(); fn != nil {
			if value, err := fn(token); err != nil {
				message = "Link has expired"
			} else if value == "" {
				message = "Link is invalid or expired"
			}
		}
	}

	content := PasswordResetContent(
		token,
		message,
		urlPasswordRestore,
		urlLogin,
		urlRegister,
		a.IsRegistrationEnabled(),
	)
	scripts := PasswordResetScripts(
		links.ApiPasswordReset(a.GetEndpoint()),
		links.Login(a.GetEndpoint()),
	)

	shared.PageRender(w, shared.PageOptions{
		Title:      "Reset Password",
		Layout:     a.GetLayout(),
		Content:    content,
		Scripts:    scripts,
		Logger:     a.GetLogger(),
		LogMessage: "failed to write password reset page response",
	})
}
