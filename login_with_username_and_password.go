package auth

import (
	"context"

	"github.com/dracory/auth/internal/core"
	"github.com/dracory/auth/types"
	"github.com/dracory/str"
)

type LoginUsernameAndPasswordResponse struct {
	ErrorMessage   string
	SuccessMessage string
	Token          string
}

func (a authImplementation) LoginWithUsernameAndPassword(ctx context.Context, email string, password string, options types.UserAuthOptions) (response LoginUsernameAndPasswordResponse) {
	logger := a.GetLogger()

	deps := core.LoginWithUsernameAndPasswordDeps{
		FuncUserLogin:          a.funcUserLogin,
		FuncUserStoreAuthToken: a.funcUserStoreAuthToken,
		TokenGenerator: func() (string, error) {
			return str.RandomFromGamma(32, LoginCodeGamma)
		},
		Logger: logger,
		HandleCodeGenerationError: func(err error) (string, string) {
			authErr := NewCodeGenerationError(err)
			return authErr.Message, authErr.Code
		},
		HandleTokenStoreError: func(err error) (string, string) {
			authErr := NewTokenStoreError(err)
			return authErr.Message, authErr.Code
		},
	}

	res := core.LoginWithUsernameAndPassword(ctx, email, password, options, deps)
	return LoginUsernameAndPasswordResponse{
		ErrorMessage:   res.ErrorMessage,
		SuccessMessage: res.SuccessMessage,
		Token:          res.Token,
	}
}
