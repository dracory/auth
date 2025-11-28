package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/mail"

	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/str"
)

type RegisterUsernameAndPasswordResponse struct {
	ErrorMessage   string
	SuccessMessage string
	Token          string
}

func (a Auth) RegisterWithUsernameAndPassword(ctx context.Context, email string, password string, firstName string, lastName string, options UserAuthOptions) (response RegisterUsernameAndPasswordResponse) {
	if firstName == "" {
		response.ErrorMessage = "First name is required field"
		return response
	}

	if lastName == "" {
		response.ErrorMessage = "Last name is required field"
		return response
	}

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

	if a.funcUserRegister == nil {
		response.ErrorMessage = "registration failed. FuncUserRegister function not defined"
		return response
	}

	if !a.enableVerification {
		err := a.funcUserRegister(ctx, email, password, firstName, lastName, options)

		if err != nil {
			response.ErrorMessage = "registration failed."
			return response
		}

		response.SuccessMessage = "registration success"
		return response
	}

	verificationCode, errRandom := str.RandomFromGamma(
		authutils.LoginCodeLength(a.disableRateLimit),
		authutils.LoginCodeGamma(a.disableRateLimit),
	)
	if errRandom != nil {
		response.ErrorMessage = "Error generating random string"
		return response
	}

	json, errJson := json.Marshal(map[string]string{
		"email":      email,
		"first_name": firstName,
		"last_name":  lastName,
		"password":   password,
	})

	if errJson != nil {
		response.ErrorMessage = "Error serializing data"
		return response
	}

	errTempTokenSave := a.funcTemporaryKeySet(verificationCode, string(json), 3600)

	if errTempTokenSave != nil {
		response.ErrorMessage = "token store failed."
		return response
	}

	emailContent := a.funcEmailTemplateRegisterCode(ctx, email, verificationCode, options)

	errEmailSent := a.funcEmailSend(ctx, email, "Registration Code", emailContent)

	if errEmailSent != nil {
		log.Println(errEmailSent)
		response.ErrorMessage = "Registration code failed to be send. Please try again later"
		return response
	}

	response.SuccessMessage = "Registration code was sent successfully"
	return response
}
