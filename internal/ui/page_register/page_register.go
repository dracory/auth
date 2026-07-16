package page_register

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/auth/types"
)

// PageRegister renders the register page using the provided dependencies and
// writes the result to the ResponseWriter.

func PageRegister(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
	// With ak_key: render AuthKnight register form (check first before AuthKnight redirect guard)
	akKey := r.URL.Query().Get("ak_key")
	if akKey != "" {
		email := getAuthKnightEmail(a, akKey)
		if email != "" {
			content := RegisterAuthKnightContent(links.Login(a.GetEndpoint()), email, akKey)
			scripts := RegisterAuthKnightScripts(
				links.ApiRegister(a.GetEndpoint()),
				a.LinkRedirectOnSuccess(),
			)
			shared.PageRender(w, shared.PageOptions{
				Title:      "Register",
				Layout:     a.GetLayout(),
				Content:    content,
				Scripts:    scripts,
				Logger:     a.GetLogger(),
				LogMessage: "failed to write register page response",
			})
			return
		}
	}

	if shared.RedirectAuthKnightIfActive(w, r, a) {
		return
	}

	content := ""
	scripts := ""

	if a.IsPasswordless() {
		content = RegisterPasswordlessContent(links.Login(a.GetEndpoint()))
		scripts = RegisterPasswordlessScripts(
			links.ApiRegister(a.GetEndpoint()),
			links.RegisterCodeVerify(a.GetEndpoint()),
		)
	} else {
		content = RegisterUsernameAndPasswordContent(
			links.Login(a.GetEndpoint()),
			links.PasswordRestore(a.GetEndpoint()),
		)
		urlSuccess := links.Login(a.GetEndpoint())
		if a.IsVerificationEnabled() {
			urlSuccess = links.RegisterCodeVerify(a.GetEndpoint())
		}
		scripts = RegisterUsernameAndPasswordScripts(
			links.ApiRegister(a.GetEndpoint()),
			urlSuccess,
		)
	}

	shared.PageRender(w, shared.PageOptions{
		Title:      "Register",
		Layout:     a.GetLayout(),
		Content:    content,
		Scripts:    scripts,
		Logger:     a.GetLogger(),
		LogMessage: "failed to write register page response",
	})
}

// getAuthKnightEmail retrieves the verified email from the temporary key store
// using the provided ak_key.
func getAuthKnightEmail(a types.AuthSharedInterface, akKey string) string {
	if akKey == "" {
		return ""
	}

	fn := a.GetFuncTemporaryKeyGet()
	if fn == nil {
		return ""
	}

	email, err := fn(akKey)
	if err != nil {
		return ""
	}
	return email
}
