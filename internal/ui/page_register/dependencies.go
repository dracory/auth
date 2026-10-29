package page_register

import (
	"log/slog"

	"github.com/dracory/auth/internal/ui/shared"
)

// Dependencies contains the dependencies required to render the register page.
type Dependencies struct {
	Passwordless       bool
	EnableVerification bool

	Endpoint string

	Layout shared.LayoutFunc

	Logger *slog.Logger
}
