package browser

import (
	"testing"
)

// TestLoginPageRenders verifies the server-rendered login page carries
// the full expected markup: card chrome, both fields with their labels,
// the submit button, the forgot-password link, and the register footer.
func TestLoginPageRenders(t *testing.T) {
	ctx := newTab(t)
	navigate(t, ctx, pageURL("login"))

	mustExist(t, ctx,
		// Page chrome — shared auth box + card panel
		`.auth-box`,
		`.auth-panel.card.shadow-sm`,
		`.auth-panel-body.card-body.p-4`,
		// Header
		`h1.h3.fw-bold`,
		`p.text-muted`,
		// Alert pair the page scripts address (hidden until used)
		`.alert-group`,
		`.alert.alert-success`,
		`.alert.alert-danger`,
		// Email field: uppercase small label + lg required input
		`label.form-label.small.fw-bold.text-uppercase.text-secondary`,
		`input[name=email].form-control.form-control-lg[required][autocomplete="username"][placeholder="you@example.com"]`,
		// Password field
		`input[name=password][type=password].form-control.form-control-lg[required][autocomplete="current-password"]`,
		// Submit button: marker class + bootstrap styling + spinner
		`button.ButtonLogin.btn.btn-primary.btn-lg.w-100.fw-bold`,
		`button.ButtonLogin i.bi.bi-box-arrow-in-right`,
		`button.ButtonLogin .ImgLoading.spinner-border`,
		// Forgot-password link inside the card
		`a[href*="password-restore"]`,
		// Footer register link (registration is enabled)
		`.text-muted.small.text-center.mt-4 a.fw-bold[href*="register"]`,
	)

	mustContain(t, ctx,
		"Login to your account",
		"Enter your email and password to sign in.",
		"EMAIL",
		"PASSWORD",
		"Sign in",
		"Forgot password?",
		"Don't have an account?",
		"Register",
	)
}
