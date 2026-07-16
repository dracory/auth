package api_authknight_callback

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/dracory/auth/types"
)

// Dependencies defines the dependencies required for the AuthKnight callback handler.
type Dependencies struct {
	// BaseURL is the AuthKnight base URL (e.g. "https://authknight.com").
	BaseURL string

	// HTTPTimeout for calls to the AuthKnight API. Defaults to 10 seconds.
	HTTPTimeout time.Duration

	// UserFindByEmail finds a user by email and returns their ID.
	UserFindByEmail func(ctx context.Context, email string, options types.UserAuthOptions) (userID string, err error)

	// UserStoreAuthToken stores the auth token for a user.
	UserStoreAuthToken func(ctx context.Context, token, userID string, options types.UserAuthOptions) error

	// TemporaryKeySet stores a temporary key-value pair with expiration.
	TemporaryKeySet func(key string, value string, expiresSeconds int) error

	// TemporaryKeyGet retrieves a temporary value by key.
	TemporaryKeyGet func(key string) (string, error)

	// RedirectURL calculates the redirect URL after successful auth.
	// If nil, falls back to UrlRedirectOnSuccess.
	RedirectURL func(ctx context.Context, userID string) string

	// UrlRedirectOnSuccess is the default redirect URL.
	UrlRedirectOnSuccess string

	// RegisterURL is the URL of the register page (with ?ak_key= appended).
	RegisterURL string

	// UseCookies controls whether the auth token should be written as a cookie.
	UseCookies bool

	// SetAuthCookie writes the auth cookie.
	SetAuthCookie func(w http.ResponseWriter, r *http.Request, token string)

	// Logger is used to log internal errors. If nil, slog.Default() is used.
	Logger *slog.Logger
}
