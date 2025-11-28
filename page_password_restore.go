package auth

import (
	"net/http"

	page_password_restore "github.com/dracory/auth/internal/ui/page_password_restore"
)

func (a authImplementation) pagePasswordRestore(w http.ResponseWriter, r *http.Request) {
	deps := page_password_restore.Dependencies{
		EnableRegistration: a.enableRegistration,
		Endpoint:           a.endpoint,
		Layout:             a.funcLayout,
		Logger:             a.GetLogger(),
	}

	page_password_restore.PagePasswordRestore(deps, w, r)
}
