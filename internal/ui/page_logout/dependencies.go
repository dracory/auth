package page_logout

import "log/slog"

// Dependencies contains the dependencies required to render the logout page.
type Dependencies struct {
	Endpoint string

	Layout func(content string) string

	Logger *slog.Logger
}
