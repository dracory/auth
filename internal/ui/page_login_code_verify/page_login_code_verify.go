package page_login_code_verify

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
)

// PageLoginCodeVerify renders the login code verification page using the
// provided dependencies and writes the result to the ResponseWriter.
func PageLoginCodeVerify(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	content := LoginCodeVerifyContent(links.Login(deps.Endpoint))
	scripts := LoginCodeVerifyScripts(
		links.ApiLoginCodeVerify(deps.Endpoint),
		deps.RedirectOnSuccess,
	)

	html := shared.BuildPage("Verify Login Code", deps.Layout, content, scripts)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(html)); err != nil {
		if deps.Logger != nil {
			deps.Logger.Error("failed to write login code verify page response", "error", err)
		}
	}
}
