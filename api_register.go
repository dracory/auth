package auth

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/req"
	"github.com/dracory/str"
)

func (a Auth) apiRegister(w http.ResponseWriter, r *http.Request) {
	// Check CSRF token
	if a.enableCSRFProtection && !a.funcCSRFTokenValidate(r) {
		api.Respond(w, r, api.Forbidden("Invalid CSRF token"))
		return
	}

	if a.passwordless {
		a.apiRegisterPasswordless(w, r)
	} else {
		a.apiRegisterUsernameAndPassword(w, r)
	}
}

func (a Auth) apiRegisterPasswordless(w http.ResponseWriter, r *http.Request) {
	// Check rate limit
	if !a.checkRateLimit(w, r, "register") {
		return
	}

	email := req.GetStringTrimmed(r, "email")
	first_name := req.GetStringTrimmed(r, "first_name")
	last_name := req.GetStringTrimmed(r, "last_name")

	if first_name == "" {
		api.Respond(w, r, api.Error("First name is required field"))
		return
	}

	if last_name == "" {
		api.Respond(w, r, api.Error("Last name is required field"))
		return
	}

	if email == "" {
		api.Respond(w, r, api.Error("Email is required field"))
		return
	}

	verificationCode, errRandomFromGamma := str.RandomFromGamma(LoginCodeLength, LoginCodeGamma)

	if errRandomFromGamma != nil {
		api.Respond(w, r, api.Error("Error generating random string"))
		return
	}

	json, errJson := json.Marshal(map[string]string{
		"email":      email,
		"first_name": first_name,
		"last_name":  last_name,
	})

	if errJson != nil {
		api.Respond(w, r, api.Error("Error serializing data"))
		return
	}

	errTempTokenSave := a.funcTemporaryKeySet(verificationCode, string(json), 3600)

	if errTempTokenSave != nil {
		api.Respond(w, r, api.Error("token store failed. "+errTempTokenSave.Error()))
		return
	}

	emailContent := a.passwordlessFuncEmailTemplateRegisterCode(email, verificationCode, UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	errEmailSent := a.passwordlessFuncEmailSend(email, "Registration Code", emailContent)

	if errEmailSent != nil {
		log.Println(errEmailSent)
		api.Respond(w, r, api.Error("Registration code failed to be send. Please try again later"))
		return
	}

	api.Respond(w, r, api.Success("Registration code was sent successfully"))
}

func (a Auth) apiRegisterUsernameAndPassword(w http.ResponseWriter, r *http.Request) {
	// Check rate limit
	if !a.checkRateLimit(w, r, "register") {
		return
	}

	email := req.GetStringTrimmed(r, "email")
	password := req.GetStringTrimmed(r, "password")
	first_name := req.GetStringTrimmed(r, "first_name")
	last_name := req.GetStringTrimmed(r, "last_name")

	response := a.RegisterWithUsernameAndPassword(email, password, first_name, last_name, UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	if response.ErrorMessage != "" {
		api.Respond(w, r, api.Error(response.ErrorMessage))
		return
	}

	api.Respond(w, r, api.Success(response.SuccessMessage))
}
