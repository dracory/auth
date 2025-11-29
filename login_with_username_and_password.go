package auth

import (
	"context"

	"github.com/dracory/auth/types"
	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/str"
)

type LoginUsernameAndPasswordResponse struct {
	ErrorMessage   string
	SuccessMessage string
	Token          string
}

func (a authImplementation) LoginWithUsernameAndPassword(ctx context.Context, email string, password string, options types.UserAuthOptions) (response LoginUsernameAndPasswordResponse) {
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

	userID, err := a.funcUserLogin(ctx, email, password, options)

	if err != nil {
		response.ErrorMessage = "Invalid credentials"
		logger := a.GetLogger()
		logger.Error("login with username and password failed",
			"error", err,
			"email", email,
			"ip", options.UserIp,
			"user_agent", options.UserAgent,
		)
		return response
	}

	if userID == "" {
		response.ErrorMessage = "Invalid credentials"
		return response
	}

	token, errRandom := str.RandomFromGamma(32, LoginCodeGamma)
	if errRandom != nil {
		authErr := NewCodeGenerationError(errRandom)
		response.ErrorMessage = authErr.Message
		logger := a.GetLogger()
		logger.Error("auth token generation failed",
			"error", authErr.InternalErr,
			"error_code", authErr.Code,
			"email", email,
			"ip", options.UserIp,
			"user_agent", options.UserAgent,
		)
		return response
	}

	errSession := a.funcUserStoreAuthToken(ctx, token, userID, options)

	if errSession != nil {
		authErr := NewTokenStoreError(errSession)
		response.ErrorMessage = authErr.Message
		logger := a.GetLogger()
		logger.Error("auth token store failed",
			"error", authErr.InternalErr,
			"error_code", authErr.Code,
			"email", email,
			"user_id", userID,
			"ip", options.UserIp,
			"user_agent", options.UserAgent,
		)
		return response
	}

	response.SuccessMessage = "login success"
	response.Token = token
	return response
}
