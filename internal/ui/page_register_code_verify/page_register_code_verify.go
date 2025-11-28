package page_register_code_verify

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
)

// PageRegisterCodeVerify renders the registration code verification page using
// the provided dependencies and writes the result to the ResponseWriter.
func PageRegisterCodeVerify(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	content := RegisterCodeVerifyContent(links.Register(deps.Endpoint))
	scripts := RegisterCodeVerifyScripts(
		links.ApiRegisterCodeVerify(deps.Endpoint),
		deps.RedirectOnSuccess,
	)

	shared.PageRender(w, shared.PageOptions{
		Title:      "Verify Registration Code",
		Layout:     deps.Layout,
		Content:    content,
		Scripts:    scripts,
		Logger:     deps.Logger,
		LogMessage: "failed to write register code verify page response",
	})
}
