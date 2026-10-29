package page_password_restore

import (
	"log/slog"

	"github.com/dracory/auth/internal/ui/shared"
)

// Dependencies contains the dependencies required to render the password restore page.
type Dependencies struct {
	EnableRegistration bool

	Endpoint string

	Layout shared.LayoutFunc

	Logger *slog.Logger
}
