package auth

import (
	"net/http"

	page_login_code_verify "github.com/dracory/auth/internal/ui/page_login_code_verify"
)

func (a authImplementation) pageLoginCodeVerify(w http.ResponseWriter, r *http.Request) {
	deps := page_login_code_verify.Dependencies{
		Endpoint:          a.endpoint,
		RedirectOnSuccess: a.LinkRedirectOnSuccess(),
		Layout:            a.funcLayout,
		Logger:            a.GetLogger(),
	}

	page_login_code_verify.PageLoginCodeVerify(deps, w, r)
}
