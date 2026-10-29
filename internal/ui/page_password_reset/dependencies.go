package page_password_reset

import (
	"log/slog"

	"github.com/dracory/auth/internal/ui/shared"
)

// Dependencies contains the dependencies required to render the password reset page.
type Dependencies struct {
	Endpoint           string
	EnableRegistration bool
	Token              string
	ErrorMessage       string

	Layout shared.LayoutFunc

	Logger *slog.Logger
}
