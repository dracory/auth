package page_login

import (
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

// alternativesBlock renders an "OR" divider with one outlined button per
// sibling login method.
func alternativesBlock(x LoginExtras) string {
	if len(x.Alternatives) == 0 {
		return ""
	}
	btns := ""
	for _, alt := range x.Alternatives {
		icon := ""
		if alt.Icon != "" {
			icon = hb.NewI().Class("bi " + alt.Icon).Style("margin-right:8px").ToHTML()
		}
		btns += hb.NewHyperlink().
			Class("btn btn-outline-primary w-100 mb-2").
			Href(alt.URL).
			HTML(icon + hb.NewSpan().Text(alt.Label).ToHTML()).ToHTML()
	}
	divider := hb.NewDiv().
		Class("text-center text-muted my-3").
		Style("display:flex;align-items:center;gap:10px;font-size:.8rem;letter-spacing:.12em").
		HTML(`<span style="flex:1;border-top:1px solid currentColor;opacity:.4"></span>OR<span style="flex:1;border-top:1px solid currentColor;opacity:.4"></span>`).
		ToHTML()
	return divider + btns
}

// LoginPasswordlessContent builds the HTML content for the passwordless login page.
func LoginPasswordlessContent(enableRegistration bool, urlRegister string) string {
	return LoginPasswordlessContentX(enableRegistration, urlRegister, LoginExtras{})
}

// LoginPasswordlessContentX builds the passwordless login page with extras.
func LoginPasswordlessContentX(enableRegistration bool, urlRegister string, x LoginExtras) string {
	// Elements for the form
	alertSuccess := hb.NewDiv().
		Class("alert alert-success").
		Style("display:none")
	alertDanger := hb.NewDiv().
		Class("alert alert-danger").
		Style("display:none")
	alertGroup := hb.NewDiv().
		Class("alert-group").
		Child(alertSuccess).
		Child(alertDanger)

	header := hb.NewHeading5().
		Text("Login").
		Style("margin:0px;")

	emailLabel := hb.NewLabel().
		Text("E-mail Address")
	emailInput := hb.NewInput().
		Class("form-control").
		Name("email").
		Placeholder("Enter e-mail address")
	emailFormGroup := hb.NewDiv().
		Class("form-group mt-3").
		Child(emailLabel).
		Child(emailInput)

	buttonLogin := hb.NewButton().
		Class("ButtonLogin btn btn-lg btn-success btn-block w-100").
		OnClick("loginFormValidate()").
		Children([]hb.TagInterface{
			hb.NewI().Class("bi bi-send").Style("margin-right:8px;margin-top:-2px;"),
			hb.NewSpan().Text("Send me a login code"),
			hb.NewDiv().
				Class("ImgLoading spinner-border spinner-border-sm text-light").
				Style("display:none;margin-left:10px;"),
		})
	buttonLoginFormGroup := hb.NewDiv().
		Class("form-group mt-3 mb-3").
		Child(buttonLogin)
	buttonRegister := hb.NewHyperlink().
		Class("btn btn-info text-white float-start").
		Children([]hb.TagInterface{
			hb.NewI().
				Class("bi bi-person-circle").
				Style("margin-right:8px;margin-top:-2px;"),
			hb.NewSpan().Text("Register"),
		}).Href(urlRegister)

	// Add elements in a card
	cardHeader := hb.NewDiv().
		Class("card-header").
		Child(header)
	cardBody := hb.NewDiv().
		Class("card-body").
		Children([]hb.TagInterface{
			alertGroup,
			emailFormGroup,
			buttonLoginFormGroup,
			hb.NewHTML(alternativesBlock(x)),
		})
	cardFooter := hb.NewDiv().
		Class("card-footer").
		AddChildren([]hb.TagInterface{})
	if enableRegistration {
		cardFooter.AddChild(buttonRegister)
	}

	card := hb.NewDiv().
		Class("card card-default").
		Style("margin:0 auto;max-width: 360px;").
		Child(cardHeader).
		Child(cardBody).
		Child(cardFooter)

	container := hb.NewDiv().
		Class("container").
		HTML(brandBlock(x)).
		Child(card)

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
	// Elements for the form
	alertSuccess := hb.NewDiv().Class("alert alert-success").Style("display:none")
	alertDanger := hb.NewDiv().Class("alert alert-danger").Style("display:none")
	alertGroup := hb.NewDiv().Class("alert-group").AddChild(alertSuccess).AddChild(alertDanger)

	header := hb.NewHeading5().Text("Login").Style("margin:0px;")

	emailLabel := hb.NewLabel().
		HTML("E-mail Address")
	emailInput := hb.NewInput().
		Class("form-control").
		Name("email").
		Placeholder("Enter e-mail address")
	emailFormGroup := hb.NewDiv().Class("form-group mt-3").
		Child(emailLabel).
		Child(emailInput)

	passwordLabel := hb.NewLabel().
		HTML("Password")
	passwordInput := hb.NewInput().Class("form-control").
		Name("password").
		Type("password").
		Placeholder("Enter password")
	passwordFormGroup := hb.NewDiv().Class("form-group mt-3").
		Child(passwordLabel).
		Child(passwordInput)

	buttonLogin := hb.NewButton().
		Class("ButtonLogin btn btn-lg btn-success text-white btn-block w-100").
		Children([]hb.TagInterface{
			hb.NewI().Class("bi bi-door-open").Style("margin-right:8px;margin-top:-2px;"),
			hb.NewSpan().
				HTML("Log in"),
			hb.NewDiv().
				Class("ImgLoading spinner-border spinner-border-sm text-light").
				Style("display:none;margin-left:10px;"),
		}).
		OnClick("loginFormValidate()")

	buttonLoginFormGroup := hb.NewDiv().Class("form-group mt-3 mb-3").Child(buttonLogin)

	buttonForgotPassword := hb.NewHyperlink().
		Class("d-block text-center mb-3").
		Href(urlPasswordRestore).
		Child(hb.NewSpan().Text("Forgot password?"))

	// Add elements in a card
	cardHeader := hb.NewDiv().Class("card-header").AddChild(header)
	cardBody := hb.NewDiv().Class("card-body").
		AddChildren([]hb.TagInterface{
			alertGroup,
			emailFormGroup,
			passwordFormGroup,
			buttonLoginFormGroup,
			buttonForgotPassword,
			hb.NewHTML(alternativesBlock(x)),
		})
	cardFooter := hb.NewDiv().Class("card-footer text-center")

	if enableRegistration {
		cardFooter.Child(hb.NewDiv().
			Style("color:#6c757d").
			Children([]hb.TagInterface{
				hb.NewSpan().Text("Don't have an account? "),
				hb.NewHyperlink().Href(urlRegister).Child(hb.NewSpan().Text("Register")),
			}))
	}

	card := hb.NewDiv().Class("card card-default").
		Style("margin:0 auto;max-width: 360px;").
		Child(cardHeader).
		Child(cardBody).
		Child(cardFooter)

	container := hb.NewDiv().Class("container").HTML(brandBlock(x)).AddChild(card)

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
