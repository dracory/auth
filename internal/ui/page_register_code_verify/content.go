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

				$$.setAuthToken(response.data.token);
				$$.setAuthUser(response.data.user);
				sessionStorage.removeItem('auth_register_nonce');
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
			var el = document.querySelector('input[name=verification_code]');
			if (el) { el.focus(); }
		});
	`
}
