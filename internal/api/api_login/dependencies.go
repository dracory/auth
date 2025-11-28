package api_login

import (
	"context"
	"net/http"
)

// Dependencies aggregates all dependencies required for handling the /api/login
// endpoint for both passwordless and username+password flows.
type Dependencies struct {
	// Passwordless controls which flow is executed. When true, the passwordless
	// email-code flow is used; otherwise the username+password flow is used.
	Passwordless bool

	// PasswordlessDependencies contains the business-logic dependencies for the
	// passwordless login flow.
	PasswordlessDependencies LoginPasswordlessDeps

	// LoginWithUsernameAndPassword performs the username+password login flow
	// and returns success message, token and error message. If error message is
	// non-empty, the operation is considered failed.
	LoginWithUsernameAndPassword func(
		ctx context.Context,
		email, password, ip, userAgent string,
	) (successMessage, token, errorMessage string)

	// UseCookies controls whether the auth token should be written as a cookie
	// when the username+password flow succeeds.
	UseCookies bool

	// SetAuthCookie writes the auth cookie. It is only used when UseCookies is
	// true and must be non-nil in that case.
	SetAuthCookie func(w http.ResponseWriter, r *http.Request, token string)
}
