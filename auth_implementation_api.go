package auth

import (
	"net/http"

	"github.com/dracory/auth/internal/api/api_authenticate_via_username"
	"github.com/dracory/auth/internal/api/api_login"
	"github.com/dracory/auth/internal/api/api_login_code_verify"
	"github.com/dracory/auth/internal/api/api_logout"
	"github.com/dracory/auth/internal/api/api_password_reset"
	"github.com/dracory/auth/internal/api/api_password_restore"
	"github.com/dracory/auth/internal/api/api_register"
	"github.com/dracory/auth/internal/api/api_register_code_verify"
)

func (a authImplementation) apiLogin(w http.ResponseWriter, r *http.Request) {
	api_login.ApiLoginWithAuth(w, r, &a)
}

func (a authImplementation) apiRegister(w http.ResponseWriter, r *http.Request) {
	api_register.ApiRegisterWithAuth(w, r, &a)
}

func (a authImplementation) apiLogout(w http.ResponseWriter, r *http.Request) {
	api_logout.ApiLogoutWithAuth(w, r, &a)
}

func (a authImplementation) apiPasswordRestore(w http.ResponseWriter, r *http.Request) {
	api_password_restore.ApiPasswordRestoreWithAuth(w, r, &a)
}

func (a authImplementation) apiPasswordReset(w http.ResponseWriter, r *http.Request) {
	api_password_reset.ApiPasswordResetWithAuth(w, r, &a)
}

func (a authImplementation) apiLoginCodeVerify(w http.ResponseWriter, r *http.Request) {
	api_login_code_verify.ApiLoginCodeVerifyWithAuth(w, r, &a)
}

func (a authImplementation) apiRegisterCodeVerify(w http.ResponseWriter, r *http.Request) {
	api_register_code_verify.ApiRegisterCodeVerifyWithAuth(w, r, &a)
}

func (a authImplementation) authenticateViaUsername(w http.ResponseWriter, r *http.Request, username string, firstName string, lastName string) {
	api_authenticate_via_username.ApiAuthenticateViaUsernameWithAuth(w, r, username, firstName, lastName, &a)
}
