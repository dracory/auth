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

// LoginWithUsernameAndPassword is a standalone helper that performs the
// username/password login flow using the provided AuthPasswordInterface
// implementation.
func LoginWithUsernameAndPassword(
	ctx context.Context,
	a types.AuthPasswordInterface,
	email string,
	password string,
	options types.UserAuthOptions,
) LoginUsernameAndPasswordResponse {
	res := core.LoginWithUsernameAndPassword(ctx, a, email, password, options)
	return LoginUsernameAndPasswordResponse{
		ErrorMessage:   res.ErrorMessage,
		SuccessMessage: res.SuccessMessage,
		Token:          res.Token,
	}
}

type RegisterUsernameAndPasswordResponse struct {
	ErrorMessage   string
	SuccessMessage string
	Token          string
}

// RegisterWithUsernameAndPassword is a standalone helper that performs the
// username/password registration flow using the provided AuthPasswordInterface
// implementation.
func RegisterWithUsernameAndPassword(
	ctx context.Context,
	a types.AuthPasswordInterface,
	email string,
	password string,
	firstName string,
	lastName string,
	options types.UserAuthOptions,
) RegisterUsernameAndPasswordResponse {
	res := core.RegisterWithUsernameAndPassword(
		ctx,
		email,
		password,
		firstName,
		lastName,
		options,
		a,
		DefaultVerificationCodeExpiration,
	)

	return RegisterUsernameAndPasswordResponse{
		ErrorMessage:   res.ErrorMessage,
		SuccessMessage: res.SuccessMessage,
		Token:          res.Token,
	}
}
