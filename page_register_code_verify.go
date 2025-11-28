package auth

import (
	"net/http"

	page_register_code_verify "github.com/dracory/auth/internal/ui/page_register_code_verify"
)

func (a authImplementation) pageRegisterCodeVerify(w http.ResponseWriter, r *http.Request) {
	deps := page_register_code_verify.Dependencies{
		Endpoint:          a.endpoint,
		RedirectOnSuccess: a.LinkRedirectOnSuccess(),
		Layout:            a.funcLayout,
		Logger:            a.GetLogger(),
	}

	page_register_code_verify.PageRegisterCodeVerify(deps, w, r)
}
