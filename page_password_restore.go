package auth

import (
	"net/http"

	"github.com/dracory/auth/internal/ui"
)

func (a authImplementation) pagePasswordRestore(w http.ResponseWriter, r *http.Request) {
	webpage := webpage("Restore Password", a.pagePasswordRestoreContent(), a.pagePasswordRestoreScripts())
	logger := a.GetLogger()

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(webpage.ToHTML())); err != nil {
		logger.Error("failed to write password restore page response", "error", err)
	}
}

func (a authImplementation) pagePasswordRestoreContent() string {
	return ui.PasswordRestoreContent(a.enableRegistration, a.LinkLogin(), a.LinkRegister())
}

func (a authImplementation) pagePasswordRestoreScripts() string {
	urlApiPasswordRestore := a.LinkApiPasswordRestore()
	urlSuccess := a.LinkLogin()

	return ui.PasswordRestoreScripts(urlApiPasswordRestore, urlSuccess)
}
