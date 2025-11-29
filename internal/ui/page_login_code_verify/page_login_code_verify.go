package page_login_code_verify

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/auth/types"
)

// PageLoginCodeVerify renders the login code verification page using the
// provided dependencies and writes the result to the ResponseWriter.

func PageLoginCodeVerify(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
	content := LoginCodeVerifyContent(links.Login(a.GetEndpoint()))
	scripts := LoginCodeVerifyScripts(
		links.ApiLoginCodeVerify(a.GetEndpoint()),
		a.LinkRedirectOnSuccess(),
	)

	shared.PageRender(w, shared.PageOptions{
		Title:      "Login Code Verification",
		Layout:     a.GetLayout(),
		Content:    content,
		Scripts:    scripts,
		Logger:     a.GetLogger(),
		LogMessage: "failed to write login code verify page response",
	})
}
