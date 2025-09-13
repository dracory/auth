package auth

import (
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/req"
)

func (a Auth) apiPasswordReset(w http.ResponseWriter, r *http.Request) {
	token := req.GetStringTrimmed(r, "token")
	password := req.GetStringTrimmed(r, "password")
	passwordConfirm := req.GetStringTrimmed(r, "password_confirm")

	if token == "" {
		api.Respond(w, r, api.Error("Token is required field"))
		return
	}

	if password == "" {
		api.Respond(w, r, api.Error("Password is required field"))
		return
	}

	if password != passwordConfirm {
		api.Respond(w, r, api.Error("Passwords do not match"))
		return
	}

	userID, errToken := a.funcTemporaryKeyGet(token)

	if errToken != nil {
		api.Respond(w, r, api.Error("Link not valid of expired"))
		return
	}

	if userID == "" {
		api.Respond(w, r, api.Error("Link not valid of expired"))
		return
	}

	errPasswordChange := a.funcUserPasswordChange(userID, password, UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	if errPasswordChange != nil {
		api.Respond(w, r, api.Error("authentication failed. "+errPasswordChange.Error()))
		return
	}

	api.Respond(w, r, api.SuccessWithData("login success", map[string]interface{}{
		"token": token,
	}))
}
