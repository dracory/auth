package auth

import (
	"context"
	"errors"

	authtypes "github.com/dracory/auth/types"
)

// testSetupUsernameAndPasswordAuth creates a new Auth for testing
func testSetupUsernameAndPasswordAuth() (*authImplementation, error) {
	endpoint := "http://localhost/auth"
	instance, err := NewUsernameAndPasswordAuth(ConfigUsernameAndPassword{
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
	if err != nil {
		return nil, err
	}
	auth, ok := instance.(*authImplementation)
	if !ok {
		return nil, errors.New("unexpected concrete type from NewUsernameAndPasswordAuth")
	}
	return auth, nil
}

// testSetupPasswordlessAuth creates a new Auth for testing
func testSetupPasswordlessAuth() (*authImplementation, error) {
	endpoint := "http://localhost/auth"
	instance, err := NewPasswordlessAuth(ConfigPasswordless{
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
	if err != nil {
		return nil, err
	}
	auth, ok := instance.(*authImplementation)
	if !ok {
		return nil, errors.New("unexpected concrete type from NewPasswordlessAuth")
	}
	return auth, nil
}
