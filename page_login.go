package auth

import (
	"net/http"

	"github.com/dracory/auth/internal/ui"
)

func (a authImplementation) pageLogin(w http.ResponseWriter, r *http.Request) {
	content := ""
	scripts := ""
	if a.passwordless {
		content = a.pageLoginPasswordlessContent()
		scripts = a.pageLoginPasswordlessScripts()
	} else {
		content = a.pageLoginContent()
		scripts = a.pageLoginScripts()
	}

	webpage := webpage("Login", a.funcLayout(content), scripts)
	logger := a.GetLogger()

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(webpage.ToHTML())); err != nil {
		logger.Error("failed to write login page response", "error", err)
	}
}

func (a authImplementation) pageLoginPasswordlessContent() string {
	return ui.LoginPasswordlessContent(a.enableRegistration, a.LinkRegister())
}

func (a authImplementation) pageLoginPasswordlessScripts() string {
	urlApiLogin := a.LinkApiLogin()
	urlSuccess := a.LinkLoginCodeVerify()

	return ui.LoginPasswordlessScripts(urlApiLogin, urlSuccess)
}

func (a authImplementation) pageLoginContent() string {
	return ui.LoginContent(a.enableRegistration, a.LinkRegister(), a.LinkPasswordRestore())
}

func (a authImplementation) pageLoginScripts() string {
	urlApiLogin := a.LinkApiLogin()
	urlSuccess := a.LinkRedirectOnSuccess()

	return ui.LoginScripts(urlApiLogin, urlSuccess)
}
