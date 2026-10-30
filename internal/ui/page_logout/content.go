package page_logout

import (
	_ "embed"
	"strings"

	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/hb"
)

// LogoutContent builds the HTML for the logout page.
func LogoutContent() string {
	card := shared.Card(
		shared.AlertGroup(),
		shared.PrimaryButton("ButtonLogout", "bi-box-arrow-right", "Logout", "logoutFormValidate()"),
	)

	return hb.NewDiv().Children([]hb.TagInterface{
		shared.PageHeader("Sign out", "Are you sure you want to sign out of your account?"),
		card,
	}).ToHTML()
}

// LogoutScripts builds the JS for the logout page.
//
//go:embed logout.js
var logoutJS string

func LogoutScripts(urlApiLogout, urlOnSuccess string) string {
	s := logoutJS
	s = strings.ReplaceAll(s, "__URL_API_LOGOUT__", urlApiLogout)
	s = strings.ReplaceAll(s, "__URL_ON_SUCCESS__", urlOnSuccess)
	return s
}
