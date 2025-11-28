package page_register

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui"
)

// Dependencies contains the dependencies required to render the register page.
type Dependencies struct {
	Passwordless       bool
	EnableVerification bool

	Endpoint string

	Layout func(content string) string

	Logger interface {
		Error(msg string, keyvals ...interface{})
	}
}

// PageRegister renders the register page using the provided dependencies and
// writes the result to the ResponseWriter.
func PageRegister(deps Dependencies, w http.ResponseWriter, r *http.Request) {
	content := ""
	scripts := ""

	if deps.Passwordless {
		content = ui.RegisterPasswordlessContent(links.Login(deps.Endpoint))
		scripts = ui.RegisterPasswordlessScripts(
			links.ApiRegister(deps.Endpoint),
			links.RegisterCodeVerify(deps.Endpoint),
		)
	} else {
		content = ui.RegisterUsernameAndPasswordContent(
			links.Login(deps.Endpoint),
			links.PasswordRestore(deps.Endpoint),
		)
		urlSuccess := links.Login(deps.Endpoint)
		if deps.EnableVerification {
			urlSuccess = links.RegisterCodeVerify(deps.Endpoint)
		}
		scripts = ui.RegisterUsernameAndPasswordScripts(
			links.ApiRegister(deps.Endpoint),
			urlSuccess,
		)
	}

	html := ui.BuildPage("Register", deps.Layout, content, scripts)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(html)); err != nil {
		if deps.Logger != nil {
			deps.Logger.Error("failed to write register page response", "error", err)
		}
	}
}
