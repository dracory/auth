package auth

import (
	"net/http"

	"github.com/dracory/auth/internal/ui/page_login"
)

func (a authImplementation) pageLogin(w http.ResponseWriter, r *http.Request) {
	deps := page_login.Dependencies{
		Passwordless:       a.passwordless,
		EnableRegistration: a.enableRegistration,
		Endpoint:           a.endpoint,
		RedirectOnSuccess:  a.LinkRedirectOnSuccess(),
		Layout:             a.funcLayout,
		Logger:             a.GetLogger(),
	}

	page_login.PageLogin(deps, w, r)
}
