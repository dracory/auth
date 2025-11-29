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
)

func (a authImplementation) pageLogin(w http.ResponseWriter, r *http.Request) {
	page_login.PageLogin(w, r, &a)
}

func (a authImplementation) pageRegister(w http.ResponseWriter, r *http.Request) {
	page_register.PageRegister(w, r, &a)
}

func (a authImplementation) pageRegisterCodeVerify(w http.ResponseWriter, r *http.Request) {
	page_register_code_verify.PageRegisterCodeVerify(w, r, &a)
}

func (a authImplementation) pageLogout(w http.ResponseWriter, r *http.Request) {
	page_logout.PageLogout(w, r, &a)
}

func (a authImplementation) pagePasswordRestore(w http.ResponseWriter, r *http.Request) {
	page_password_restore.PagePasswordRestore(w, r, &a)
}

func (a authImplementation) pagePasswordReset(w http.ResponseWriter, r *http.Request) {
	page_password_reset.PagePasswordReset(w, r, &a)
}

func (a authImplementation) pageLoginCodeVerify(w http.ResponseWriter, r *http.Request) {
	page_login_code_verify.PageLoginCodeVerify(w, r, &a)
}
