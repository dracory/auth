package page_login_code_verify

import (
	"log/slog"

	"github.com/dracory/auth/internal/ui/shared"
)

// Dependencies contains the dependencies required to render the login code verify
// page.
type Dependencies struct {
	Endpoint          string
	RedirectOnSuccess string

	Layout shared.LayoutFunc

	Logger *slog.Logger
}
