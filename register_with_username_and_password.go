package auth

import (
	"context"

	"github.com/dracory/auth/internal/core"
	"github.com/dracory/auth/types"
)

type RegisterUsernameAndPasswordResponse struct {
	ErrorMessage   string
	SuccessMessage string
	Token          string
}

func (a authImplementation) RegisterWithUsernameAndPassword(ctx context.Context, email string, password string, firstName string, lastName string, options types.UserAuthOptions) (response RegisterUsernameAndPasswordResponse) {
	res := core.RegisterWithUsernameAndPassword(ctx, email, password, firstName, lastName, options, &a, DefaultVerificationCodeExpiration)

	return RegisterUsernameAndPasswordResponse{
		ErrorMessage:   res.ErrorMessage,
		SuccessMessage: res.SuccessMessage,
		Token:          res.Token,
	}
}
