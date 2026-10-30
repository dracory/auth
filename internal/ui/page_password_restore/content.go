package page_password_restore

import (
	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/hb"
)

// PasswordRestoreContent builds the HTML for the password restore page.
func PasswordRestoreContent(enableRegistration bool, urlLogin, urlRegister string) string {
	emailInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("email").
		Placeholder("you@example.com").
		Attr("autocomplete", "username").
		Attr("required", "required")

	card := shared.Card(
		shared.AlertGroup(),
		shared.NameRow(false),
		shared.FieldGroup("Email", emailInput),
		shared.PrimaryButton("ButtonContinue", "bi-envelope", "Send reset link", "passwordRestoreFormValidate()"),
	)

	container := hb.NewDiv().Children([]hb.TagInterface{
		shared.PageHeader("Forgot your password?", "Enter your email and we'll send you a reset link."),
		card,
		shared.FooterText("Remembered it?", urlLogin, "Login"),
	})

	if enableRegistration {
		container.Child(shared.FooterText("Don't have an account?", urlRegister, "Register"))
	}

	return container.ToHTML()
}

// PasswordRestoreScripts builds the JS for the password restore page.
func PasswordRestoreScripts(urlApiPasswordRestore, urlOnSuccess string) string {
	return `
		var urlApiPasswordRestore = "` + urlApiPasswordRestore + `";
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

		function passwordRestoreFormValidate() {
			var first_name = inputVal('first_name');
			var last_name = inputVal('last_name');
			var email = inputVal('email');

			spinner(true);

			postForm(urlApiPasswordRestore, {"first_name": first_name, "last_name": last_name, "email": email}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return passwordRestoreFormRaiseError(response.message);
				}

				passwordRestoreFormRaiseSuccess('Success');
				setTimeout(function () {
					window.location.href = urlOnSuccess;
				}, 2000);
			}).catch(function (error) {
				console.log(error);
				spinner(false);
				return passwordRestoreFormRaiseError('There was an error. Try again later!');
			});
		}
		onReady(function () {
			var el = document.querySelector('input[name=first_name]');
			if (el) { el.focus(); }
		});
	`
}
