package auth

import (
	"encoding/json"
	"net/http"

	"github.com/dracory/api"
	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/req"
	"github.com/dracory/str"
)

func (a Auth) apiRegisterCodeVerify(w http.ResponseWriter, r *http.Request) {
	// Check rate limit
	if !a.checkRateLimit(w, r, "register_code_verify") {
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

	registerJSON, errCode := a.funcTemporaryKeyGet(verificationCode)

	if errCode != nil {
		api.Respond(w, r, api.Error("Verification code has expired"))
		return
	}

	// Unmarshal the stored JSON (string) into a map
	registerMap := map[string]interface{}{}
	errJSON := json.Unmarshal([]byte(registerJSON), &registerMap)

	if errJSON != nil {
		api.Respond(w, r, api.Error("Serialized format is malformed"))
		return
	}

	email := ""
	if val, ok := registerMap["email"]; ok {
		email = val.(string)
	}

	firstName := ""
	if val, ok := registerMap["first_name"]; ok {
		firstName = val.(string)
	}

	lastName := ""
	if val, ok := registerMap["last_name"]; ok {
		lastName = val.(string)
	}

	password := ""
	if val, ok := registerMap["password"]; ok {
		password = val.(string)
	}

	var errRegister error = nil

	if a.passwordless {
		errRegister = a.passwordlessFuncUserRegister(email, firstName, lastName, UserAuthOptions{
			UserIp:    req.GetIP(r),
			UserAgent: r.UserAgent(),
		})
	} else {
		errRegister = a.funcUserRegister(email, password, firstName, lastName, UserAuthOptions{
			UserIp:    req.GetIP(r),
			UserAgent: r.UserAgent(),
		})
	}

	if errRegister != nil {
		api.Respond(w, r, api.Error("registration failed. "+errRegister.Error()))
		return
	}

	a.authenticateViaUsername(w, r, email, firstName, lastName)
}
