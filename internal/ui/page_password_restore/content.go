package page_password_restore

import (
	_ "embed"
	"strings"

	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/auth/types"
	"github.com/dracory/hb"
)

// PasswordRestoreContent builds the HTML for the password restore page.
// Zero-value options render the built-in copy.
func PasswordRestoreContent(enableRegistration bool, urlLogin, urlRegister string, opts types.PasswordRestorePageOptions) string {
	emailInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("email").
		Placeholder(shared.OrDefault(opts.EmailPlaceholder, "you@example.com")).
		Attr("autocomplete", "username").
		Attr("required", "required")

	card := shared.Card(
		shared.PageHeader(
			shared.OrDefault(opts.Title, "Forgot your password?"),
			shared.OrDefault(opts.Subtitle, "Enter your email and we'll send you a reset link.")),
		shared.AlertGroup(),
		shared.NameRow(false,
			opts.FirstNameLabel, opts.FirstNamePlaceholder,
			opts.LastNameLabel, opts.LastNamePlaceholder),
		shared.FieldGroup(shared.OrDefault(opts.EmailLabel, "Email"), emailInput),
		shared.PrimaryButton("ButtonContinue", "bi-envelope", shared.OrDefault(opts.SubmitLabel, "Send reset link"), "passwordRestoreFormValidate()"),
	)

	container := hb.NewDiv().Children([]hb.TagInterface{
		card,
		shared.FooterText(
			shared.OrDefault(opts.FooterPrompt, "Remembered it?"),
			urlLogin,
			shared.OrDefault(opts.FooterLinkLabel, "Login")),
	})

	if enableRegistration {
		container.Child(shared.FooterText(
			shared.OrDefault(opts.RegisterPrompt, "Don't have an account?"),
			urlRegister,
			shared.OrDefault(opts.RegisterLinkLabel, "Register")))
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
