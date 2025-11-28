package auth

import (
	"net/http"

	"github.com/dracory/auth/internal/ui"
)

func (a authImplementation) pageRegisterCodeVerify(w http.ResponseWriter, r *http.Request) {
	webpage := webpage("Verify Registration Code", a.funcLayout(a.pageRegisterCodeVerifyContent()), a.pageRegisterCodeVerifyScripts())
	logger := a.GetLogger()

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(webpage.ToHTML())); err != nil {
		logger.Error("failed to write register code verify page response", "error", err)
	}
}

func (a authImplementation) pageRegisterCodeVerifyContent() string {
	return ui.RegisterCodeVerifyContent(a.LinkRegister())
}

func (a authImplementation) pageRegisterCodeVerifyScripts() string {
	urlApiRegisterCodeVerify := a.LinkApiRegisterCodeVerify()
	urlSuccess := a.LinkRedirectOnSuccess()

	return ui.RegisterCodeVerifyScripts(urlApiRegisterCodeVerify, urlSuccess)
}
