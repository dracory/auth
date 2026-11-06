package page_register_code_verify

import (
	_ "embed"
	"strings"

	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/auth/types"
	"github.com/dracory/hb"
)

// RegisterCodeVerifyContent builds the HTML for the registration code verification page.
// Zero-value options render the built-in copy.
func RegisterCodeVerifyContent(urlBack string, opts types.CodeVerifyPageOptions, legal types.LegalLinks) string {
	verificationCodeInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("verification_code").
		Placeholder("Enter verification code").
		Attr("required", "required")

	card := shared.Card(
		shared.PageHeader(
			shared.OrDefault(opts.Title, "Check your email"),
			shared.OrDefault(opts.Subtitle, "We sent a verification code to your email to confirm your new account.")),
		shared.AlertGroup(),
		shared.FieldGroup(shared.OrDefault(opts.CodeLabel, "Verification code"), verificationCodeInput),
		shared.PrimaryButton("buttonVerify", "bi-check2-circle", shared.OrDefault(opts.SubmitLabel, "Verify registration"), "registerCodeFormValidate()"),
		hb.NewHTML(shared.LegalLine(legal)),
	)

	return hb.NewDiv().Children([]hb.TagInterface{
		card,
		shared.FooterText(
			shared.OrDefault(opts.ResendPrompt, "Didn't get a code?"),
			urlBack,
			shared.OrDefault(opts.ResendLabel, "Resend")),
	}).ToHTML()
}

// RegisterCodeVerifyScripts builds the JS for the registration code verification page.
//
//go:embed register_code_verify.js
var registerCodeVerifyJS string

func RegisterCodeVerifyScripts(urlApiRegisterCodeVerify, urlOnSuccess string) string {
	s := registerCodeVerifyJS
	s = strings.ReplaceAll(s, "__URL_API_REGISTER_CODE_VERIFY__", urlApiRegisterCodeVerify)
	s = strings.ReplaceAll(s, "__URL_ON_SUCCESS__", urlOnSuccess)
	return s
}
