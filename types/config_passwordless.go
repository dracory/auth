package types

import (
	"context"
)

type ConfigPasswordless struct {
	ConfigShared
	ConfigRateLimiting
	ConfigCSRF
	ConfigImpersonation

	// ===== START: passwordless options
	FuncUserFindByEmail           func(ctx context.Context, email string, options UserAuthOptions) (userID string, err error)
	FuncEmailTemplateLoginCode    func(ctx context.Context, email string, loginLink string, options UserAuthOptions) string    // optional
	FuncEmailTemplateRegisterCode func(ctx context.Context, email string, registerLink string, options UserAuthOptions) string // optional
	FuncEmailSend                 func(ctx context.Context, email string, emailSubject string, emailBody string) (err error)
	FuncUserRegister              func(ctx context.Context, email string, firstName string, lastName string, options UserAuthOptions) (err error)
	// ===== END: passwordless options
}
