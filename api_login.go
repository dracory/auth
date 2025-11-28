package auth

import (
	"log"
	"net/http"
	"net/mail"

	"github.com/dracory/api"
	"github.com/dracory/req"
)

func (a Auth) apiLogin(w http.ResponseWriter, r *http.Request) {
	// Check CSRF token
	if a.enableCSRFProtection && !a.funcCSRFTokenValidate(r) {
		api.Respond(w, r, api.Forbidden("Invalid CSRF token"))
		return
	}

	if a.passwordless {
		a.apiLoginPasswordless(w, r)
	} else {
		a.apiLoginUsernameAndPassword(w, r)
	}
}

func (a Auth) apiLoginPasswordless(w http.ResponseWriter, r *http.Request) {
	// Check rate limit
	if !a.checkRateLimit(w, r, "login") {
		return
	}

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
		log.Println("token store failed:", errTempTokenSave)
		api.Respond(w, r, api.Error("token store failed."))
		return
	}

	emailContent := a.passwordlessFuncEmailTemplateLoginCode(r.Context(), email, verificationCode, UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	errEmailSent := a.passwordlessFuncEmailSend(r.Context(), email, "Login Code", emailContent)

	if errEmailSent != nil {
		log.Println(errEmailSent)
		api.Respond(w, r, api.Error("Login code failed to be send. Please try again later"))
		return
	}

	api.Respond(w, r, api.Success("Login code was sent successfully"))
}

func (a Auth) apiLoginUsernameAndPassword(w http.ResponseWriter, r *http.Request) {
	// Check rate limit
	if !a.checkRateLimit(w, r, "login") {
		return
	}

	email := req.GetStringTrimmed(r, "email")
	password := req.GetStringTrimmed(r, "password")

	response := a.LoginWithUsernameAndPassword(r.Context(), email, password, UserAuthOptions{
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
