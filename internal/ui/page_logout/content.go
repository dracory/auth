package page_logout

import (
	_ "embed"
	"strings"

	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/auth/types"
	"github.com/dracory/hb"
)

// LogoutContent builds the HTML for the logout page.
// Zero-value options render the built-in copy.
func LogoutContent(opts types.LogoutPageOptions) string {
	card := shared.Card(
		shared.PageHeader(
			shared.OrDefault(opts.Title, "Sign out"),
			shared.OrDefault(opts.Subtitle, "Are you sure you want to sign out of your account?")),
		shared.AlertGroup(),
		shared.PrimaryButton("ButtonLogout", "bi-box-arrow-right", shared.OrDefault(opts.SubmitLabel, "Logout"), "logoutFormValidate()"),
	)

	return hb.NewDiv().Children([]hb.TagInterface{
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
