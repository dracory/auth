package page_password_reset

import "log/slog"

// Dependencies contains the dependencies required to render the password reset page.
type Dependencies struct {
	Endpoint           string
	EnableRegistration bool
	Token              string
	ErrorMessage       string

	Layout func(content string) string

	Logger *slog.Logger
}
