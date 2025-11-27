package auth

import (
	"log"
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/req"
	"github.com/dracory/str"
)

func (a Auth) apiPasswordRestore(w http.ResponseWriter, r *http.Request) {
	// Check rate limit
	if !a.checkRateLimit(w, r, "password_restore") {
		return
	}

	email := req.GetStringTrimmed(r, "email")
	firstName := req.GetStringTrimmed(r, "first_name")
	lastName := req.GetStringTrimmed(r, "last_name")

	if email == "" {
		api.Respond(w, r, api.Error("Email is required field"))
		return
	}

	if firstName == "" {
		api.Respond(w, r, api.Error("First name is required field"))
		return
	}

	if lastName == "" {
		api.Respond(w, r, api.Error("Last name is required field"))
		return
	}

	userID, err := a.funcUserFindByUsername(email, firstName, lastName, UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	if err != nil {
		log.Println(err.Error())
		api.Respond(w, r, api.Error("Internal server error"))
		return
	}

	if userID == "" {
		api.Respond(w, r, api.Error("User not found"))
		return
	}

	// if strings.ToLower(user.FirstName) != strings.ToLower(firstName) {
	// 	api.Respond(w, r, api.Error("First or last name not matching"))
	// 	return
	// }

	// if strings.ToLower(user.LastName) != strings.ToLower(lastName) {
	// 	api.Respond(w, r, api.Error("First or last name not matching"))
	// 	return
	// }

	token, errRandomFromGamma := str.RandomFromGamma(32, "BCDFGHJKLMNPQRSTVXYZ")

	if errRandomFromGamma != nil {
		api.Respond(w, r, api.Error("Error generating random string"))
		return
	}

	errTempTokenSave := a.funcTemporaryKeySet(token, userID, 3600)

	if errTempTokenSave != nil {
		api.Respond(w, r, api.Error("token store failed. "+errTempTokenSave.Error()))
		return
	}

	emailContent := a.funcEmailTemplatePasswordRestore(userID, a.LinkPasswordReset(token), UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	errEmailSent := a.funcEmailSend(userID, "Password Restore", emailContent)

	log.Println(errEmailSent)

	if errEmailSent != nil {
		api.Respond(w, r, api.Error("Password reset link failed to be sent. Please try again later"))
		return
	}

	api.Respond(w, r, api.Success("Password reset link was sent to your e-mail"))
}
