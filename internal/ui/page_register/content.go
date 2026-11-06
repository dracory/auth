package page_register

import (
	_ "embed"
	"strings"

	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/auth/types"
	"github.com/dracory/hb"
)

// nameRow renders the shared first/last name inputs using the register
// page options (required gates the HTML5 attribute per variant).
func nameRow(required bool, opts types.RegisterPageOptions) hb.TagInterface {
	return shared.NameRow(required,
		opts.FirstNameLabel, opts.FirstNamePlaceholder,
		opts.LastNameLabel, opts.LastNamePlaceholder)
}

// RegisterOtpContent builds the HTML for the otp registration page.
// Zero-value options render the built-in copy.
func RegisterOtpContent(urlLogin string, opts types.RegisterPageOptions, legal types.LegalLinks) string {
	emailInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("email").
		Placeholder(shared.OrDefault(opts.EmailPlaceholder, "you@example.com")).
		Attr("autocomplete", "username").
		Attr("required", "required")

	card := shared.Card(
		shared.PageHeader(
			shared.OrDefault(opts.Title, "Create your account"),
			shared.OrDefault(opts.Subtitle, "A few details and you're in.")),
		shared.AlertGroup(),
		nameRow(true, opts),
		shared.FieldGroup(shared.OrDefault(opts.EmailLabel, "Email"), emailInput),
		shared.PrimaryButton("buttonRegister", "bi-person-plus", shared.OrDefault(opts.SubmitLabel, "Create account"), "registerFormValidate()"),
		hb.NewHTML(shared.LegalLine(legal)),
	)

	return hb.NewDiv().Children([]hb.TagInterface{
		card,
		shared.FooterText(
			shared.OrDefault(opts.FooterPrompt, "Already have an account?"),
			urlLogin,
			shared.OrDefault(opts.FooterLinkLabel, "Login")),
	}).ToHTML()
}

// RegisterOtpScripts builds the JS for the otp registration page.
//
//go:embed register_otp.js
var registerOtpJS string

func RegisterOtpScripts(urlApiRegister, urlOnSuccess string) string {
	s := registerOtpJS
	s = strings.ReplaceAll(s, "__URL_API_REGISTER__", urlApiRegister)
	s = strings.ReplaceAll(s, "__URL_ON_SUCCESS__", urlOnSuccess)
	return s
}

// RegisterUsernameAndPasswordContent builds the HTML for the username/password registration page.
// Zero-value options render the built-in copy.
func RegisterUsernameAndPasswordContent(urlLogin, urlPasswordRestore string, opts types.RegisterPageOptions, legal types.LegalLinks) string {
	emailInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("email").
		Placeholder(shared.OrDefault(opts.EmailPlaceholder, "you@example.com")).
		Attr("autocomplete", "username").
		Attr("required", "required")
	passwordInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("password").
		Type(hb.TYPE_PASSWORD).
		Placeholder(shared.OrDefault(opts.PasswordPlaceholder, "Min 8 characters")).
		Attr("autocomplete", "new-password").
		Attr("required", "required")

	card := shared.Card(
		shared.PageHeader(
			shared.OrDefault(opts.Title, "Create your account"),
			shared.OrDefault(opts.Subtitle, "A few details and you're in.")),
		shared.AlertGroup(),
		nameRow(true, opts),
		shared.FieldGroup(shared.OrDefault(opts.EmailLabel, "Email"), emailInput),
		shared.FieldGroup(shared.OrDefault(opts.PasswordLabel, "Password"), passwordInput),
		shared.PrimaryButton("buttonRegister", "bi-person-plus", shared.OrDefault(opts.SubmitLabel, "Create account"), "registerFormValidate()"),
		hb.NewDiv().Class("text-center small mt-3").Child(
			hb.NewHyperlink().Href(urlPasswordRestore).Text(shared.OrDefault(opts.ForgotPasswordLabel, "Forgot password?"))),
		hb.NewHTML(shared.LegalLine(legal)),
	)

	return hb.NewDiv().Children([]hb.TagInterface{
		card,
		shared.FooterText(
			shared.OrDefault(opts.FooterPrompt, "Already have an account?"),
			urlLogin,
			shared.OrDefault(opts.FooterLinkLabel, "Login")),
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

// RegisterVerifiedEmailContent builds the HTML for the verified-email
// registration page — the unified register-first form used by AuthKnight
// OTP, magic-link, and AuthKnight flows. The email field is read-only
// (pre-populated from the verified email); verifiedKey is the temporary
// key that maps to that email in the temp key store.
// Zero-value options render the built-in copy.
func RegisterVerifiedEmailContent(urlLogin string, verifiedEmail string, verifiedKey string, opts types.RegisterPageOptions, legal types.LegalLinks) string {
	emailGroup := hb.NewDiv().Class("mb-3").Children([]hb.TagInterface{
		hb.NewLabel().Class("form-label small fw-bold text-uppercase text-secondary").Text(shared.OrDefault(opts.EmailLabel, "Email (verified)")),
		hb.NewDiv().Class("form-control-plaintext text-muted").Text(verifiedEmail),
		hb.NewInput().Type("hidden").Name("email").Value(verifiedEmail),
	})

	verifiedKeyInput := hb.NewInput().Type("hidden").Name("verified_key").Value(verifiedKey)

	card := shared.Card(
		shared.PageHeader(
			shared.OrDefault(opts.Title, "Create your account"),
			shared.OrDefault(opts.Subtitle, "Confirm your details to finish registration.")),
		shared.AlertGroup(),
		nameRow(true, opts),
		emailGroup,
		verifiedKeyInput,
		shared.PrimaryButton("buttonRegister", "bi-person-plus", shared.OrDefault(opts.SubmitLabel, "Create account"), "registerFormValidate()"),
		hb.NewHTML(shared.LegalLine(legal)),
	)

	return hb.NewDiv().Children([]hb.TagInterface{
		card,
		shared.FooterText(
			shared.OrDefault(opts.FooterPrompt, "Already have an account?"),
			urlLogin,
			shared.OrDefault(opts.FooterLinkLabel, "Login")),
	}).ToHTML()
}

// RegisterVerifiedEmailScripts builds the JS for the verified-email
// registration page.
//
//go:embed register_verified_email.js
var registerVerifiedEmailJS string

func RegisterVerifiedEmailScripts(urlApiRegister, urlOnSuccess string) string {
	s := registerVerifiedEmailJS
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
// Zero-value options render the built-in copy.
func RegisterCompletionContent(urlLogin string, verifiedEmail string, opts types.RegisterPageOptions, legal types.LegalLinks) string {
	emailDisplay := verifiedEmail
	if emailDisplay == "" {
		emailDisplay = "Your verified email"
	}

	emailGroup := hb.NewDiv().Class("mb-3").Children([]hb.TagInterface{
		hb.NewLabel().Class("form-label small fw-bold text-uppercase text-secondary").Text(shared.OrDefault(opts.EmailLabel, "Email (verified)")),
		hb.NewDiv().Class("form-control-plaintext text-muted").Text(emailDisplay),
	})

	card := shared.Card(
		shared.PageHeader(
			shared.OrDefault(opts.Title, "Complete your registration"),
			shared.OrDefault(opts.Subtitle, "Confirm your details to finish setting up your account.")),
		shared.AlertGroup(),
		nameRow(true, opts),
		emailGroup,
		shared.PrimaryButton("buttonRegister", "bi-check-circle", shared.OrDefault(opts.SubmitLabel, "Save & continue"), "registerFormValidate()"),
		hb.NewHTML(shared.LegalLine(legal)),
	)

	return hb.NewDiv().Children([]hb.TagInterface{
		card,
		shared.FooterText(
			shared.OrDefault(opts.FooterPrompt, "Already have an account?"),
			urlLogin,
			shared.OrDefault(opts.FooterLinkLabel, "Login")),
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
