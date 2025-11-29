package auth

import (
	"net/http"

	"github.com/dracory/auth/internal/ui/page_login"
	page_login_code_verify "github.com/dracory/auth/internal/ui/page_login_code_verify"
	page_logout "github.com/dracory/auth/internal/ui/page_logout"
	page_password_reset "github.com/dracory/auth/internal/ui/page_password_reset"
	page_password_restore "github.com/dracory/auth/internal/ui/page_password_restore"
	page_register "github.com/dracory/auth/internal/ui/page_register"
	page_register_code_verify "github.com/dracory/auth/internal/ui/page_register_code_verify"
	"github.com/dracory/req"
)

func (a authImplementation) pageLogin(w http.ResponseWriter, r *http.Request) {
	page_login.PageLogin(w, r, page_login.Dependencies{
		Passwordless:       a.passwordless,
		EnableRegistration: a.enableRegistration,
		Endpoint:           a.endpoint,
		RedirectOnSuccess:  a.LinkRedirectOnSuccess(),
		Layout:             a.funcLayout,
		Logger:             a.GetLogger(),
	})
}

func (a authImplementation) pageRegister(w http.ResponseWriter, r *http.Request) {
	page_register.PageRegister(w, r, page_register.Dependencies{
		Passwordless:       a.passwordless,
		EnableVerification: a.enableVerification,
		Endpoint:           a.endpoint,
		Layout:             a.funcLayout,
		Logger:             a.GetLogger(),
	})
}

func (a authImplementation) pageRegisterCodeVerify(w http.ResponseWriter, r *http.Request) {
	page_register_code_verify.PageRegisterCodeVerify(w, r, page_register_code_verify.Dependencies{
		Endpoint:          a.endpoint,
		RedirectOnSuccess: a.LinkRedirectOnSuccess(),
		Layout:            a.funcLayout,
		Logger:            a.GetLogger(),
	})
}

func (a authImplementation) pageLogout(w http.ResponseWriter, r *http.Request) {
	page_logout.PageLogout(w, r, &a)
}

func (a authImplementation) pagePasswordRestore(w http.ResponseWriter, r *http.Request) {
	page_password_restore.PagePasswordRestore(w, r, page_password_restore.Dependencies{
		EnableRegistration: a.enableRegistration,
		Endpoint:           a.endpoint,
		Layout:             a.funcLayout,
		Logger:             a.GetLogger(),
	})
}

func (a authImplementation) pagePasswordReset(w http.ResponseWriter, r *http.Request) {
	token := req.GetString(r, "t")
	message := ""

	if token == "" {
		message = "Link is invalid"
	} else {
		tokenValue, errToken := a.funcTemporaryKeyGet(token)
		if errToken != nil {
			message = "Link has expired"
		} else if tokenValue == "" {
			message = "Link is invalid or expired"
		}
	}

	page_password_reset.PagePasswordReset(w, r, page_password_reset.Dependencies{
		Endpoint:           a.endpoint,
		EnableRegistration: a.enableRegistration,
		Token:              token,
		ErrorMessage:       message,
		Layout:             a.funcLayout,
		Logger:             a.GetLogger(),
	})
}

func (a authImplementation) pageLoginCodeVerify(w http.ResponseWriter, r *http.Request) {
	page_login_code_verify.PageLoginCodeVerify(w, r, page_login_code_verify.Dependencies{
		Endpoint:          a.endpoint,
		RedirectOnSuccess: a.LinkRedirectOnSuccess(),
		Layout:            a.funcLayout,
		Logger:            a.GetLogger(),
	})
}
