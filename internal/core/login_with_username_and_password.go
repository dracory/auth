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
		response.ErrorMessage = "Email is required field"
		return response
	}

	if password == "" {
		response.ErrorMessage = "Password is required field"
		return response
	}

	if msg := authutils.ValidateEmailFormat(email); msg != "" {
		response.ErrorMessage = msg
		return response
	}

	loginFn := a.GetFuncUserLogin()
	storeFn := a.GetFuncUserStoreAuthToken()
	logger := a.GetLogger()

	userID, err := loginFn(ctx, email, password, options)

	if err != nil {
		response.ErrorMessage = "Invalid credentials"
		if logger != nil {
			logger.Error("login with username and password failed",
				"error", err,
				"email", email,
				"ip", options.UserIp,
				"user_agent", options.UserAgent,
			)
		}
		return response
	}

	if userID == "" {
		response.ErrorMessage = "Invalid credentials"
		return response
	}

	token, errRandom := str.RandomFromGamma(32, "BCDFGHJKLMNPQRSTVXYZ")
	if errRandom != nil {
		response.ErrorMessage = "Failed to generate verification code. Please try again later"
		if logger != nil {
			logger.Error("auth token generation failed",
				"error", errRandom,
				"error_code", "CODE_GENERATION_FAILED",
				"email", email,
				"ip", options.UserIp,
				"user_agent", options.UserAgent,
			)
		}
		return response
	}

	errSession := storeFn(ctx, token, userID, options)

	if errSession != nil {
		response.ErrorMessage = "Failed to process request. Please try again later"
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
		return response
	}

	response.SuccessMessage = "login success"
	response.Token = token
	return response
}
