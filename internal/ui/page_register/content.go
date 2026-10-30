package page_register

import (
	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/hb"
)

// RegisterPasswordlessContent builds the HTML for the passwordless registration page.
func RegisterPasswordlessContent(urlLogin string) string {
	emailInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("email").
		Placeholder("you@example.com").
		Attr("autocomplete", "username").
		Attr("required", "required")

	card := shared.Card(
		shared.AlertGroup(),
		shared.NameRow(true),
		shared.FieldGroup("Email", emailInput),
		shared.PrimaryButton("buttonRegister", "bi-person-plus", "Create account", "registerFormValidate()"),
	)

	return hb.NewDiv().Children([]hb.TagInterface{
		shared.PageHeader("Create your account", "A few details and you're in."),
		card,
		shared.FooterText("Already have an account?", urlLogin, "Login"),
	}).ToHTML()
}

// RegisterPasswordlessScripts builds the JS for the passwordless registration page.
func RegisterPasswordlessScripts(urlApiRegister, urlOnSuccess string) string {
	return `
		var urlApiRegister = "` + urlApiRegister + `";
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

		function registerFormValidate() {
			var first_name = inputVal('first_name');
			var last_name = inputVal('last_name');
			var email = inputVal('email');

			if (first_name === '') {
				return registerFormRaiseError('First name is required');
			}
			if (last_name === '') {
				return registerFormRaiseError('Last name is required');
			}
			if (email === '') {
				return registerFormRaiseError('Email is required');
			}

			spinner(true);

			postForm(urlApiRegister, {"first_name": first_name, "last_name": last_name, "email": email}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return registerFormRaiseError(response.message);
				}

				registerFormRaiseSuccess('Success');
				// Hand the nonce to the code-verify page via sessionStorage —
				// query params would leak into history, logs, and Referer.
				if (response.data && response.data.nonce) {
					sessionStorage.setItem('auth_register_nonce', response.data.nonce);
				}
				setTimeout(function () {
					window.location.href = urlOnSuccess;
				}, 100);
			}).catch(function (error) {
				console.log(error);
				spinner(false);
				return registerFormRaiseError('There was an error. Try again later!');
			});
		}
		onReady(function () {
			var el = document.querySelector('input[name=first_name]');
			if (el) { el.focus(); }
		});
	`
}

// RegisterUsernameAndPasswordContent builds the HTML for the username/password registration page.
func RegisterUsernameAndPasswordContent(urlLogin, urlPasswordRestore string) string {
	emailInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("email").
		Placeholder("you@example.com").
		Attr("autocomplete", "username").
		Attr("required", "required")
	passwordInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("password").
		Type(hb.TYPE_PASSWORD).
		Placeholder("Min 8 characters").
		Attr("autocomplete", "new-password").
		Attr("required", "required")

	card := shared.Card(
		shared.AlertGroup(),
		shared.NameRow(true),
		shared.FieldGroup("Email", emailInput),
		shared.FieldGroup("Password", passwordInput),
		shared.PrimaryButton("buttonRegister", "bi-person-plus", "Create account", "registerFormValidate()"),
		hb.NewDiv().Class("text-center small mt-3").Child(
			hb.NewHyperlink().Href(urlPasswordRestore).Text("Forgot password?")),
	)

	return hb.NewDiv().Children([]hb.TagInterface{
		shared.PageHeader("Create your account", "A few details and you're in."),
		card,
		shared.FooterText("Already have an account?", urlLogin, "Login"),
	}).ToHTML()
}

// RegisterUsernameAndPasswordScripts builds the JS for the username/password registration page.
func RegisterUsernameAndPasswordScripts(urlApiRegister, urlOnSuccess string) string {
	return `
		var urlApiRegister = "` + urlApiRegister + `";
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

		function registerFormValidate() {
			var first_name = inputVal('first_name');
			var last_name = inputVal('last_name');
			var email = inputVal('email');
			var password = inputVal('password');

			if (first_name === '') {
				return registerFormRaiseError('First name is required');
			}
			if (last_name === '') {
				return registerFormRaiseError('Last name is required');
			}
			if (email === '') {
				return registerFormRaiseError('Email is required');
			}
			if (password === '') {
				return registerFormRaiseError('Password is required');
			}

			spinner(true);

			postForm(urlApiRegister, {"first_name": first_name, "last_name": last_name, "email": email, "password": password}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return registerFormRaiseError(response.message);
				}

				registerFormRaiseSuccess('Success');
				if (response.data && response.data.nonce) {
					sessionStorage.setItem('auth_register_nonce', response.data.nonce);
				}
				setTimeout(function () {
					window.location.href = urlOnSuccess;
				}, 2000);
			}).catch(function (error) {
				console.log(error);
				spinner(false);
				return registerFormRaiseError('There was an error. Try again later!');
			});
		}
		onReady(function () {
			var el = document.querySelector('input[name=first_name]');
			if (el) { el.focus(); }
		});
	`
}

// RegisterAuthKnightContent builds the HTML for the AuthKnight registration page.
// The email field is read-only (pre-populated from the verified AuthKnight email).
// akKey is the temporary key that maps to the verified email in the temp key store.
func RegisterAuthKnightContent(urlLogin string, verifiedEmail string, akKey string) string {
	emailGroup := hb.NewDiv().Class("mb-3").Children([]hb.TagInterface{
		hb.NewLabel().Class("form-label small fw-bold text-uppercase text-secondary").Text("Email (verified)"),
		hb.NewDiv().Class("form-control-plaintext text-muted").Text(verifiedEmail),
		hb.NewInput().Type("hidden").Name("email").Value(verifiedEmail),
	})

	akKeyInput := hb.NewInput().Type("hidden").Name("ak_key").Value(akKey)

	card := shared.Card(
		shared.AlertGroup(),
		shared.NameRow(true),
		emailGroup,
		akKeyInput,
		shared.PrimaryButton("buttonRegister", "bi-person-plus", "Create account", "registerFormValidate()"),
	)

	return hb.NewDiv().Children([]hb.TagInterface{
		shared.PageHeader("Create your account", "Confirm your details to finish registration."),
		card,
		shared.FooterText("Already have an account?", urlLogin, "Login"),
	}).ToHTML()
}

// RegisterAuthKnightScripts builds the JS for the AuthKnight registration page.
func RegisterAuthKnightScripts(urlApiRegister, urlOnSuccess string) string {
	return `
		var urlApiRegister = "` + urlApiRegister + `";
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

		function registerFormValidate() {
			var first_name = inputVal('first_name');
			var last_name = inputVal('last_name');
			var email = inputVal('email');
			var ak_key = inputVal('ak_key');

			if (first_name === '') {
				return registerFormRaiseError('First name is required');
			}
			if (last_name === '') {
				return registerFormRaiseError('Last name is required');
			}

			spinner(true);

			postForm(urlApiRegister, {"first_name": first_name, "last_name": last_name, "email": email, "ak_key": ak_key}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return registerFormRaiseError(response.message);
				}

				registerFormRaiseSuccess('Success');
				if (response.data && response.data.nonce) {
					sessionStorage.setItem('auth_register_nonce', response.data.nonce);
				}
				setTimeout(function () {
					window.location.href = urlOnSuccess;
				}, 100);
			}).catch(function (error) {
				console.log(error);
				spinner(false);
				return registerFormRaiseError('There was an error. Try again later!');
			});
		}
		onReady(function () {
			var el = document.querySelector('input[name=first_name]');
			if (el) { el.focus(); }
		});
	`
}
