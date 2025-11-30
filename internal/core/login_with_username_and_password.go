package core

import (
	"context"
	"log/slog"

	"github.com/dracory/auth/types"
	authutils "github.com/dracory/auth/utils"
)

type LoginWithUsernameAndPasswordDeps struct {
	FuncUserLogin             func(ctx context.Context, username string, password string, options types.UserAuthOptions) (userID string, err error)
	FuncUserStoreAuthToken    func(ctx context.Context, token string, userID string, options types.UserAuthOptions) error
	TokenGenerator            func() (string, error)
	Logger                    *slog.Logger
	HandleCodeGenerationError func(err error) (message string, code string)
	HandleTokenStoreError     func(err error) (message string, code string)
}

type LoginWithUsernameAndPasswordResult struct {
	ErrorMessage   string
	SuccessMessage string
	Token          string
}

func LoginWithUsernameAndPassword(
	ctx context.Context,
	email string,
	password string,
	options types.UserAuthOptions,
	deps LoginWithUsernameAndPasswordDeps,
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

	userID, err := deps.FuncUserLogin(ctx, email, password, options)

	if err != nil {
		response.ErrorMessage = "Invalid credentials"
		logger := deps.Logger
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

	token, errRandom := deps.TokenGenerator()
	if errRandom != nil {
		msg, code := deps.HandleCodeGenerationError(errRandom)
		response.ErrorMessage = msg
		logger := deps.Logger
		if logger != nil {
			logger.Error("auth token generation failed",
				"error", errRandom,
				"error_code", code,
				"email", email,
				"ip", options.UserIp,
				"user_agent", options.UserAgent,
			)
		}
		return response
	}

	errSession := deps.FuncUserStoreAuthToken(ctx, token, userID, options)

	if errSession != nil {
		msg, code := deps.HandleTokenStoreError(errSession)
		response.ErrorMessage = msg
		logger := deps.Logger
		if logger != nil {
			logger.Error("auth token store failed",
				"error", errSession,
				"error_code", code,
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
