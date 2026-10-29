package auth

import (
	"net/http"

	"github.com/dracory/auth/internal/api/api_authenticate_via_username"
	"github.com/dracory/auth/internal/api/api_authknight_callback"
	"github.com/dracory/auth/internal/api/api_impersonate_start"
	"github.com/dracory/auth/internal/api/api_impersonate_stop"
	"github.com/dracory/auth/internal/api/api_login"
	"github.com/dracory/auth/internal/api/api_login_code_verify"
	"github.com/dracory/auth/internal/api/api_logout"
	"github.com/dracory/auth/internal/api/api_password_reset"
	"github.com/dracory/auth/internal/api/api_password_restore"
	"github.com/dracory/auth/internal/api/api_register"
	"github.com/dracory/auth/internal/api/api_register_code_verify"
	"github.com/dracory/auth/internal/core"
	"github.com/dracory/auth/types"

	"github.com/dracory/api"
	"github.com/dracory/req"
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
	// Post-auth pipeline preferred: when ANY PostAuthHook is configured,
	// the shared pipeline (allowlist → find-or-create → active → session →
	// cookie → redirect) replaces the legacy authenticate call — every
	// login method's verify path funnels through it. Checking only a
	// subset of hooks would silently skip the rest, so trigger on any.
	// Note: firstName/lastName are intentionally unused here — the pipeline
	// resolves profile via the user's own session data, unlike the legacy
	// path which forwards them.
	if hooks := a.GetPostAuthHooks(); hooks.EmailAllowed != nil ||
		hooks.UserFindByEmailOrCreate != nil || hooks.UserIsActive != nil ||
		hooks.SessionCreate != nil || hooks.SetAuthCookie != nil ||
		hooks.RedirectURL != nil || hooks.RegisterRules != nil {
		result, err := core.PostAuth(r.Context(), w, r, username, "", core.PostAuthDeps{
			Hooks:   hooks,
			Options: types.UserAuthOptions{UserIp: req.GetIP(r), UserAgent: r.UserAgent()},
			Logger:  a.GetLogger(),
		})
		if err != nil {
			api.Respond(w, r, api.Error(err.Error()))
			return
		}
		api.Respond(w, r, api.SuccessWithData(types.MsgLoginSuccess, map[string]any{
			"token":    result.Token,
			"redirect": result.Redirect,
		}))
		return
	}
	api_authenticate_via_username.ApiAuthenticateViaUsernameWithAuth(w, r, username, firstName, lastName, &a)
}

func (a authImplementation) apiImpersonateStart(w http.ResponseWriter, r *http.Request) {
	api_impersonate_start.ApiImpersonateStartWithAuth(w, r, &a)
}

func (a authImplementation) apiImpersonateStop(w http.ResponseWriter, r *http.Request) {
	api_impersonate_stop.ApiImpersonateStopWithAuth(w, r, &a)
}

func (a authImplementation) apiAuthKnightCallback(w http.ResponseWriter, r *http.Request) {
	api_authknight_callback.ApiAuthKnightCallbackWithAuth(w, r, &a)
}
