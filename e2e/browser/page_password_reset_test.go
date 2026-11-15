package browser

import (
	"testing"
)

// TestPasswordResetPageRenders verifies the password reset page renders
// its error card when no ?t= token is supplied — the password fields
// must NOT exist in that state.
func TestPasswordResetPageRenders(t *testing.T) {
	ctx := newTab(t)
	navigate(t, ctx, pageURL("password-reset"))

	mustExist(t, ctx,
		// Page chrome
		`.auth-box`,
		`.auth-panel.card.shadow-sm`,
		`h1.h3.fw-bold`,
		// Error card: visible danger alert + recovery links
		`.alert-group .alert.alert-danger`,
		`a[href*="password-restore"]`,
		`a[href*="login"]`,
		`a[href*="register"]`,
	)

	// No token → no reset form.
	mustNotExist(t, ctx,
		`input[name=password]`,
		`input[name=password_confirm]`,
		`input[name=token]`,
		`button.ButtonContinue`,
	)

	mustContain(t, ctx,
		"Set a new password",
		"request a reset of your password",
		"login to the system",
		"create a new account",
	)
}
