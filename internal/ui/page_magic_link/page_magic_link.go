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

	emailInput := hb.NewInput().
		Class("form-control form-control-lg").
		Name("email").
		Placeholder("you@example.com").
		Attr("autocomplete", "username").
		Attr("required", "required")

	card := shared.Card(
		shared.AlertGroup(),
		shared.FieldGroup("Email", emailInput),
		shared.PrimaryButton("ButtonLogin", "bi-envelope", "Send me a quick login link", "loginFormValidate()"),
	)

	container := hb.NewDiv().Children([]hb.TagInterface{
		shared.PageHeader("Login to your account", "We'll email you a secure link to log in instantly — no password needed."),
		card,
	})

	if enableRegistration {
		container.Child(shared.FooterText("Don't have an account?", urlRegister, "Register"))
	}

	content := container.ToHTML()

	scripts := `

		function loginFormValidate() {
			var email = inputVal('email');

			if (email === '') {
				return loginFormRaiseError('Email is required');
			}

			spinner(true);

			// The send ajax shares this page's path (GET renders, POST sends).
			postForm(window.location.pathname, {"email": email}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return loginFormRaiseError(response.message);
				}

				loginFormRaiseSuccess('Check your inbox — we emailed you a sign-in link.');
				document.querySelector('input[name=email]').value = '';
			}).catch(function () {
				spinner(false);
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
