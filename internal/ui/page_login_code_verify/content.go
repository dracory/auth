package page_login_code_verify

import (
	"fmt"

	"github.com/dracory/hb"
)

// LoginCodeVerifyContent builds the HTML for the login code verification page.
func LoginCodeVerifyContent(urlBack string) string {
	// Elements for the form
	alertSuccess := hb.NewDiv().Class("alert alert-success").Style("display:none")
	alertDanger := hb.NewDiv().Class("alert alert-danger").Style("display:none")
	alertGroup := hb.NewDiv().Class("alert-group").AddChild(alertSuccess).AddChild(alertDanger)

	header := hb.NewHeading5().Text("Login Code Verification").Style("margin:0px;")
	infoParagraph := hb.NewParagraph().Class("text-info").Text("We sent you a 6-digit login code to your email. Please check your mailbox")
	verificationCodeLabel := hb.NewLabel().Text("Verification code").Style("display:block;text-align:center;font-weight:700;letter-spacing:.06em")

	// Segmented 6-digit input — one box per digit, synced into a hidden
	// field so the verify JS keeps posting `verification_code` unchanged.
	verificationCodeInput := hb.NewInput().
		Name("verification_code").
		Type("hidden")
	digitStyle := "width:48px;height:56px;text-align:center;font-size:1.5rem;font-weight:700;"
	digitInputs := hb.NewDiv().
		Class("otp-inputs").
		Style("display:flex;justify-content:center;gap:10px;margin:14px 0 4px")
	for i := 0; i < 6; i++ {
		digitInputs.Child(hb.NewInput().
			Class("otp-digit form-control").
			Attr("maxlength", "1").
			Attr("inputmode", "numeric").
			Attr("autocomplete", "one-time-code").
			Attr("data-index", fmt.Sprint(i)).
			Style(digitStyle))
	}
	verificationCodeFormGroup := hb.NewDiv().Class("form-group mt-3").
		Child(verificationCodeLabel).
		Child(digitInputs).
		Child(verificationCodeInput)
	buttonLogin := hb.NewButton().Class("ButtonLogin btn btn-lg btn-success btn-block w-100 text-white").Children([]hb.TagInterface{
		hb.NewI().Class("bi bi-send").Style("margin-right:8px;margin-top:-2px;"),
		hb.NewSpan().Text("Login"),
		hb.NewDiv().Class("ImgLoading spinner-border spinner-border-sm text-light").Style("display:none;margin-left:10px;"),
	}).OnClick("loginFormValidate()")
	buttonLoginFormGroup := hb.NewDiv().Class("form-group mt-3 mb-3").AddChild(buttonLogin)
	buttonBack := hb.NewHyperlink().Class("btn btn-info text-white float-start").Children([]hb.TagInterface{
		hb.NewI().Class("bi bi-chevron-left").Style("margin-right:8px;margin-top:-2px;"),
		hb.NewSpan().Text("Resend code"),
	}).Href(urlBack)

	// Add elements in a card
	cardHeader := hb.NewDiv().Class("card-header").Child(header)
	cardBody := hb.NewDiv().Class("card-body").Children([]hb.TagInterface{
		alertGroup,
		infoParagraph,
		verificationCodeFormGroup,
		buttonLoginFormGroup,
	})
	cardFooter := hb.NewDiv().Class("card-footer").Children([]hb.TagInterface{
		buttonBack,
	})

	card := hb.NewDiv().Class("card card-default").Style("margin:0 auto;max-width: 360px;").Children([]hb.TagInterface{
		cardHeader,
		cardBody,
		cardFooter,
	})

	container := hb.NewDiv().Class("container").Child(card)

	return container.ToHTML()
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
