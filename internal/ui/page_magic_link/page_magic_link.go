// Package page_magic_link renders the package's generic magic-link login
// page — an email form posting to the same path that serves it (the
// method's send ajax is mounted on the same route).
package page_magic_link

import (
	"net/http"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/auth/types"
	"github.com/dracory/hb"
)

// PageMagicLink renders a minimal "email me a sign-in link" form. The form
// posts to window.location.pathname so it works whether the method is
// mounted as the primary (/login) or a secondary (/login-magiclink) route.
func PageMagicLink(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
	if shared.RedirectAuthKnightIfActive(w, r, a) {
		return
	}

	urlRegister := links.Register(a.GetEndpoint())
	enableRegistration := a.IsRegistrationEnabled()

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
			hb.NewI().Class("bi bi-envelope").Style("margin-right:8px;margin-top:-2px;"),
			hb.NewSpan().Text("Email me a sign-in link"),
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

	cardHeader := hb.NewDiv().
		Class("card-header").
		Child(header)
	cardBody := hb.NewDiv().
		Class("card-body").
		Children([]hb.TagInterface{
			alertGroup,
			emailFormGroup,
			buttonLoginFormGroup,
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

	content := hb.NewDiv().
		Class("container").
		Child(card).
		ToHTML()

	scripts := `
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

			// The send ajax shares this page's path (GET renders, POST sends).
			var data = {"email": email};

			$.post(window.location.pathname, data).then(function (response) {
				$('.ButtonLogin .ImgLoading').hide();

				if (response.status !== "success") {
					return loginFormRaiseError(response.message);
				}

				loginFormRaiseSuccess('Check your inbox — we emailed you a sign-in link.');
				$('input[name=email]').val('');
			}).fail(function () {
				$('.ButtonLogin .ImgLoading').hide();
				loginFormRaiseError('Request failed. Please try again.');
			});
			return false;
		}
	`

	shared.PageRender(w, r, shared.PageOptions{
		Title:      "Login",
		Auth:       a,
		Layout:     a.GetLayout(),
		Content:    content,
		Scripts:    scripts,
		Logger:     a.GetLogger(),
		LogMessage: "failed to write magic link login page response",
	})
}
