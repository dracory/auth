package api_logout

import (
	"context"
	"net/http"
)

// Dependencies defines the dependencies required for performing a logout.
// The concrete implementation details (how to look up a user from a token
// and how to log a user out) are provided by the caller via these
// function fields.
type Dependencies struct {
	// UserFromToken resolves a user ID from the provided authentication token.
	UserFromToken func(ctx context.Context, token string) (userID string, err error)

	// LogoutUser performs the actual user logout.
	LogoutUser func(ctx context.Context, userID string) error

	// UseCookies indicates whether the auth token is stored in cookies and
	// should be removed on successful logout.
	UseCookies bool

	// AuthTokenRetrieve extracts the authentication token from the HTTP
	// request, typically from cookies or headers.
	AuthTokenRetrieve func(r *http.Request, useCookies bool) string

	// RemoveAuthCookie removes the authentication cookie after a successful
	// logout when UseCookies is true.
	RemoveAuthCookie func(w http.ResponseWriter, r *http.Request)
}
