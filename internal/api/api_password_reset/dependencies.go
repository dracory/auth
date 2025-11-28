package api_password_reset

import (
	"context"

	"github.com/dracory/auth/types"
)

// Dependencies defines the dependencies required for the password reset
// flow (changing the user's password given a valid token).
type Dependencies struct {
	PasswordStrength *types.PasswordStrengthConfig

	TemporaryKeyGet func(key string) (string, error)

	UserPasswordChange func(ctx context.Context, userID, password string) error
	LogoutUser         func(ctx context.Context, userID string) error
}
