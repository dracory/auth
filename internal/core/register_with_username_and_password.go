package core

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/dracory/auth/types"
	authutils "github.com/dracory/auth/utils"
)

type RegisterWithUsernameAndPasswordDeps struct {
	EnableVerification         bool
	DisableRateLimit           bool
	PasswordStrength           *types.PasswordStrengthConfig
	VerificationCodeExpiration time.Duration

	FuncUserRegister              func(ctx context.Context, username, password, firstName, lastName string, options types.UserAuthOptions) error
	FuncTemporaryKeySet           func(key string, value string, expiresSeconds int) error
	FuncEmailTemplateRegisterCode func(ctx context.Context, email string, passwordRestoreLink string, options types.UserAuthOptions) string
	FuncEmailSend                 func(ctx context.Context, userID string, emailSubject string, emailBody string) error

	Logger *slog.Logger

	HandleCodeGenerationError func(err error) (message string, code string)
	HandleSerializationError  func(err error) (message string, code string)
	HandleTokenStoreError     func(err error) (message string, code string)
	HandleEmailSendError      func(err error) (message string, code string)
}

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
	deps RegisterWithUsernameAndPasswordDeps,
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

	if err := authutils.ValidatePasswordStrength(password, deps.PasswordStrength); err != nil {
		response.ErrorMessage = err.Error()
		return response
	}

	if msg := authutils.ValidateEmailFormat(email); msg != "" {
		response.ErrorMessage = msg
		return response
	}

	if deps.FuncUserRegister == nil {
		response.ErrorMessage = "registration failed. FuncUserRegister function not defined"
		return response
	}

	if !deps.EnableVerification {
		if err := deps.FuncUserRegister(ctx, email, password, firstName, lastName, options); err != nil {
			response.ErrorMessage = "registration failed."
			return response
		}

		response.SuccessMessage = "registration success"
		return response
	}

	verificationCode, errRandom := authutils.GenerateVerificationCode(deps.DisableRateLimit)
	if errRandom != nil {
		msg, code := deps.HandleCodeGenerationError(errRandom)
		response.ErrorMessage = msg
		logger := deps.Logger
		if logger != nil {
			logger.Error("registration code generation failed",
				"error", errRandom,
				"error_code", code,
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
		msg, code := deps.HandleSerializationError(errJson)
		response.ErrorMessage = msg
		logger := deps.Logger
		if logger != nil {
			logger.Error("registration data serialization failed",
				"error", errJson,
				"error_code", code,
				"email", email,
				"ip", options.UserIp,
				"user_agent", options.UserAgent,
			)
		}
		return response
	}

	errTempTokenSave := deps.FuncTemporaryKeySet(verificationCode, string(jsonPayload), int(deps.VerificationCodeExpiration.Seconds()))
	if errTempTokenSave != nil {
		msg, code := deps.HandleTokenStoreError(errTempTokenSave)
		response.ErrorMessage = msg
		logger := deps.Logger
		if logger != nil {
			logger.Error("registration code token store failed",
				"error", errTempTokenSave,
				"error_code", code,
				"email", email,
				"ip", options.UserIp,
				"user_agent", options.UserAgent,
			)
		}
		return response
	}

	emailContent := deps.FuncEmailTemplateRegisterCode(ctx, email, verificationCode, options)

	if errEmailSent := deps.FuncEmailSend(ctx, email, "Registration Code", emailContent); errEmailSent != nil {
		msg, code := deps.HandleEmailSendError(errEmailSent)
		response.ErrorMessage = msg
		logger := deps.Logger
		if logger != nil {
			logger.Error("registration email send failed",
				"error", errEmailSent,
				"error_code", code,
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
