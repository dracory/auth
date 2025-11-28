package page_register_code_verify

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui"
)

// Dependencies contains the dependencies required to render the register code verify
// page.
type Dependencies struct {
	Endpoint          string
	RedirectOnSuccess string

	Layout func(content string) string

	Logger interface {
		Error(msg string, keyvals ...interface{})
	}
}

// PageRegisterCodeVerify renders the registration code verification page using
// the provided dependencies and writes the result to the ResponseWriter.
func PageRegisterCodeVerify(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	content := ui.RegisterCodeVerifyContent(links.Register(deps.Endpoint))
	scripts := ui.RegisterCodeVerifyScripts(
		links.ApiRegisterCodeVerify(deps.Endpoint),
		deps.RedirectOnSuccess,
	)

	html := ui.BuildPage("Verify Registration Code", deps.Layout, content, scripts)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(html)); err != nil {
		if deps.Logger != nil {
			deps.Logger.Error("failed to write register code verify page response", "error", err)
		}
	}
}
