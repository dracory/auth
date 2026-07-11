package core

import (
	"context"

	"github.com/dracory/auth/types"
	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/str"
)

type LoginWithUsernameAndPasswordResult struct {
	ErrorMessage   string
	SuccessMessage string
	Token          string
}

func LoginWithUsernameAndPassword(
	ctx context.Context,
	a types.AuthPasswordInterface,
	email string,
	password string,
	options types.UserAuthOptions,
) LoginWithUsernameAndPasswordResult {
	var response LoginWithUsernameAndPasswordResult

	if email == "" {
		response.ErrorMessage = types.MsgEmailRequired
		return response
	}

	if password == "" {
		response.ErrorMessage = types.MsgPasswordRequired
		return response
	}

	if msg := authutils.ValidateEmailFormat(email); msg != "" {
		response.ErrorMessage = msg
		return response
	}

	loginFn := a.GetFuncUserLogin()
	storeFn := a.GetFuncUserStoreAuthToken()
	logger := a.GetLogger()
	hooks := a.GetObservabilityHooks()

	userID, err := loginFn(ctx, email, password, options)

	if err != nil {
		response.ErrorMessage = types.MsgInvalidCredentials
		if logger != nil {
			logger.Error("login with username and password failed",
				"error", err,
				"email", email,
				"ip", options.UserIp,
				"user_agent", options.UserAgent,
			)
		}
		hooks.RecordLoginAttempt(types.LoginMethodPassword, false, err)
		return response
	}

	if userID == "" {
		response.ErrorMessage = types.MsgInvalidCredentials
		hooks.RecordLoginAttempt(types.LoginMethodPassword, false, nil)
		return response
	}

	// Reuse the shared login/verification gamma from utils to avoid
	// duplicating the character set. We keep length at 32 for auth tokens.
	token, errRandom := str.RandomFromGamma(32, authutils.LoginCodeGamma(false))
	if errRandom != nil {
		response.ErrorMessage = types.MsgFailedToGenerateCode
		if logger != nil {
			logger.Error("auth token generation failed",
				"error", errRandom,
				"error_code", "CODE_GENERATION_FAILED",
				"email", email,
				"ip", options.UserIp,
				"user_agent", options.UserAgent,
			)
		}
		hooks.RecordLoginAttempt(types.LoginMethodPassword, false, errRandom)
		return response
	}

	errSession := storeFn(ctx, token, userID, options)

	if errSession != nil {
		response.ErrorMessage = types.MsgFailedToProcess
		if logger != nil {
			logger.Error("auth token store failed",
				"error", errSession,
				"error_code", "TOKEN_STORE_FAILED",
				"email", email,
				"user_id", userID,
				"ip", options.UserIp,
				"user_agent", options.UserAgent,
			)
		}
		hooks.RecordLoginAttempt(types.LoginMethodPassword, false, errSession)
		return response
	}

	response.SuccessMessage = types.MsgLoginSuccess
	response.Token = token
	hooks.RecordLoginAttempt(types.LoginMethodPassword, true, nil)
	hooks.RecordSessionCreated(userID)
	return response
}
