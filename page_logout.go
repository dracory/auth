package auth

import (
	"net/http"

	"github.com/dracory/auth/internal/ui"
)

func (a authImplementation) pageLogout(w http.ResponseWriter, r *http.Request) {
	webpage := webpage("Logout", a.funcLayout(a.pageLogoutContent()), a.pageLogoutScripts())
	logger := a.GetLogger()

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(webpage.ToHTML())); err != nil {
		logger.Error("failed to write logout page response", "error", err)
	}
}

func (a authImplementation) pageLogoutContent() string {
	return ui.LogoutContent()
}

func (a authImplementation) pageLogoutScripts() string {
	urlApiLogout := a.LinkApiLogout()
	urlSuccess := a.LinkLogin()
	logger := a.GetLogger()
	logger.Debug("logout page initialized",
		"api_logout_url", urlApiLogout,
	)

	return ui.LogoutScripts(urlApiLogout, urlSuccess)
}
