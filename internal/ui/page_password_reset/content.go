package page_password_reset

import (
	_ "embed"
	"strings"

	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/auth/types"
	"github.com/dracory/hb"
)

// PasswordResetContent builds the HTML for the password reset page.
// Zero-value options render the built-in copy.
func PasswordResetContent(token, errorMessage, urlPasswordRestore, urlLogin, urlRegister string, enableRegistration bool, opts types.PasswordResetPageOptions) string {
	alertSuccess := hb.NewDiv().Class("alert alert-success").Style("display:none")
	alertDanger := hb.NewDiv().Class("alert alert-danger")
	if errorMessage != "" {
		alertDanger.Text(errorMessage)
	} else {
		alertDanger.Style("display:none")
	}
	alertGroup := hb.NewDiv().Class("alert-group").AddChild(alertSuccess).AddChild(alertDanger)

	header := shared.PageHeader(
		shared.OrDefault(opts.Title, "Set a new password"),
		shared.OrDefault(opts.Subtitle, "Choose a strong password for your account."))

	var cardBody hb.TagInterface
	if errorMessage == "" {
		tokenInput := hb.NewInput().Name("token").Value(token)
		passwordInput := hb.NewInput().Class("form-control form-control-lg").Name("password").Type(hb.TYPE_PASSWORD).
			Placeholder(shared.OrDefault(opts.NewPasswordPlaceholder, "Min 8 characters")).Attr("autocomplete", "new-password").Attr("required", "required")
		passwordConfirmInput := hb.NewInput().Class("form-control form-control-lg").Name("password_confirm").Type(hb.TYPE_PASSWORD).
			Placeholder(shared.OrDefault(opts.ConfirmPasswordPlaceholder, "Repeat password")).Attr("autocomplete", "new-password").Attr("required", "required")
		cardBody = shared.Card(
			header,
			alertGroup,
			tokenInput,
			shared.FieldGroup(shared.OrDefault(opts.NewPasswordLabel, "New password"), passwordInput),
			shared.FieldGroup(shared.OrDefault(opts.ConfirmPasswordLabel, "Confirm new password"), passwordConfirmInput),
			shared.PrimaryButton("ButtonContinue", "bi-shield-check", shared.OrDefault(opts.SubmitLabel, "Reset password"), "resetFormValidate()"),
		)
	} else {
		cardBody = shared.Card(
			header,
			alertGroup,
			hb.NewParagraph().Text(shared.OrDefault(opts.ErrorText, "Sorry, there was an error processing your request. Please select one of the following options:")),
			hb.NewParagraph().AddChild(hb.NewHyperlink().Href(urlPasswordRestore).Text(shared.OrDefault(opts.ErrorRestoreLabel, "request a reset of your password"))),
			hb.NewParagraph().AddChild(hb.NewHyperlink().Href(urlLogin).Text(shared.OrDefault(opts.ErrorLoginLabel, "login to the system"))),
			hb.NewParagraph().AddChild(hb.NewHyperlink().Href(urlRegister).Text(shared.OrDefault(opts.ErrorRegisterLabel, "create a new account"))),
		)
	}

	container := hb.NewDiv().Children([]hb.TagInterface{
		cardBody,
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

// PasswordResetScripts builds the JS for the password reset page.
//
//go:embed password_reset.js
var passwordResetJS string

func PasswordResetScripts(urlApiPasswordReset, urlOnSuccess string) string {
	s := passwordResetJS
	s = strings.ReplaceAll(s, "__URL_API_PASSWORD_RESET__", urlApiPasswordReset)
	s = strings.ReplaceAll(s, "__URL_ON_SUCCESS__", urlOnSuccess)
	return s
}
