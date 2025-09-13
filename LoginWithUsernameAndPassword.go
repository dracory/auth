package auth

import (
	"github.com/dracory/str"
	validator "github.com/gouniverse/validator"
)

type LoginUsernameAndPasswordResponse struct {
	ErrorMessage   string
	SuccessMessage string
	Token          string
}

func (a Auth) LoginWithUsernameAndPassword(email string, password string, options UserAuthOptions) (response LoginUsernameAndPasswordResponse) {
	if email == "" {
		response.ErrorMessage = "Email is required field"
		return response
	}

	if password == "" {
		response.ErrorMessage = "Password is required field"
		return response
	}

	if !validator.IsEmail(email) {
		response.ErrorMessage = "This is not a valid email: " + email
		return response
	}

	userID, err := a.funcUserLogin(email, password, options)

	if err != nil {
		response.ErrorMessage = "authentication failed. " + err.Error()
		return response
	}

	if userID == "" {
		response.ErrorMessage = "User not found"
		return response
	}

	token := str.RandomFromGamma(32, LoginCodeGamma)

	errSession := a.funcUserStoreAuthToken(token, userID, options)

	if errSession != nil {
		response.ErrorMessage = "token store failed. " + errSession.Error()
		return response
	}

	response.SuccessMessage = "login success"
	response.Token = token
	return response
}
