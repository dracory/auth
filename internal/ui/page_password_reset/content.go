package page_password_reset

import (
	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/hb"
)

// PasswordResetContent builds the HTML for the password reset page.
func PasswordResetContent(token, errorMessage, urlPasswordRestore, urlLogin, urlRegister string, enableRegistration bool) string {
	alertSuccess := hb.NewDiv().Class("alert alert-success").Style("display:none")
	alertDanger := hb.NewDiv().Class("alert alert-danger")
	if errorMessage != "" {
		alertDanger.Text(errorMessage)
	} else {
		alertDanger.Style("display:none")
	}
	alertGroup := hb.NewDiv().Class("alert-group").AddChild(alertSuccess).AddChild(alertDanger)

	var cardBody hb.TagInterface
	if errorMessage == "" {
		tokenInput := hb.NewInput().Name("token").Value(token)
		passwordInput := hb.NewInput().Class("form-control form-control-lg").Name("password").Type(hb.TYPE_PASSWORD).
			Placeholder("Min 8 characters").Attr("autocomplete", "new-password").Attr("required", "required")
		passwordConfirmInput := hb.NewInput().Class("form-control form-control-lg").Name("password_confirm").Type(hb.TYPE_PASSWORD).
			Placeholder("Repeat password").Attr("autocomplete", "new-password").Attr("required", "required")
		cardBody = shared.Card(
			alertGroup,
			tokenInput,
			shared.FieldGroup("New password", passwordInput),
			shared.FieldGroup("Confirm new password", passwordConfirmInput),
			shared.PrimaryButton("ButtonContinue", "bi-shield-check", "Reset password", "resetFormValidate()"),
		)
	} else {
		cardBody = shared.Card(
			alertGroup,
			hb.NewParagraph().Text("Sorry, there was an error processing your request. Please select one of the following options:"),
			hb.NewParagraph().AddChild(hb.NewHyperlink().Href(urlPasswordRestore).Text("request a reset of your password")),
			hb.NewParagraph().AddChild(hb.NewHyperlink().Href(urlLogin).Text("login to the system")),
			hb.NewParagraph().AddChild(hb.NewHyperlink().Href(urlRegister).Text("create a new account")),
		)
	}

	container := hb.NewDiv().Children([]hb.TagInterface{
		shared.PageHeader("Set a new password", "Choose a strong password for your account."),
		cardBody,
		shared.FooterText("Remembered it?", urlLogin, "Login"),
	})

	if enableRegistration {
		container.Child(shared.FooterText("Don't have an account?", urlRegister, "Register"))
	}

	return container.ToHTML()
}

// PasswordResetScripts builds the JS for the password reset page.
func PasswordResetScripts(urlApiPasswordReset, urlOnSuccess string) string {
	return `
		var urlApiPasswordReset = "` + urlApiPasswordReset + `";
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

		function resetFormValidate() {
			var token = inputVal('token');
			var password = inputVal('password');
			var passwordConfirm = inputVal('password_confirm');

			spinner(true);

			postForm(urlApiPasswordReset, {"password": password, "password_confirm": passwordConfirm, "token": token}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return resetFormRaiseError(response.message);
				}

				resetFormRaiseSuccess('Success');
				setTimeout(function () {
					window.location.href = urlOnSuccess;
				}, 2000);
			}).catch(function (error) {
				console.log(error);
				spinner(false);
				return resetFormRaiseError('There was an error. Try again later!');
			});
		}
		onReady(function () {
			var el = document.querySelector('input[name=password]');
			if (el) { el.focus(); }
		});
	`
}
