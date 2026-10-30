package page_login_code_verify

import (
	_ "embed"
	"strings"

	"fmt"

	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/hb"
)

// LoginCodeVerifyContent builds the HTML for the login code verification page.
func LoginCodeVerifyContent(urlBack string) string {
	// Segmented input — one box per code character, synced into a hidden
	// field so the verify JS keeps posting `verification_code` unchanged.
	// Login codes are 8 uppercase consonants (see utils.LoginCodeLength),
	// so the boxes accept letters, not just digits.
	verificationCodeInput := hb.NewInput().
		Name("verification_code").
		Type("hidden")
	digitInputs := hb.NewDiv().Class("row g-2 mb-4")
	for i := 0; i < 8; i++ {
		digit := hb.NewInput().
			Class("otp-digit form-control form-control-lg text-center fw-bold").
			Attr("maxlength", "1").
			Attr("inputmode", "text").
			Attr("data-index", fmt.Sprint(i))
		if i == 0 {
			digit.Attr("autocomplete", "one-time-code")
		}
		digitInputs.Child(hb.NewDiv().Class("col").Child(digit))
	}

	card := shared.Card(
		shared.AlertGroup(),
		digitInputs,
		verificationCodeInput,
		shared.PrimaryButton("ButtonLogin", "bi-check2-circle", "Verify code", "loginFormValidate()"),
	)

	return hb.NewDiv().Children([]hb.TagInterface{
		shared.PageHeader("Check your email", "We sent you a code by email. Enter it below to sign in."),
		card,
		shared.FooterText("Didn't get a code?", urlBack, "Resend"),
	}).ToHTML()
}

// LoginCodeVerifyScripts builds the JS for the login code verification page.
//
//go:embed login_code_verify.js
var loginCodeVerifyJS string

func LoginCodeVerifyScripts(urlApiLoginCodeVerify, urlOnSuccess string) string {
	s := loginCodeVerifyJS
	s = strings.ReplaceAll(s, "__URL_API_LOGIN_CODE_VERIFY__", urlApiLoginCodeVerify)
	s = strings.ReplaceAll(s, "__URL_ON_SUCCESS__", urlOnSuccess)
	return s
}
