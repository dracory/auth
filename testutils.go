package auth

import (
	"context"

	authtypes "github.com/dracory/auth/types"
)

// testSetupUsernameAndPasswordAuth creates a new Auth for testing
func testSetupUsernameAndPasswordAuth() (*Auth, error) {
	endpoint := "http://localhost/auth"
	return NewUsernameAndPasswordAuth(ConfigUsernameAndPassword{
		Endpoint:             endpoint,
		UrlRedirectOnSuccess: "http://localhost/dashboard",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key string, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, token string, options UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserFindByUsername: func(ctx context.Context, username string, firstName string, lastName string, options UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogin: func(ctx context.Context, username string, password string, options UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogout: func(ctx context.Context, userID string, options UserAuthOptions) (err error) { return nil },
		FuncUserStoreAuthToken: func(ctx context.Context, sessionID string, userID string, options UserAuthOptions) error {
			return nil
		},
		FuncEmailSend: func(ctx context.Context, userID string, emailSubject string, emailBody string) (err error) {
			return nil
		},
		PasswordStrength: &authtypes.PasswordStrengthConfig{},
		UseCookies:       true,
	})
}

// testSetupPasswordlessAuth creates a new Auth for testing
func testSetupPasswordlessAuth() (*Auth, error) {
	endpoint := "http://localhost/auth"
	return NewPasswordlessAuth(ConfigPasswordless{
		Endpoint:             endpoint,
		UrlRedirectOnSuccess: "http://localhost/dashboard",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key string, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, token string, options UserAuthOptions) (userID string, err error) {
			return "111", nil
		},
		FuncUserFindByEmail: func(ctx context.Context, email string, options UserAuthOptions) (userID string, err error) {
			return "111", nil
		},
		FuncUserLogout: func(ctx context.Context, userID string, options UserAuthOptions) (err error) { return nil },
		FuncUserStoreAuthToken: func(ctx context.Context, sessionID string, userID string, options UserAuthOptions) error {
			return nil
		},
		FuncEmailSend: func(ctx context.Context, email string, emailSubject string, emailBody string) (err error) { return nil },
		UseCookies:    true,
	})
}
