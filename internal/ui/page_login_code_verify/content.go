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

	function alertShow(cls, message) {
		var other = cls === 'div.alert-danger' ? 'div.alert-success' : 'div.alert-danger';
		var o = document.querySelector(other);
		o.innerHTML = ''; o.style.display = 'none';
		var el = document.querySelector(cls);
		el.innerHTML = message; el.style.display = 'block';
		setTimeout(function () { el.innerHTML = ''; el.style.display = 'none'; }, 10000);
		return false;
	}
	function loginFormRaiseError(error) { return alertShow('div.alert-danger', error); }
	function loginFormRaiseSuccess(success) { return alertShow('div.alert-success', success); }
	function registerFormRaiseError(error) { return alertShow('div.alert-danger', error); }
	function registerFormRaiseSuccess(success) { return alertShow('div.alert-success', success); }
	function resetFormRaiseError(error) { return alertShow('div.alert-danger', error); }
	function resetFormRaiseSuccess(success) { return alertShow('div.alert-success', success); }
	function passwordRestoreFormRaiseError(error) { return alertShow('div.alert-danger', error); }
	function passwordRestoreFormRaiseSuccess(success) { return alertShow('div.alert-success', success); }
	function logoutFormRaiseError(error) { return alertShow('div.alert-danger', error); }
	function logoutFormRaiseSuccess(success) { return alertShow('div.alert-success', success); }

	// postForm posts url-encoded form data and parses the JSON response,
	// mirroring the $.post behavior the pages previously relied on.
	function postForm(url, data) {
		return fetch(url, {
			method: 'POST',
			headers: { 'Content-Type': 'application/x-www-form-urlencoded;charset=UTF-8' },
			body: new URLSearchParams(data)
		}).then(function (response) { return response.json(); });
	}
	function inputVal(name) {
		var el = document.querySelector('input[name=' + name + ']');
		return el ? el.value.trim() : '';
	}
	function spinner(show) {
		var el = document.querySelector('.ImgLoading');
		if (el) { el.style.display = show ? 'inline-block' : 'none'; }
	}
	function onReady(fn) {
		if (document.readyState === 'loading') {
			document.addEventListener('DOMContentLoaded', fn);
		} else {
			fn();
		}
	}

		function loginFormValidate() {
			var verificationCode = inputVal('verification_code');

			if (verificationCode === '') {
				return loginFormRaiseError('Code is required');
			}

			spinner(true);

			// The nonce travels via sessionStorage (set by the login page),
			// not the URL — keeps it out of history, logs, and Referer.
			var nonce = sessionStorage.getItem('auth_login_nonce') || '';
			var email = sessionStorage.getItem('auth_login_email') || '';

			postForm(urlApiLoginCodeVerify, {"verification_code": verificationCode, "nonce": nonce, "email": email}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return loginFormRaiseError(response.message);
				}

				$$.setAuthToken(response.data.token);
				$$.setAuthUser(response.data.user);
				sessionStorage.removeItem('auth_login_nonce');
				sessionStorage.removeItem('auth_login_email');
				loginFormRaiseSuccess('Verification successful');
				setTimeout(function () {
					$$.to(urlOnSuccess);
				}, 2000);
			}).catch(function (error) {
				console.log(error);
				spinner(false);
				return loginFormRaiseError('There was an error. Try again later!');
			});
		}
		onReady(function () {
			var digits = Array.prototype.slice.call(document.querySelectorAll('.otp-digit'));
			var hidden = document.querySelector('input[name=verification_code]');

			function syncDigits() {
				var code = '';
				digits.forEach(function (d) { code += d.value; });
				hidden.value = code;
				return code;
			}

			digits.forEach(function (d, i) {
				d.addEventListener('input', function () {
					var v = d.value.replace(/[^0-9]/g, '');
					d.value = v.slice(-1);
					syncDigits();
					if (v && digits[i + 1]) {
						digits[i + 1].focus();
					}
					if (syncDigits().length === digits.length) {
						loginFormValidate();
					}
				});
				d.addEventListener('keydown', function (e) {
					if (e.key === 'Backspace' && !d.value && digits[i - 1]) {
						digits[i - 1].focus();
					}
				});
				d.addEventListener('paste', function (e) {
					e.preventDefault();
					var pasted = ((e.clipboardData || window.clipboardData).getData('text'))
						.replace(/[^0-9]/g, '').slice(0, digits.length);
					digits.forEach(function (d2, j) { d2.value = pasted[j] || ''; });
					syncDigits();
					digits[Math.min(pasted.length, digits.length - 1)].focus();
					if (pasted.length === digits.length) {
						loginFormValidate();
					}
				});
			});
			if (digits[0]) { digits[0].focus(); }
		});
	`
}
