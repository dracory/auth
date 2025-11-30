package auth

import (
	"context"

	"github.com/dracory/auth/internal/core"
	"github.com/dracory/auth/types"
)

type LoginUsernameAndPasswordResponse struct {
	ErrorMessage   string
	SuccessMessage string
	Token          string
}

func (a authImplementation) LoginWithUsernameAndPassword(ctx context.Context, email string, password string, options types.UserAuthOptions) (response LoginUsernameAndPasswordResponse) {
	res := core.LoginWithUsernameAndPassword(ctx, &a, email, password, options)
	return LoginUsernameAndPasswordResponse{
		ErrorMessage:   res.ErrorMessage,
		SuccessMessage: res.SuccessMessage,
		Token:          res.Token,
	}
}
