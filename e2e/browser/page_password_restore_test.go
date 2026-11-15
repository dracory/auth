package browser

import (
	"testing"
)

// TestPasswordRestorePageRenders verifies the password restore page
// carries the full expected markup: optional name row, required email,
// the send-reset-link button, and both footer links.
func TestPasswordRestorePageRenders(t *testing.T) {
	ctx := newTab(t)
	navigate(t, ctx, pageURL("password-restore"))

	mustExist(t, ctx,
		// Page chrome
		`.auth-box`,
		`.auth-panel.card.shadow-sm`,
		`.alert-group`,
		// Name row exists but is NOT required on restore
		`.row.g-3 .col-6 input[name=first_name].form-control.form-control-lg`,
		`.row.g-3 .col-6 input[name=last_name].form-control.form-control-lg`,
		// Required email field
		`input[name=email].form-control.form-control-lg[required][autocomplete="username"]`,
		// Submit button
		`button.ButtonContinue.btn.btn-primary.btn-lg.w-100.fw-bold`,
		`button.ButtonContinue i.bi.bi-envelope`,
		// Footer links: login + register
		`.text-muted.small.text-center.mt-4 a.fw-bold[href*="login"]`,
		`.text-muted.small.text-center.mt-4 a.fw-bold[href*="register"]`,
	)

	// Restore sends names but does not gate on them — no required attr.
	mustNotExist(t, ctx,
		`input[name=first_name][required]`,
		`input[name=last_name][required]`,
	)

	mustContain(t, ctx,
		"Forgot your password?",
		"Enter your email and we'll send you a reset link.",
		"FIRST NAME",
		"LAST NAME",
		"EMAIL",
		"Send reset link",
		"Remembered it?",
		"Login",
		"Don't have an account?",
		"Register",
	)
}
