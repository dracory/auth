package auth

import (
	"net/http"

	"github.com/dracory/auth/internal/ui"
	"github.com/dracory/req"
)

func (a authImplementation) pagePasswordReset(w http.ResponseWriter, r *http.Request) {
	token := req.GetString(r, "t")
	errorMessage := ""

	if token == "" {
		errorMessage = "Link is invalid"
	} else {
		tokenValue, errToken := a.funcTemporaryKeyGet(token)
		if errToken != nil {
			errorMessage = "Link has expired"
		} else if tokenValue == "" {
			errorMessage = "Link is invalid or expired"
		}
	}

	h := a.pagePasswordResetContent(token, errorMessage)
	webpage := webpage("Reset Password", h, a.pagePasswordResetScripts())
	logger := a.GetLogger()

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(webpage.ToHTML())); err != nil {
		logger.Error("failed to write password reset page response", "error", err)
	}
}

func (a authImplementation) pagePasswordResetContent(token string, errorMessage string) string {
	urlPasswordRestore := a.LinkPasswordRestore()
	urlLogin := a.LinkLogin()
	urlRegister := a.LinkRegister()

	return ui.PasswordResetContent(token, errorMessage, urlPasswordRestore, urlLogin, urlRegister, a.enableRegistration)
}

func (a authImplementation) pagePasswordResetScripts() string {
	urlApiPasswordReset := a.LinkApiPasswordReset()
	urlSuccess := a.LinkLogin()

	return ui.PasswordResetScripts(urlApiPasswordReset, urlSuccess)
}
