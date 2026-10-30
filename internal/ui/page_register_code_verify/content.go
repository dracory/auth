package page_register_code_verify

import (
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
func RegisterCodeVerifyScripts(urlApiRegisterCodeVerify, urlOnSuccess string) string {
	return `
		var urlApiRegisterCodeVerify = "` + urlApiRegisterCodeVerify + `";
		var urlOnSuccess = "` + urlOnSuccess + `";

		function registerCodeFormValidate() {
			var verificationCode = inputVal('verification_code');

			if (verificationCode === '') {
				return loginFormRaiseError('Code is required');
			}

			spinner(true);

			// The nonce travels via sessionStorage (set by the register
			// page), not the URL — keeps it out of history, logs, Referer.
			var nonce = sessionStorage.getItem('auth_register_nonce') || '';

			postForm(urlApiRegisterCodeVerify, {"verification_code": verificationCode, "nonce": nonce}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return loginFormRaiseError(response.message);
				}

				authStore('AuthToken', response.data.token);
				authStore('AuthUser', response.data.user);
				sessionStorage.removeItem('auth_register_nonce');
				loginFormRaiseSuccess('Verification successful');
				setTimeout(function () {
					authTo(urlOnSuccess);
				}, 2000);
			}).catch(function (error) {
				console.log(error);
				spinner(false);
				return loginFormRaiseError('There was an error. Try again later!');
			});
		}
		onReady(function () {
			var el = document.querySelector('input[name=verification_code]');
			if (el) { el.focus(); }
		});
	`
}
