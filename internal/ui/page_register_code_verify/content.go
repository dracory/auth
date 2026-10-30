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
		/**
		 * Raises an error message
		 * @param  {String} error
		 * @returns  {Boolean}
		 */
		function loginFormRaiseError(error) {
			$('div.alert-success').html('').hide();
			$('div.alert-danger').html(error).show();
			setTimeout(function () {
				$('div.alert-danger').html('').hide();
			}, 10000);
			return false;
		}

		function loginFormRaiseSuccess(success) {
			$('div.alert-danger').html('').hide();
			$('div.alert-success').html(success).show();
			setTimeout(function () {
				$('div.alert-success').html('').hide();
			}, 10000);
			return false;
		}

		/**
		 * Validate Login Form
		 * @returns  {Boolean}
		 */
		function registerCodeFormValidate() {
			var verificationCode = $.trim($('input[name=verification_code]').val());

			if (verificationCode === '') {
				return loginFormRaiseError('Code is required');
			}

			$('.buttonLogin .imgLoading').show();

			// The nonce travels via sessionStorage (set by the register
			// page), not the URL — keeps it out of history, logs, Referer.
			var nonce = sessionStorage.getItem('auth_register_nonce') || '';

			var data = {"verification_code": verificationCode, "nonce": nonce};

			$.post(urlApiRegisterCodeVerify, data).then(function (response) {
				$('.buttonLogin .imgLoading').hide();

				if (response.status !== "success") {
					return loginFormRaiseError(response.message);
				}

				$$.setAuthToken(response.data.token);
				$$.setAuthUser(response.data.user);
				sessionStorage.removeItem('auth_register_nonce');
				loginFormRaiseSuccess('Verification successful');
				$('div.alert-danger').html('').hide();
				setTimeout(function () {
					$$.to(urlOnSuccess);
				}, 2000);
				return;
			}).fail(function (error) {
				console.log(error);
				$('.buttonLogin .imgLoading').hide();
				return loginFormRaiseError('There was an error. Try again later!');
			});
		}
		$(function () {
			$("#email").focus();
		});
	`
}
