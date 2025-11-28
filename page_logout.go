package auth

import (
	"net/http"

	page_logout "github.com/dracory/auth/internal/ui/page_logout"
)

func (a authImplementation) pageLogout(w http.ResponseWriter, r *http.Request) {
	logger := a.GetLogger()
	urlApiLogout := a.LinkApiLogout()
	logger.Debug("logout page initialized",
		"api_logout_url", urlApiLogout,
	)

	deps := page_logout.Dependencies{
		Endpoint: a.endpoint,
		Layout:   a.funcLayout,
		Logger:   logger,
	}

	page_logout.PageLogout(deps, w, r)
}
