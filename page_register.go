package auth

import (
	"net/http"

	page_register "github.com/dracory/auth/internal/ui/page_register"
)

func (a authImplementation) pageRegister(w http.ResponseWriter, r *http.Request) {
	deps := page_register.Dependencies{
		Passwordless:       a.passwordless,
		EnableVerification: a.enableVerification,
		Endpoint:           a.endpoint,
		Layout:             a.funcLayout,
		Logger:             a.GetLogger(),
	}

	page_register.PageRegister(deps, w, r)
}
