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
			var email = $.trim($('input[name=email]').val());

			if (email === '') {
				return loginFormRaiseError('Email is required');
			}

			$('.ButtonLogin .ImgLoading').show();

			var data = {"email": email};

			$.post(urlApiLogin, data).then(function (response) {
				$('.ButtonLogin .ImgLoading').hide();

				if (response.status !== "success") {
					return loginFormRaiseError(response.message);
				}

				loginFormRaiseSuccess('Success');
				$('div.alert-danger').html('').hide();
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
				return;
			}).fail(function (error) {
				console.log(error);
				$('.ButtonLogin .ImgLoading').hide();
				return loginFormRaiseError('There was an error. Try again later!');
			});
		}
		$(function () {
			$("#email").focus();
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
			var email = $.trim($('input[name=email]').val());
			var password = $.trim($('input[name=password]').val());

			if (email === '') {
				return loginFormRaiseError('Email is required');
			}

			if (password === '') {
				return loginFormRaiseError('Password is required');
			}

			$('.ButtonLogin .ImgLoading').show();

			var data = {"email": email, "password": password};

			$.post(urlApiLogin, data).then(function (response) {
				$('.ButtonLogin .ImgLoading').hide();

				if (response.status !== "success") {
					return loginFormRaiseError(response.message);
				}

				$$.setAuthToken(response.data.token);
				$$.setAuthUser(response.data.user);
				loginFormRaiseSuccess('Success');
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
			$("#email").focus();
		});
	`
}
