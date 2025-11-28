package auth

import (
	"net/http"

	"github.com/dracory/auth/internal/ui"
)

func (a authImplementation) pageRegister(w http.ResponseWriter, r *http.Request) {
	content := ""
	scripts := ""
	if a.passwordless {
		content = a.pageRegisterPasswordlessContent()
		scripts = a.pageRegisterPasswordlessScripts()
	} else {
		content = a.pageRegisterUsernameAndPasswordContent()
		scripts = a.pageRegisterUsernameAndPasswordScripts()
	}

	webpage := webpage("Register", a.funcLayout(content), scripts)
	logger := a.GetLogger()

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(webpage.ToHTML())); err != nil {
		logger.Error("failed to write register page response", "error", err)
	}
}

func (a authImplementation) pageRegisterPasswordlessContent() string {
	return ui.RegisterPasswordlessContent(a.LinkLogin())
}

func (a authImplementation) pageRegisterPasswordlessScripts() string {
	urlApiRegister := a.LinkApiRegister()
	urlSuccess := a.LinkRegisterCodeVerify()

	return ui.RegisterPasswordlessScripts(urlApiRegister, urlSuccess)
}

func (a authImplementation) pageRegisterUsernameAndPasswordContent() string {
	return ui.RegisterUsernameAndPasswordContent(a.LinkLogin(), a.LinkPasswordRestore())
}

func (a authImplementation) pageRegisterUsernameAndPasswordScripts() string {
	urlApiRegister := a.LinkApiRegister()
	urlSuccess := a.LinkLogin()
	if a.enableVerification {
		urlSuccess = a.LinkRegisterCodeVerify()
	}

	return ui.RegisterUsernameAndPasswordScripts(urlApiRegister, urlSuccess)
}
