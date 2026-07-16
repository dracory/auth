package api_register

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dracory/auth/types"
)

// Dependencies defines the dependencies required for handling the registration
// API endpoint. It combines passwordless, username+password, and AuthKnight flows
// behind a shared interface.
type Dependencies struct {
	// Passwordless indicates whether the registration flow is passwordless.
	Passwordless bool

	// RegisterPasswordlessInitDependencies are used when Passwordless is true.
	RegisterPasswordlessInitDependencies RegisterPasswordlessInitDependencies

	// RegisterWithUsernameAndPassword performs the username+password
	// registration when Passwordless is false. It is responsible for all
	// validation and business rules, and returns a user-facing success or
	// error message.
	RegisterWithUsernameAndPassword func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (successMessage, errorMessage string)

	// AuthKnight indicates whether the registration flow is AuthKnight-based.
	AuthKnight bool

	// AuthKnightRegisterDependencies are used when AuthKnight is true.
	AuthKnightRegisterDependencies AuthKnightRegisterDependencies

	// Logger is used to log internal errors. If nil, slog.Default() is used.
	Logger *slog.Logger
}

// AuthKnightRegisterDependencies defines the dependencies required for
// the AuthKnight registration flow (creating a user from a verified email).
type AuthKnightRegisterDependencies struct {
	// TemporaryKeyGet retrieves the verified email by the ak_key.
	TemporaryKeyGet func(key string) (string, error)

	// TemporaryKeySet stores a temporary key-value pair with expiration.
	// Used to clear the ak_key after successful registration.
	TemporaryKeySet func(key string, value string, expiresSeconds int) error

	// UserRegister creates a new user with email, first name, and last name.
	// The email is already verified by AuthKnight.
	UserRegister func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (userID string, err error)

	// UserStoreAuthToken stores the auth token for the newly registered user.
	UserStoreAuthToken func(ctx context.Context, token, userID string, options types.UserAuthOptions) error

	// UseCookies controls whether the auth token should be written as a cookie.
	UseCookies bool

	// SetAuthCookie writes the auth cookie.
	SetAuthCookie func(w http.ResponseWriter, r *http.Request, token string)
}
