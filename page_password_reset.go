package auth

import (
	"net/http"

	page_password_reset "github.com/dracory/auth/internal/ui/page_password_reset"
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

	deps := page_password_reset.Dependencies{
		Endpoint:           a.endpoint,
		EnableRegistration: a.enableRegistration,
		Token:              token,
		ErrorMessage:       errorMessage,
		Layout:             a.funcLayout,
		Logger:             a.GetLogger(),
	}

	page_password_reset.PagePasswordReset(deps, w, r)
}
