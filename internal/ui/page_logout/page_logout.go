package page_logout

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/auth/types"
)

// // PageLogout renders the logout page using the provided dependencies and
// // writes the result to the ResponseWriter.
// func PageLogout(w http.ResponseWriter, r *http.Request, deps Dependencies) {
// 	content := LogoutContent()
// 	scripts := LogoutScripts(
// 		links.ApiLogout(deps.Endpoint),
// 		links.Login(deps.Endpoint),
// 	)

// 	shared.PageRender(w, shared.PageOptions{
// 		Title:      "Logout",
// 		Layout:     deps.Layout,
// 		Content:    content,
// 		Scripts:    scripts,
// 		Logger:     deps.Logger,
// 		LogMessage: "failed to write logout page response",
// 	})
// }

// PageLogout renders the logout page using the provided dependencies and
// writes the result to the ResponseWriter.
func PageLogout(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
	content := LogoutContent()
	scripts := LogoutScripts(
		links.ApiLogout(a.GetEndpoint()),
		links.Login(a.GetEndpoint()),
	)

	shared.PageRender(w, shared.PageOptions{
		Title:      "Logout",
		Layout:     a.GetLayout(),
		Content:    content,
		Scripts:    scripts,
		Logger:     a.GetLogger(),
		LogMessage: "failed to write logout page response",
	})
}
