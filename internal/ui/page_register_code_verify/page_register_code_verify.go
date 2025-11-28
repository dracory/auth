package page_register_code_verify

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
)

// PageRegisterCodeVerify renders the registration code verification page using
// the provided dependencies and writes the result to the ResponseWriter.
func PageRegisterCodeVerify(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	content := RegisterCodeVerifyContent(links.Register(deps.Endpoint))
	scripts := RegisterCodeVerifyScripts(
		links.ApiRegisterCodeVerify(deps.Endpoint),
		deps.RedirectOnSuccess,
	)

	html := shared.BuildPage("Verify Registration Code", deps.Layout, content, scripts)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(html)); err != nil {
		if deps.Logger != nil {
			deps.Logger.Error("failed to write register code verify page response", "error", err)
		}
	}
}
