package page_password_restore

import (
	_ "embed"
	"strings"

	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/hb"
)

// PasswordRestoreContent builds the HTML for the password restore page.
func PasswordRestoreContent(enableRegistration bool, urlLogin, urlRegister string) string {
	emailInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("email").
		Placeholder("you@example.com").
		Attr("autocomplete", "username").
		Attr("required", "required")

	card := shared.Card(
		shared.AlertGroup(),
		shared.NameRow(false),
		shared.FieldGroup("Email", emailInput),
		shared.PrimaryButton("ButtonContinue", "bi-envelope", "Send reset link", "passwordRestoreFormValidate()"),
	)

	container := hb.NewDiv().Children([]hb.TagInterface{
		shared.PageHeader("Forgot your password?", "Enter your email and we'll send you a reset link."),
		card,
		shared.FooterText("Remembered it?", urlLogin, "Login"),
	})

	if enableRegistration {
		container.Child(shared.FooterText("Don't have an account?", urlRegister, "Register"))
	}

	return container.ToHTML()
}

// PasswordRestoreScripts builds the JS for the password restore page.
//
//go:embed password_restore.js
var passwordRestoreJS string

func PasswordRestoreScripts(urlApiPasswordRestore, urlOnSuccess string) string {
	s := passwordRestoreJS
	s = strings.ReplaceAll(s, "__URL_API_PASSWORD_RESTORE__", urlApiPasswordRestore)
	s = strings.ReplaceAll(s, "__URL_ON_SUCCESS__", urlOnSuccess)
	return s
}
