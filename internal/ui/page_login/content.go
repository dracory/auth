package page_login

import (
	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/auth/types"
	"github.com/dracory/hb"
)

// LoginExtras carries optional branding and sibling-method links for the
// built-in login pages.
type LoginExtras struct {
	BrandName    string
	BrandTagline string
	BrandLogoURL string
	Alternatives []types.LoginAlternative
}

// brandBlock renders the optional brand header above the login card.
func brandBlock(x LoginExtras) string {
	if x.BrandName == "" && x.BrandLogoURL == "" {
		return ""
	}
	logo := ""
	if x.BrandLogoURL != "" {
		logo = hb.NewImage().
			Src(x.BrandLogoURL).
			Style("width:56px;height:56px;object-fit:contain;margin-bottom:10px;border-radius:14px").
			ToHTML()
	}
	name := ""
	if x.BrandName != "" {
		name = hb.NewDiv().
			Text(x.BrandName).
			Style("font-size:1.45rem;font-weight:800;letter-spacing:.04em").ToHTML()
	}
	tagline := ""
	if x.BrandTagline != "" {
		tagline = hb.NewDiv().
			Text(x.BrandTagline).
			Style("font-size:.72rem;font-weight:700;letter-spacing:.22em;opacity:.65").ToHTML()
	}
	return hb.NewDiv().
		Class("auth-brand").
		Style("text-align:center;margin:0 auto 22px;max-width:400px").
		HTML(logo + name + tagline).ToHTML()
}

// alternativesBlock renders an "or" divider with one outlined button per
// sibling login method.
func alternativesBlock(x LoginExtras) string {
	if len(x.Alternatives) == 0 {
		return ""
	}
	btns := ""
	for _, alt := range x.Alternatives {
		btns += shared.AltLink(alt.URL, alt.Icon, alt.Label).ToHTML()
	}
	return shared.OrDivider().ToHTML() + btns
}

// LoginPasswordlessContent builds the HTML content for the passwordless login page.
func LoginPasswordlessContent(enableRegistration bool, urlRegister string) string {
	return LoginPasswordlessContentX(enableRegistration, urlRegister, LoginExtras{})
}

// LoginPasswordlessContentX builds the passwordless login page with extras.
func LoginPasswordlessContentX(enableRegistration bool, urlRegister string, x LoginExtras) string {
	emailInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("email").
		Placeholder("you@example.com").
		Attr("autocomplete", "username").
		Attr("required", "required")

	card := shared.Card(
		shared.AlertGroup(),
		shared.FieldGroup("Email", emailInput),
		shared.PrimaryButton("ButtonLogin", "bi-send", "Send me a login code", "loginFormValidate()"),
		hb.NewHTML(alternativesBlock(x)),
	)

	container := hb.NewDiv().Children([]hb.TagInterface{
		hb.NewHTML(brandBlock(x)),
		shared.PageHeader("Login to your account", "No password needed. We'll email you a code."),
		card,
	})

	if enableRegistration {
		container.Child(shared.FooterText("Don't have an account?", urlRegister, "Register"))
	}

	return container.ToHTML()
}

// LoginPasswordlessScripts builds the JavaScript for the passwordless login page.
func LoginPasswordlessScripts(urlApiLogin, urlSuccess string) string {
	return `
		var urlApiLogin = "` + urlApiLogin + `";
		var urlOnSuccess = "` + urlSuccess + `";

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
			var email = inputVal('email');

			if (email === '') {
				return loginFormRaiseError('Email is required');
			}

			spinner(true);

			postForm(urlApiLogin, {"email": email}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return loginFormRaiseError(response.message);
				}

				loginFormRaiseSuccess('Success');
				// Hand the nonce (and email) to the code-verify page via
				// sessionStorage — query params would leak into browser
				// history, access logs, and Referer headers.
				if (response.data && response.data.nonce) {
					sessionStorage.setItem('auth_login_nonce', response.data.nonce);
					sessionStorage.setItem('auth_login_email', email);
				}
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
			var el = document.querySelector('input[name=email]');
			if (el) { el.focus(); }
		});
	`
}

// LoginContent builds the HTML content for the standard login page.
func LoginContent(enableRegistration bool, urlRegister, urlPasswordRestore string) string {
	return LoginContentX(enableRegistration, urlRegister, urlPasswordRestore, LoginExtras{})
}

// LoginContentX builds the standard login page with extras.
func LoginContentX(enableRegistration bool, urlRegister, urlPasswordRestore string, x LoginExtras) string {
	emailInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("email").
		Placeholder("you@example.com").
		Attr("autocomplete", "username").
		Attr("required", "required")
	passwordInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("password").
		Type("password").
		Placeholder("Enter your password").
		Attr("autocomplete", "current-password").
		Attr("required", "required")

	card := shared.Card(
		shared.AlertGroup(),
		shared.FieldGroup("Email", emailInput),
		shared.FieldGroup("Password", passwordInput),
		shared.PrimaryButton("ButtonLogin", "bi-box-arrow-in-right", "Sign in", "loginFormValidate()"),
		hb.NewDiv().Class("text-center small mt-3").Child(
			hb.NewHyperlink().Href(urlPasswordRestore).Text("Forgot password?")),
		hb.NewHTML(alternativesBlock(x)),
	)

	container := hb.NewDiv().Children([]hb.TagInterface{
		hb.NewHTML(brandBlock(x)),
		shared.PageHeader("Login to your account", "Enter your email and password to sign in."),
		card,
	})

	if enableRegistration {
		container.Child(shared.FooterText("Don't have an account?", urlRegister, "Register"))
	}

	return container.ToHTML()
}

// LoginScripts builds the JavaScript for the standard login page.
func LoginScripts(urlApiLogin, urlOnSuccess string) string {
	return `
		var urlApiLogin = "` + urlApiLogin + `";
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
			var email = inputVal('email');
			var password = inputVal('password');

			if (email === '') {
				return loginFormRaiseError('Email is required');
			}

			if (password === '') {
				return loginFormRaiseError('Password is required');
			}

			spinner(true);

			postForm(urlApiLogin, {"email": email, "password": password}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return loginFormRaiseError(response.message);
				}

				$$.setAuthToken(response.data.token);
				$$.setAuthUser(response.data.user);
				loginFormRaiseSuccess('Success');
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
			var el = document.querySelector('input[name=email]');
			if (el) { el.focus(); }
		});
	`
}
