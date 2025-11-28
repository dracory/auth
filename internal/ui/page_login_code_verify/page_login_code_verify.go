package page_login_code_verify

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
)

// PageLoginCodeVerify renders the login code verification page using the
// provided dependencies and writes the result to the ResponseWriter.
func PageLoginCodeVerify(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	content := LoginCodeVerifyContent(links.Login(deps.Endpoint))
	scripts := LoginCodeVerifyScripts(
		links.ApiLoginCodeVerify(deps.Endpoint),
		deps.RedirectOnSuccess,
	)

	shared.PageRender(w, shared.PageOptions{
		Title:      "Verify Login Code",
		Layout:     deps.Layout,
		Content:    content,
		Scripts:    scripts,
		Logger:     deps.Logger,
		LogMessage: "failed to write login code verify page response",
	})
}
