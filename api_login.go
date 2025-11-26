package auth

import (
	"log"
	"net/http"
	"net/mail"

	"github.com/dracory/api"
	"github.com/dracory/req"
)

func (a Auth) apiLogin(w http.ResponseWriter, r *http.Request) {
	if a.passwordless {
		a.apiLoginPasswordless(w, r)
	} else {
		a.apiLoginUsernameAndPassword(w, r)
	}
}

func (a Auth) apiLoginPasswordless(w http.ResponseWriter, r *http.Request) {
	email := req.GetStringTrimmed(r, "email")

	if email == "" {
		api.Respond(w, r, api.Error("Email is required field"))
		return
	}

	if _, err := mail.ParseAddress(email); err != nil {
		api.Respond(w, r, api.Error("This is not a valid email: "+email))
		return
	}

	verificationCode := req.GetStringTrimmed(r, "verification_code")

	errTempTokenSave := a.funcTemporaryKeySet(verificationCode, email, 3600)

	if errTempTokenSave != nil {
		api.Respond(w, r, api.Error("token store failed. "+errTempTokenSave.Error()))
		return
	}

	emailContent := a.passwordlessFuncEmailTemplateLoginCode(email, verificationCode, UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	errEmailSent := a.passwordlessFuncEmailSend(email, "Login Code", emailContent)

	if errEmailSent != nil {
		log.Println(errEmailSent)
		api.Respond(w, r, api.Error("Login code failed to be send. Please try again later"))
		return
	}

	api.Respond(w, r, api.Success("Login code was sent successfully"))
}

func (a Auth) apiLoginUsernameAndPassword(w http.ResponseWriter, r *http.Request) {
	email := req.GetStringTrimmed(r, "email")
	password := req.GetStringTrimmed(r, "password")

	response := a.LoginWithUsernameAndPassword(email, password, UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	if response.ErrorMessage != "" {
		api.Respond(w, r, api.Error(response.ErrorMessage))
		return
	}

	if a.useCookies {
		AuthCookieSet(w, r, response.Token)
	}

	api.Respond(w, r, api.SuccessWithData(response.SuccessMessage, map[string]any{
		"token": response.Token,
	}))
}
