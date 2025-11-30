package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dracory/auth/types"
	authutils "github.com/dracory/auth/utils"
)

type RegisterWithUsernameAndPasswordResult struct {
	ErrorMessage   string
	SuccessMessage string
	Token          string
}

func RegisterWithUsernameAndPassword(
	ctx context.Context,
	email string,
	password string,
	firstName string,
	lastName string,
	options types.UserAuthOptions,
	a types.AuthPasswordInterface,
	verificationExpiration time.Duration,
) RegisterWithUsernameAndPasswordResult {
	var response RegisterWithUsernameAndPasswordResult

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

	if err := authutils.ValidatePasswordStrength(password, a.GetPasswordStrength()); err != nil {
		response.ErrorMessage = err.Error()
		return response
	}

	if msg := authutils.ValidateEmailFormat(email); msg != "" {
		response.ErrorMessage = msg
		return response
	}

	registerFn := a.GetFuncUserRegister()
	if registerFn == nil {
		response.ErrorMessage = "registration failed. FuncUserRegister function not defined"
		return response
	}

	if !a.IsVerificationEnabled() {
		if err := registerFn(ctx, email, password, firstName, lastName, options); err != nil {
			response.ErrorMessage = "registration failed."
			return response
		}

		response.SuccessMessage = "registration success"
		return response
	}

	logger := a.GetLogger()

	verificationCode, errRandom := authutils.GenerateVerificationCode(a.GetDisableRateLimit())
	if errRandom != nil {
		response.ErrorMessage = "Failed to generate verification code. Please try again later"
		if logger != nil {
			logger.Error("registration code generation failed",
				"error", errRandom,
				"error_code", "CODE_GENERATION_FAILED",
				"email", email,
				"ip", options.UserIp,
				"user_agent", options.UserAgent,
			)
		}
		return response
	}

	jsonPayload, errJson := json.Marshal(map[string]string{
		"email":      email,
		"first_name": firstName,
		"last_name":  lastName,
		"password":   password,
	})
	if errJson != nil {
		response.ErrorMessage = "Failed to process request. Please try again later"
		if logger != nil {
			logger.Error("registration data serialization failed",
				"error", errJson,
				"error_code", "SERIALIZATION_FAILED",
				"email", email,
				"ip", options.UserIp,
				"user_agent", options.UserAgent,
			)
		}
		return response
	}

	temporaryKeySet := a.GetFuncTemporaryKeySet()
	errTempTokenSave := temporaryKeySet(verificationCode, string(jsonPayload), int(verificationExpiration.Seconds()))
	if errTempTokenSave != nil {
		response.ErrorMessage = "Failed to process request. Please try again later"
		if logger != nil {
			logger.Error("registration code token store failed",
				"error", errTempTokenSave,
				"error_code", "TOKEN_STORE_FAILED",
				"email", email,
				"ip", options.UserIp,
				"user_agent", options.UserAgent,
			)
		}
		return response
	}

	emailTemplate := a.GetFuncEmailTemplateRegisterCode()
	if emailTemplate == nil {
		response.ErrorMessage = "registration failed. FuncEmailTemplateRegisterCode function not defined"
		return response
	}

	emailSend := a.GetFuncEmailSend()
	if emailSend == nil {
		response.ErrorMessage = "registration failed. FuncEmailSend function not defined"
		return response
	}

	emailContent := emailTemplate(ctx, email, verificationCode, options)

	if errEmailSent := emailSend(ctx, email, "Registration Code", emailContent); errEmailSent != nil {
		response.ErrorMessage = "Failed to send email. Please try again later"
		if logger != nil {
			logger.Error("registration email send failed",
				"error", errEmailSent,
				"error_code", "EMAIL_SEND_FAILED",
				"email", email,
				"ip", options.UserIp,
				"user_agent", options.UserAgent,
			)
		}
		return response
	}

	response.SuccessMessage = "Registration code was sent successfully"
	return response
}
