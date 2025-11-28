package auth

import (
	"net/http"

	"github.com/dracory/auth/internal/ui"
)

func (a authImplementation) pageLoginCodeVerify(w http.ResponseWriter, r *http.Request) {
	webpage := webpage("Verify Login Code", a.funcLayout(a.pageLoginCodeVerifyContent()), a.pageLoginCodeVerifyContentScripts())
	logger := a.GetLogger()

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(webpage.ToHTML())); err != nil {
		logger.Error("failed to write login code verify page response", "error", err)
	}
}

func (a authImplementation) pageLoginCodeVerifyContent() string {
	return ui.LoginCodeVerifyContent(a.LinkLogin())
}

func (a authImplementation) pageLoginCodeVerifyContentScripts() string {
	urlApiLoginCodeVerify := a.LinkApiLoginCodeVerify()
	urlSuccess := a.LinkRedirectOnSuccess()

	return ui.LoginCodeVerifyScripts(urlApiLoginCodeVerify, urlSuccess)
}
