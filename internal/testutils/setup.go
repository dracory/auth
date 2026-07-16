package testutils

import (
	"context"

	"github.com/dracory/auth/types"
)

// NewUsernameAndPasswordConfigForTest returns a ConfigUsernameAndPassword
// pre-populated with safe defaults for tests.
func NewUsernameAndPasswordConfigForTest() types.ConfigUsernameAndPassword {
	endpoint := "http://localhost/auth"
	return types.ConfigUsernameAndPassword{
		ConfigShared: types.ConfigShared{
			Endpoint:             endpoint,
			UrlRedirectOnSuccess: "http://localhost/dashboard",
			UseCookies:           true,
			EnableRegistration:   true,
			FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
			FuncTemporaryKeySet:  func(key string, value string, expiresSeconds int) (err error) { return nil },
			FuncUserFindByAuthToken: func(ctx context.Context, token string, options types.UserAuthOptions) (userID string, err error) {
				return "", nil
			},
			FuncUserLogout: func(ctx context.Context, userID string, options types.UserAuthOptions) (err error) { return nil },
			FuncUserStoreAuthToken: func(ctx context.Context, sessionID string, userID string, options types.UserAuthOptions) error {
				return nil
			},
		},
		FuncUserFindByUsername: func(ctx context.Context, username string, firstName string, lastName string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogin: func(ctx context.Context, username string, password string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncEmailSend: func(ctx context.Context, userID string, emailSubject string, emailBody string) (err error) {
			return nil
		},
		FuncUserRegister: func(ctx context.Context, username string, password string, firstName string, lastName string, options types.UserAuthOptions) error {
			return nil
		},
		PasswordStrength: &types.PasswordStrengthConfig{},
	}
}

// NewPasswordlessConfigForTest returns a ConfigPasswordless pre-populated
// with safe defaults for tests.
func NewPasswordlessConfigForTest() types.ConfigPasswordless {
	endpoint := "http://localhost/auth"
	return types.ConfigPasswordless{
		ConfigShared: types.ConfigShared{
			Endpoint:             endpoint,
			UrlRedirectOnSuccess: "http://localhost/dashboard",
			UseCookies:           true,
			EnableRegistration:   true,
			FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
			FuncTemporaryKeySet:  func(key string, value string, expiresSeconds int) (err error) { return nil },
			FuncUserFindByAuthToken: func(ctx context.Context, token string, options types.UserAuthOptions) (userID string, err error) {
				return "111", nil
			},
			FuncUserLogout: func(ctx context.Context, userID string, options types.UserAuthOptions) (err error) { return nil },
			FuncUserStoreAuthToken: func(ctx context.Context, sessionID string, userID string, options types.UserAuthOptions) error {
				return nil
			},
		},
		FuncUserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (userID string, err error) {
			return "111", nil
		},
		FuncEmailSend: func(ctx context.Context, email string, emailSubject string, emailBody string) (err error) { return nil },
		FuncUserRegister: func(ctx context.Context, email string, firstName string, lastName string, options types.UserAuthOptions) error {
			return nil
		},
	}
}

// NewAuthKnightConfigForTest returns a ConfigAuthKnight pre-populated
// with safe defaults for tests.
func NewAuthKnightConfigForTest() types.ConfigAuthKnight {
	endpoint := "http://localhost/auth"
	return types.ConfigAuthKnight{
		ConfigShared: types.ConfigShared{
			Endpoint:             endpoint,
			UrlRedirectOnSuccess: "http://localhost/dashboard",
			UseCookies:           true,
			EnableRegistration:   true,
			FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
			FuncTemporaryKeySet:  func(key string, value string, expiresSeconds int) (err error) { return nil },
			FuncUserFindByAuthToken: func(ctx context.Context, token string, options types.UserAuthOptions) (userID string, err error) {
				return "111", nil
			},
			FuncUserLogout: func(ctx context.Context, userID string, options types.UserAuthOptions) (err error) { return nil },
			FuncUserStoreAuthToken: func(ctx context.Context, sessionID string, userID string, options types.UserAuthOptions) error {
				return nil
			},
		},
		FuncUserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (userID string, err error) {
			return "111", nil
		},
		FuncUserRegister: func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (userID string, err error) {
			return "222", nil
		},
	}
}
