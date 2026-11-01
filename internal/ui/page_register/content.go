package page_register

import (
	_ "embed"
	"strings"

	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/hb"
)

// RegisterPasswordlessContent builds the HTML for the passwordless registration page.
func RegisterPasswordlessContent(urlLogin string) string {
	emailInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("email").
		Placeholder("you@example.com").
		Attr("autocomplete", "username").
		Attr("required", "required")

	card := shared.Card(
		shared.PageHeader("Create your account", "A few details and you're in."),
		shared.AlertGroup(),
		shared.NameRow(true),
		shared.FieldGroup("Email", emailInput),
		shared.PrimaryButton("buttonRegister", "bi-person-plus", "Create account", "registerFormValidate()"),
	)

	return hb.NewDiv().Children([]hb.TagInterface{
		card,
		shared.FooterText("Already have an account?", urlLogin, "Login"),
	}).ToHTML()
}

// RegisterPasswordlessScripts builds the JS for the passwordless registration page.
//
//go:embed register_passwordless.js
var registerPasswordlessJS string

func RegisterPasswordlessScripts(urlApiRegister, urlOnSuccess string) string {
	s := registerPasswordlessJS
	s = strings.ReplaceAll(s, "__URL_API_REGISTER__", urlApiRegister)
	s = strings.ReplaceAll(s, "__URL_ON_SUCCESS__", urlOnSuccess)
	return s
}

// RegisterUsernameAndPasswordContent builds the HTML for the username/password registration page.
func RegisterUsernameAndPasswordContent(urlLogin, urlPasswordRestore string) string {
	emailInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("email").
		Placeholder("you@example.com").
		Attr("autocomplete", "username").
		Attr("required", "required")
	passwordInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("password").
		Type(hb.TYPE_PASSWORD).
		Placeholder("Min 8 characters").
		Attr("autocomplete", "new-password").
		Attr("required", "required")

	card := shared.Card(
		shared.PageHeader("Create your account", "A few details and you're in."),
		shared.AlertGroup(),
		shared.NameRow(true),
		shared.FieldGroup("Email", emailInput),
		shared.FieldGroup("Password", passwordInput),
		shared.PrimaryButton("buttonRegister", "bi-person-plus", "Create account", "registerFormValidate()"),
		hb.NewDiv().Class("text-center small mt-3").Child(
			hb.NewHyperlink().Href(urlPasswordRestore).Text("Forgot password?")),
	)

	return hb.NewDiv().Children([]hb.TagInterface{
		card,
		shared.FooterText("Already have an account?", urlLogin, "Login"),
	}).ToHTML()
}

// RegisterUsernameAndPasswordScripts builds the JS for the username/password registration page.
//
//go:embed register_username_password.js
var registerUsernamePasswordJS string

func RegisterUsernameAndPasswordScripts(urlApiRegister, urlOnSuccess string) string {
	s := registerUsernamePasswordJS
	s = strings.ReplaceAll(s, "__URL_API_REGISTER__", urlApiRegister)
	s = strings.ReplaceAll(s, "__URL_ON_SUCCESS__", urlOnSuccess)
	return s
}

// RegisterAuthKnightContent builds the HTML for the AuthKnight registration page.
// The email field is read-only (pre-populated from the verified AuthKnight email).
// akKey is the temporary key that maps to the verified email in the temp key store.
func RegisterAuthKnightContent(urlLogin string, verifiedEmail string, akKey string) string {
	emailGroup := hb.NewDiv().Class("mb-3").Children([]hb.TagInterface{
		hb.NewLabel().Class("form-label small fw-bold text-uppercase text-secondary").Text("Email (verified)"),
		hb.NewDiv().Class("form-control-plaintext text-muted").Text(verifiedEmail),
		hb.NewInput().Type("hidden").Name("email").Value(verifiedEmail),
	})

	akKeyInput := hb.NewInput().Type("hidden").Name("ak_key").Value(akKey)

	card := shared.Card(
		shared.PageHeader("Create your account", "Confirm your details to finish registration."),
		shared.AlertGroup(),
		shared.NameRow(true),
		emailGroup,
		akKeyInput,
		shared.PrimaryButton("buttonRegister", "bi-person-plus", "Create account", "registerFormValidate()"),
	)

	return hb.NewDiv().Children([]hb.TagInterface{
		card,
		shared.FooterText("Already have an account?", urlLogin, "Login"),
	}).ToHTML()
}

// RegisterAuthKnightScripts builds the JS for the AuthKnight registration page.
//
//go:embed register_authknight.js
var registerAuthKnightJS string

func RegisterAuthKnightScripts(urlApiRegister, urlOnSuccess string) string {
	s := registerAuthKnightJS
	s = strings.ReplaceAll(s, "__URL_API_REGISTER__", urlApiRegister)
	s = strings.ReplaceAll(s, "__URL_ON_SUCCESS__", urlOnSuccess)
	return s
}

// RegisterCompletionContent builds the HTML for the authenticated
// registration-completion page — shown when the requester already has a
// session (e.g. created implicitly by an OTP/magic-link login). The
// email was already verified so it is rendered read-only and is never
// read back from the form. verifiedEmail may be empty when the host
// provides no FuncUserEmailGet hook — a generic label is shown then.
func RegisterCompletionContent(urlLogin string, verifiedEmail string) string {
	emailDisplay := verifiedEmail
	if emailDisplay == "" {
		emailDisplay = "Your verified email"
	}

	emailGroup := hb.NewDiv().Class("mb-3").Children([]hb.TagInterface{
		hb.NewLabel().Class("form-label small fw-bold text-uppercase text-secondary").Text("Email (verified)"),
		hb.NewDiv().Class("form-control-plaintext text-muted").Text(emailDisplay),
	})

	card := shared.Card(
		shared.PageHeader("Complete your registration", "Confirm your details to finish setting up your account."),
		shared.AlertGroup(),
		shared.NameRow(true),
		emailGroup,
		shared.PrimaryButton("buttonRegister", "bi-check-circle", "Save & continue", "registerFormValidate()"),
	)

	return hb.NewDiv().Children([]hb.TagInterface{
		card,
		shared.FooterText("Already have an account?", urlLogin, "Login"),
	}).ToHTML()
}

// RegisterCompletionScripts builds the JS for the completion page.
//
//go:embed register_completion.js
var registerCompletionJS string

func RegisterCompletionScripts(urlApiRegister, urlOnSuccess string) string {
	s := registerCompletionJS
	s = strings.ReplaceAll(s, "__URL_API_REGISTER__", urlApiRegister)
	s = strings.ReplaceAll(s, "__URL_ON_SUCCESS__", urlOnSuccess)
	return s
}
