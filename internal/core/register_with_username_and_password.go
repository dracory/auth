package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dracory/auth/internal/rules"
	"github.com/dracory/auth/types"
	authutils "github.com/dracory/auth/utils"
)

type RegisterWithUsernameAndPasswordResult struct {
	ErrorMessage   string
	SuccessMessage string
	Token          string
	// Nonce identifies the temporary-key under which the registration
	// payload (including the code) is stored when verification is enabled.
	Nonce string
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
		response.ErrorMessage = types.MsgFirstNameRequired
		return response
	}

	if lastName == "" {
		response.ErrorMessage = types.MsgLastNameRequired
		return response
	}

	if email == "" {
		response.ErrorMessage = types.MsgEmailRequired
		return response
	}

	if password == "" {
		response.ErrorMessage = types.MsgPasswordRequired
		return response
	}

	if check := rules.NewPasswordStrengthRule(password, a.GetPasswordStrength()); check.Fails() {
		response.ErrorMessage = check.FailMessageFirst()
		return response
	}

	if check := rules.NewEmailFormatRule(email); check.Fails() {
		response.ErrorMessage = check.FailMessageFirst()
		return response
	}

	registerFn := a.GetFuncUserRegister()
	if registerFn == nil {
		response.ErrorMessage = types.MsgRegistrationFailedFn
		a.GetObservabilityHooks().RecordRegistrationAttempt(false, nil)
		return response
	}

	if !a.IsVerificationEnabled() {
		if err := registerFn(ctx, email, password, firstName, lastName, options); err != nil {
			response.ErrorMessage = types.MsgRegistrationFailed
			a.GetObservabilityHooks().RecordRegistrationAttempt(false, err)
			return response
		}

		response.SuccessMessage = types.MsgRegistrationSuccess
		a.GetObservabilityHooks().RecordRegistrationAttempt(true, nil)
		return response
	}

	logger := a.GetLogger()
	hooks := a.GetObservabilityHooks()

	// Per-email send throttle — prevents the endpoint being used for email
	// bombing; enforced before code generation.
	if check := rules.NewEmailSendThrottleRule(a.GetFuncTemporaryKeyGet(), email); check.Fails() {
		response.ErrorMessage = check.FailMessageFirst()
		hooks.RecordRegistrationAttempt(false, nil)
		return response
	}

	verificationCode, errRandom := authutils.GenerateVerificationCode(a.GetDisableRateLimit())
	if errRandom != nil {
		response.ErrorMessage = types.MsgFailedToGenerateCode
		if logger != nil {
			logger.Error("registration code generation failed",
				"error", errRandom,
				"error_code", "CODE_GENERATION_FAILED",
				"email", email,
				"ip", options.UserIp,
				"user_agent", options.UserAgent,
			)
		}
		hooks.RecordRegistrationAttempt(false, errRandom)
		return response
	}

	nonce, errNonce := authutils.GenerateNonce()
	if errNonce != nil {
		response.ErrorMessage = types.MsgFailedToGenerateCode
		if logger != nil {
			logger.Error("registration nonce generation failed",
				"error", errNonce,
				"error_code", "NONCE_GENERATION_FAILED",
				"email", email,
				"ip", options.UserIp,
				"user_agent", options.UserAgent,
			)
		}
		hooks.RecordRegistrationAttempt(false, errNonce)
		return response
	}

	jsonPayload, errJson := json.Marshal(map[string]string{
		"email":      email,
		"first_name": firstName,
		"last_name":  lastName,
		"password":   password,
		"code":       verificationCode,
	})
	if errJson != nil {
		response.ErrorMessage = types.MsgFailedToProcess
		if logger != nil {
			logger.Error("registration data serialization failed",
				"error", errJson,
				"error_code", "SERIALIZATION_FAILED",
				"email", email,
				"ip", options.UserIp,
				"user_agent", options.UserAgent,
			)
		}
		hooks.RecordRegistrationAttempt(false, errJson)
		return response
	}

	temporaryKeySet := a.GetFuncTemporaryKeySet()
	errTempTokenSave := temporaryKeySet(authutils.RegisterCodeKeyPrefix+nonce, string(jsonPayload), int(verificationExpiration.Seconds()))
	if errTempTokenSave != nil {
		response.ErrorMessage = types.MsgFailedToProcess
		if logger != nil {
			logger.Error("registration code token store failed",
				"error", errTempTokenSave,
				"error_code", "TOKEN_STORE_FAILED",
				"email", email,
				"ip", options.UserIp,
				"user_agent", options.UserAgent,
			)
		}
		hooks.RecordRegistrationAttempt(false, errTempTokenSave)
		return response
	}

	// Seed the per-nonce attempt counter alongside the payload.
	if errAttempts := temporaryKeySet(authutils.RegisterCodeAttemptsKeyPrefix+nonce, "0", int(verificationExpiration.Seconds())); errAttempts != nil {
		response.ErrorMessage = types.MsgFailedToProcess
		hooks.RecordRegistrationAttempt(false, errAttempts)
		return response
	}

	emailTemplate := a.GetFuncEmailTemplateRegisterCode()
	if emailTemplate == nil {
		response.ErrorMessage = types.MsgRegistrationFailedEmailTpl
		hooks.RecordRegistrationAttempt(false, nil)
		return response
	}

	emailSend := a.GetFuncEmailSend()
	if emailSend == nil {
		response.ErrorMessage = types.MsgRegistrationFailedEmailSend
		hooks.RecordRegistrationAttempt(false, nil)
		return response
	}

	emailContent := emailTemplate(ctx, email, verificationCode, options)

	if errEmailSent := emailSend(ctx, email, types.EmailSubjectRegistrationCode, emailContent); errEmailSent != nil {
		response.ErrorMessage = types.MsgFailedToSendEmail
		if logger != nil {
			logger.Error("registration email send failed",
				"error", errEmailSent,
				"error_code", "EMAIL_SEND_FAILED",
				"email", email,
				"ip", options.UserIp,
				"user_agent", options.UserAgent,
			)
		}
		hooks.RecordRegistrationAttempt(false, errEmailSent)
		return response
	}

	response.SuccessMessage = types.MsgRegistrationCodeSent
	response.Nonce = nonce

	// Consume send quota only after the email was handed off successfully.
	if get := a.GetFuncTemporaryKeyGet(); get != nil {
		_ = authutils.RecordEmailSend(get, temporaryKeySet, email, int(verificationExpiration.Seconds()))
	}

	hooks.RecordRegistrationAttempt(true, nil)
	return response
}
