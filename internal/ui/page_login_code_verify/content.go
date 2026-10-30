package page_login_code_verify

import (
	"fmt"

	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/hb"
)

// LoginCodeVerifyContent builds the HTML for the login code verification page.
func LoginCodeVerifyContent(urlBack string) string {
	// Segmented 6-digit input — one box per digit, synced into a hidden
	// field so the verify JS keeps posting `verification_code` unchanged.
	verificationCodeInput := hb.NewInput().
		Name("verification_code").
		Type("hidden")
	digitInputs := hb.NewDiv().Class("row g-2 mb-4")
	for i := 0; i < 6; i++ {
		digit := hb.NewInput().
			Class("otp-digit form-control form-control-lg text-center fw-bold").
			Attr("maxlength", "1").
			Attr("inputmode", "numeric").
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
		shared.PageHeader("Check your email", "We sent a 6-digit code to your email. Enter it below to sign in."),
		card,
		shared.FooterText("Didn't get a code?", urlBack, "Resend"),
	}).ToHTML()
}

// LoginCodeVerifyScripts builds the JS for the login code verification page.
func LoginCodeVerifyScripts(urlApiLoginCodeVerify, urlOnSuccess string) string {
	return `
		var urlApiLoginCodeVerify = "` + urlApiLoginCodeVerify + `";
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
		function loginFormValidate() {
			var verificationCode = $.trim($('input[name=verification_code]').val());

			if (verificationCode === '') {
				return loginFormRaiseError('Code is required');
			}

			$('.ButtonLogin .ImgLoading').show();

			// The nonce travels via sessionStorage (set by the login page),
			// not the URL — keeps it out of history, logs, and Referer.
			var nonce = sessionStorage.getItem('auth_login_nonce') || '';
			var email = sessionStorage.getItem('auth_login_email') || '';

			var data = {"verification_code": verificationCode, "nonce": nonce, "email": email};

			$.post(urlApiLoginCodeVerify, data).then(function (response) {
				$('.ButtonLogin .ImgLoading').hide();

				if (response.status !== "success") {
					return loginFormRaiseError(response.message);
				}

				$$.setAuthToken(response.data.token);
				$$.setAuthUser(response.data.user);
				sessionStorage.removeItem('auth_login_nonce');
				sessionStorage.removeItem('auth_login_email');
				loginFormRaiseSuccess('Verification successful');
				$('div.alert-danger').html('').hide();
				setTimeout(function () {
					$$.to(urlOnSuccess);
				}, 2000);
				return;
			}).fail(function (error) {
				console.log(error);
				$('.ButtonLogin .ImgLoading').hide();
				return loginFormRaiseError('There was an error. Try again later!');
			});
		}
		$(function () {
			var $digits = $('.otp-digit');
			var $hidden = $('input[name=verification_code]');

			function syncDigits() {
				var code = '';
				$digits.each(function () { code += $(this).val(); });
				$hidden.val(code);
				return code;
			}

			$digits.on('input', function () {
				var v = $(this).val().replace(/[^0-9]/g, '');
				$(this).val(v.slice(-1));
				syncDigits();
				if (v && this.nextElementSibling) {
					$(this).nextAll('.otp-digit').first().focus();
				}
				if (syncDigits().length === $digits.length) {
					loginFormValidate();
				}
			});
			$digits.on('keydown', function (e) {
				if (e.key === 'Backspace' && !$(this).val()) {
					$(this).prevAll('.otp-digit').first().focus();
				}
			});
			$digits.on('paste', function (e) {
				e.preventDefault();
				var pasted = (e.originalEvent.clipboardData || window.clipboardData)
					.getData('text').replace(/[^0-9]/g, '').slice(0, $digits.length);
				$digits.each(function (i) { $(this).val(pasted[i] || ''); });
				syncDigits();
				$digits.eq(Math.min(pasted.length, $digits.length - 1)).focus();
				if (pasted.length === $digits.length) {
					loginFormValidate();
				}
			});
			$digits.first().focus();
		});
	`
}
