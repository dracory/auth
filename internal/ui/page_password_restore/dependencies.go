package page_password_restore

import "log/slog"

// Dependencies contains the dependencies required to render the password restore page.
type Dependencies struct {
	EnableRegistration bool

	Endpoint string

	Layout func(content string) string

	Logger *slog.Logger
}
