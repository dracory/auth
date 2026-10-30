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

		/**
		 * Raises an error message
		 * @param  {String} error
		 * @returns  {Boolean}
		 */
		function passwordRestoreFormRaiseError(error) {
			$('div.alert-success').html('').hide();
			$('div.alert-danger').html(error).show();
			setTimeout(function () {
				$('div.alert-danger').html('').hide();
			}, 10000);
			return false;
		}

		function passwordRestoreFormRaiseSuccess(success) {
			$('div.alert-danger').html('').hide();
			$('div.alert-success').html(success).show();
			setTimeout(function () {
				$('div.alert-success').html('').hide();
			}, 10000);
			return false;
		}

		/**
		 * Validates the Password Restore Form
		 * and sends the request to the backend
		 * @returns  {Boolean}
		 */
		function passwordRestoreFormValidate() {
			var first_name = $.trim($('input[name=first_name]').val());
			var last_name = $.trim($('input[name=last_name]').val());
			var email = $.trim($('input[name=email]').val());

			$('.ButtonContinue .imgLoading').show();

			var data = {"first_name": first_name, "last_name": last_name, "email": email};

			$.post(urlApiPasswordRestore, data).then(function (response) {
				$('.ButtonContinue .imgLoading').hide();

				if (response.status !== "success") {
					return passwordRestoreFormRaiseError(response.message);
				}

				passwordRestoreFormRaiseSuccess('Success');
				$('div.alert-danger').html('').hide();

				setTimeout(function () {
					window.location.href=urlOnSuccess;
				}, 2000);
				return;
			}).fail(function (error) {
				console.log(error);
				$('.ButtonContinue .imgLoading').hide();
				return passwordRestoreFormRaiseError('There was an error. Try again later!');
			});
		}
		$(function () {
			$("input[name=first_name").focus();
		});
	`
}
