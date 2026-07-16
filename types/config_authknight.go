package types

import (
	"context"
	"time"
)

// ConfigAuthKnight contains configuration for AuthKnight-based authentication.
// It embeds the shared config structs and adds AuthKnight-specific callbacks.
type ConfigAuthKnight struct {
	ConfigShared
	ConfigRateLimiting
	ConfigCSRF
	ConfigImpersonation

	// FuncUserFindByEmail finds a user by email and returns their ID.
	// Required.
	FuncUserFindByEmail func(ctx context.Context, email string, options UserAuthOptions) (userID string, err error)

	// FuncUserRegister creates a new user with email, first name, and last name.
	// The email is already verified by AuthKnight — no verification step needed.
	// Called when the user submits the register form (first name + last name).
	// Required.
	FuncUserRegister func(ctx context.Context, email string, firstName string, lastName string, options UserAuthOptions) (userID string, err error)

	// FuncRedirectURL calculates the redirect URL after successful auth.
	// If nil, falls back to UrlRedirectOnSuccess.
	FuncRedirectURL func(ctx context.Context, userID string) string

	// HTTPTimeout for calls to the AuthKnight API. Defaults to 10 seconds.
	HTTPTimeout time.Duration
}
