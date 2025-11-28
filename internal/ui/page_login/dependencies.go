package page_login

import "log/slog"

// Dependencies contains the dependencies required to render the login page.
type Dependencies struct {
	Passwordless       bool
	EnableRegistration bool

	Endpoint          string
	RedirectOnSuccess string

	// Layout is the outer layout function supplied by the auth package.
	Layout func(content string) string

	Logger *slog.Logger
}
