package browser

import (
	"testing"
)

// TestLogoutPageRenders verifies the logout confirmation page carries
// the full expected markup: header, alert pair, and the logout button.
func TestLogoutPageRenders(t *testing.T) {
	ctx := newTab(t)
	navigate(t, ctx, pageURL("logout"))

	mustExist(t, ctx,
		// Page chrome
		`.auth-box`,
		`.auth-panel.card.shadow-sm`,
		`.alert-group`,
		// Logout button
		`button.ButtonLogout.btn.btn-primary.btn-lg.w-100.fw-bold`,
		`button.ButtonLogout i.bi.bi-box-arrow-right`,
		`button.ButtonLogout .ImgLoading.spinner-border`,
	)

	mustContain(t, ctx,
		"Sign out",
		"Are you sure you want to sign out of your account?",
		"Logout",
	)
}
