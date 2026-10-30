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
