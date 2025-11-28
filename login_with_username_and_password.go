package auth

import (
	"context"
	"log"
	"net/mail"

	"github.com/dracory/str"
)

type LoginUsernameAndPasswordResponse struct {
	ErrorMessage   string
	SuccessMessage string
	Token          string
}

func (a Auth) LoginWithUsernameAndPassword(ctx context.Context, email string, password string, options UserAuthOptions) (response LoginUsernameAndPasswordResponse) {
	if email == "" {
		response.ErrorMessage = "Email is required field"
		return response
	}

	if password == "" {
		response.ErrorMessage = "Password is required field"
		return response
	}

	if _, err := mail.ParseAddress(email); err != nil {
		response.ErrorMessage = "This is not a valid email: " + email
		return response
	}

	userID, err := a.funcUserLogin(ctx, email, password, options)

	if err != nil {
		response.ErrorMessage = "authentication failed."
		log.Println("login failed:", err)
		return response
	}

	if userID == "" {
		response.ErrorMessage = "User not found"
		return response
	}

	token, errRandom := str.RandomFromGamma(32, LoginCodeGamma)
	if errRandom != nil {
		response.ErrorMessage = "token generation failed."
		log.Println("token generation failed:", errRandom)
		return response
	}

	errSession := a.funcUserStoreAuthToken(ctx, token, userID, options)

	if errSession != nil {
		response.ErrorMessage = "token store failed."
		log.Println("token store failed:", errSession)
		return response
	}

	response.SuccessMessage = "login success"
	response.Token = token
	return response
}
