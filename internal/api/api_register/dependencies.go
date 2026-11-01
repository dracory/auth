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
	RegisterWithUsernameAndPassword func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (successMessage, nonce, errorMessage string)

	// CurrentUserID resolves the authenticated user ID from the request
	// (session cookie or middleware context). When it returns a non-empty
	// ID the endpoint treats registration as profile completion: the
	// email is already verified and the existing user is updated via
	// UserProfileUpdate instead of a new account being created.
	CurrentUserID func(r *http.Request) string

	// UserProfileUpdate persists the completion form (first/last name)
	// for an existing authenticated user.
	UserProfileUpdate func(ctx context.Context, userID, firstName, lastName string, options types.UserAuthOptions) error

	// RedirectOnSuccess is the URL returned to the client after a
	// successful profile completion.
	RedirectOnSuccess string

	// AuthKnight indicates whether the registration flow is AuthKnight-based.
	AuthKnight bool

	// VerifiedEmailRegisterDependencies are used for the verified-email
	// token branch — the AuthKnight callback and OTP/magic-link verifications.
	VerifiedEmailRegisterDependencies VerifiedEmailRegisterDependencies

	// Logger is used to log internal errors. If nil, slog.Default() is used.
	Logger *slog.Logger
}

// VerifiedEmailRegisterDependencies defines the dependencies required for
// the verified-email token registration flow (creating a user from a
// verified email).
type VerifiedEmailRegisterDependencies struct {
	// TemporaryKeyGet retrieves the verified email by the token
	// (verified_key).
	TemporaryKeyGet func(key string) (string, error)

	// TemporaryKeySet stores a temporary key-value pair with expiration.
	// Used to clear the token after successful registration.
	TemporaryKeySet func(key string, value string, expiresSeconds int) error

	// UserRegister creates a new user with email, first name, and last name.
	// The email is already verified by the issuing flow (AuthKnight, OTP,
	// or magic link).
	UserRegister func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (userID string, err error)

	// UserStoreAuthToken stores the auth token for the newly registered user.
	UserStoreAuthToken func(ctx context.Context, token, userID string, options types.UserAuthOptions) error

	// UseCookies controls whether the auth token should be written as a cookie.
	UseCookies bool

	// SetAuthCookie writes the auth cookie.
	SetAuthCookie func(w http.ResponseWriter, r *http.Request, token string)
}
