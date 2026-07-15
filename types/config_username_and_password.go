package types

import (
	"context"
)

// Config defines the available configuration options for authentication
type ConfigUsernameAndPassword struct {
	ConfigShared
	ConfigRateLimiting
	ConfigCSRF
	ConfigImpersonation

	// ===== START: username(email) and password options
	EnableVerification               bool
	FuncEmailTemplatePasswordRestore func(ctx context.Context, userID string, passwordRestoreLink string, options UserAuthOptions) string // optional
	FuncEmailTemplateRegisterCode    func(ctx context.Context, userID string, registerLink string, options UserAuthOptions) string        // optional
	FuncEmailSend                    func(ctx context.Context, userID string, emailSubject string, emailBody string) (err error)
	FuncUserFindByUsername           func(ctx context.Context, username string, firstName string, lastName string, options UserAuthOptions) (userID string, err error)
	FuncUserLogin                    func(ctx context.Context, username string, password string, options UserAuthOptions) (userID string, err error)
	FuncUserPasswordChange           func(ctx context.Context, username string, newPassword string, options UserAuthOptions) (err error)
	FuncUserRegister                 func(ctx context.Context, username string, password string, firstName string, lastName string, options UserAuthOptions) (err error)
	PasswordStrength                 *PasswordStrengthConfig
	LabelUsername                    string
	// ===== END: username(email) and password options
}
