package page_login

import (
	"log/slog"

	"github.com/dracory/auth/internal/ui/shared"
)

// Dependencies contains the dependencies required to render the login page.
type Dependencies struct {
	Passwordless       bool
	EnableRegistration bool

	Endpoint          string
	RedirectOnSuccess string

	// Layout is the outer layout function supplied by the auth package.
	Layout shared.LayoutFunc

	Logger *slog.Logger
}
