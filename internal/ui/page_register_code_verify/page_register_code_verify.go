package page_register_code_verify

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/auth/types"
)

// PageRegisterCodeVerify renders the registration code verification page using
// the provided dependencies and writes the result to the ResponseWriter.

func PageRegisterCodeVerify(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
	content := RegisterCodeVerifyContent(links.Register(a.GetEndpoint()))
	scripts := RegisterCodeVerifyScripts(
		links.ApiRegisterCodeVerify(a.GetEndpoint()),
		a.LinkRedirectOnSuccess(),
	)

	shared.PageRender(w, shared.PageOptions{
		Title:      "Verify Registration Code",
		Layout:     a.GetLayout(),
		Content:    content,
		Scripts:    scripts,
		Logger:     a.GetLogger(),
		LogMessage: "failed to write register code verify page response",
	})
}
