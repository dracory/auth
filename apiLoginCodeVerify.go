package auth

import (
	"net/http"
	"strings"

	"github.com/gouniverse/api"
	"github.com/gouniverse/utils"
)

// apiLoginCodeVerify used for passwordless login code verification
func (a Auth) apiLoginCodeVerify(w http.ResponseWriter, r *http.Request) {
	verificationCode := strings.Trim(utils.Req(r, "verification_code", ""), " ")

	if verificationCode == "" {
		api.Respond(w, r, api.Error("Verification code is required field"))
		return
	}

	if len(verificationCode) != LoginCodeLength {
		api.Respond(w, r, api.Error("Verification code is invalid length"))
		return
	}

	if !utils.StrContainsOnlySpecifiedCharacters(verificationCode, LoginCodeGamma) {
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
		userID, errUser = a.passwordlessFuncUserFindByEmail(username)
	} else {
		userID, errUser = a.funcUserFindByUsername(username, firstName, lastName)
	}

	if errUser != nil {
		api.Respond(w, r, api.Error("authentication failed. "+errUser.Error()))
		return
	}

	if userID == "" {
		api.Respond(w, r, api.Error("authentication failed. user not found"))
		return
	}

	token := utils.StrRandom(32)

	errSession := a.funcUserStoreAuthToken(token, userID, utils.IP(r), r.UserAgent())

	if errSession != nil {
		api.Respond(w, r, api.Error("token store failed. "+errSession.Error()))
		return
	}

	if a.useCookies {
		AuthCookieSet(w, r, token)
	}

	api.Respond(w, r, api.SuccessWithData("login success", map[string]interface{}{
		"token": token,
	}))
}
