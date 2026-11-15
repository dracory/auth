package browser

import (
	"testing"
)

// TestRegisterPageRenders verifies the username+password register page
// carries the full expected markup: required name row, email and
// password fields, the create-account button, and the login footer.
func TestRegisterPageRenders(t *testing.T) {
	ctx := newTab(t)
	navigate(t, ctx, pageURL("register"))

	mustExist(t, ctx,
		// Page chrome
		`.auth-box`,
		`.auth-panel.card.shadow-sm`,
		`.alert-group`,
		// Name row: two col-6 fields, both required
		`.row.g-3 .col-6 input[name=first_name].form-control.form-control-lg[required]`,
		`.row.g-3 .col-6 input[name=last_name].form-control.form-control-lg[required]`,
		// Email + password fields
		`input[name=email].form-control.form-control-lg[required][autocomplete="username"]`,
		`input[name=password][type=password].form-control.form-control-lg[required][autocomplete="new-password"]`,
		// Submit button
		`button.buttonRegister.btn.btn-primary.btn-lg.w-100.fw-bold`,
		`button.buttonRegister i.bi.bi-person-plus`,
		`button.buttonRegister .ImgLoading.spinner-border`,
		// Forgot-password link
		`a[href*="password-restore"]`,
		// Footer login link
		`.text-muted.small.text-center.mt-4 a.fw-bold[href*="login"]`,
	)

	mustContain(t, ctx,
		"Create your account",
		"A few details and you're in.",
		"FIRST NAME",
		"LAST NAME",
		"EMAIL",
		"PASSWORD",
		"Create account",
		"Forgot password?",
		"Already have an account?",
		"Login",
	)
}
