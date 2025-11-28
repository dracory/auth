package auth

import (
	"net/http"

	"github.com/dracory/api"
	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/req"
	"github.com/dracory/str"
)

// apiLoginCodeVerify used for passwordless login code verification
func (a Auth) apiLoginCodeVerify(w http.ResponseWriter, r *http.Request) {
	// Check rate limit
	if !a.checkRateLimit(w, r, "login_code_verify") {
		return
	}

	verificationCode := req.GetStringTrimmed(r, "verification_code")

	if verificationCode == "" {
		api.Respond(w, r, api.Error("Verification code is required field"))
		return
	}

	if len(verificationCode) != authutils.LoginCodeLength(a.disableRateLimit) {
		api.Respond(w, r, api.Error("Verification code is invalid length"))
		return
	}

	if !str.ContainsOnly(verificationCode, authutils.LoginCodeGamma(a.disableRateLimit)) {
		api.Respond(w, r, api.Error("Verification code contains invalid characters"))
		return
	}

	email, errCode := a.funcTemporaryKeyGet(verificationCode)

	if errCode != nil {
		api.Respond(w, r, api.Error("Verification code has expired"))
		return
	}

	a.authenticateViaUsername(w, r, email, "", "")
}

// authenticateViaEmail used for passwordless login and registration
// username is an email in passwordless auth
// firstName is used only in username and password auth
// lastName is used only in username and password auth
func (a Auth) authenticateViaUsername(w http.ResponseWriter, r *http.Request, username string, firstName string, lastName string) {
	var userID string
	var errUser error
	if a.passwordless {
		userID, errUser = a.passwordlessFuncUserFindByEmail(r.Context(), username, UserAuthOptions{
			UserIp:    req.GetIP(r),
			UserAgent: r.UserAgent(),
		})
	} else {
		userID, errUser = a.funcUserFindByUsername(r.Context(), username, firstName, lastName, UserAuthOptions{
			UserIp:    req.GetIP(r),
			UserAgent: r.UserAgent(),
		})
	}

	if errUser != nil {
		api.Respond(w, r, api.Error("Invalid credentials"))
		return
	}

	if userID == "" {
		api.Respond(w, r, api.Error("Invalid credentials"))
		return
	}

	token, errRandomFromGamma := str.RandomFromGamma(32, "BCDFGHJKLMNPQRSTVXYZ")

	if errRandomFromGamma != nil {
		api.Respond(w, r, api.Error("Error generating random string"))
		return
	}

	errSession := a.funcUserStoreAuthToken(r.Context(), token, userID, UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	if errSession != nil {
		api.Respond(w, r, api.Error("token store failed."))
		return
	}

	if a.useCookies {
		AuthCookieSet(w, r, token)
	}

	api.Respond(w, r, api.SuccessWithData("login success", map[string]interface{}{
		"token": token,
	}))
}
