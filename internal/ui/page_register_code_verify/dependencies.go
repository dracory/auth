package page_register_code_verify

import "log/slog"

// Dependencies contains the dependencies required to render the register code verify
// page.
type Dependencies struct {
	Endpoint          string
	RedirectOnSuccess string

	Layout func(content string) string

	Logger *slog.Logger
}
