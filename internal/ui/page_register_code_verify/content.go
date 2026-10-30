package page_register_code_verify

import (
	_ "embed"
	"strings"

	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/hb"
)

// RegisterCodeVerifyContent builds the HTML for the registration code verification page.
func RegisterCodeVerifyContent(urlBack string) string {
	verificationCodeInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("verification_code").
		Placeholder("Enter verification code").
		Attr("required", "required")

	card := shared.Card(
		shared.AlertGroup(),
		shared.FieldGroup("Verification code", verificationCodeInput),
		shared.PrimaryButton("buttonVerify", "bi-check2-circle", "Verify registration", "registerCodeFormValidate()"),
	)

	return hb.NewDiv().Children([]hb.TagInterface{
		shared.PageHeader("Check your email", "We sent a verification code to your email to confirm your new account."),
		card,
		shared.FooterText("Didn't get a code?", urlBack, "Resend"),
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
