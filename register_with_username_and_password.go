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
	logger := a.GetLogger()

	deps := core.RegisterWithUsernameAndPasswordDeps{
		EnableVerification:            a.enableVerification,
		DisableRateLimit:              a.disableRateLimit,
		PasswordStrength:              a.passwordStrength,
		VerificationCodeExpiration:    DefaultVerificationCodeExpiration,
		FuncUserRegister:              a.funcUserRegister,
		FuncTemporaryKeySet:           a.funcTemporaryKeySet,
		FuncEmailTemplateRegisterCode: a.funcEmailTemplateRegisterCode,
		FuncEmailSend:                 a.funcEmailSend,
		Logger:                        logger,
		HandleCodeGenerationError: func(err error) (string, string) {
			v := NewCodeGenerationError(err)
			return v.Message, v.Code
		},
		HandleSerializationError: func(err error) (string, string) {
			v := NewSerializationError(err)
			return v.Message, v.Code
		},
		HandleTokenStoreError: func(err error) (string, string) {
			v := NewTokenStoreError(err)
			return v.Message, v.Code
		},
		HandleEmailSendError: func(err error) (string, string) {
			v := NewEmailSendError(err)
			return v.Message, v.Code
		},
	}

	res := core.RegisterWithUsernameAndPassword(ctx, email, password, firstName, lastName, options, deps)

	return RegisterUsernameAndPasswordResponse{
		ErrorMessage:   res.ErrorMessage,
		SuccessMessage: res.SuccessMessage,
		Token:          res.Token,
	}
}
