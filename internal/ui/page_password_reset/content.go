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
		/**
		 * Raises an error message
		 * @param  {String} error
		 * @returns  {Boolean}
		 */
		function resetFormRaiseError(error) {
			$('div.alert-success').html('').hide();
			$('div.alert-danger').html(error).show();
			setTimeout(function () {
				$('div.alert-danger').html('').hide();
			}, 10000);
			return false;
		}

		function resetFormRaiseSuccess(success) {
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
		function resetFormValidate() {
			var token = $.trim($('input[name=token]').val());
			var password = $.trim($('input[name=password]').val());
			var passwordConfirm = $.trim($('input[name=password_confirm]').val());

			$('.ButtonContinue .imgLoading').show();

			var data = {"password": password, "password_confirm": passwordConfirm, "token": token};

			$.post(urlApiPasswordReset, data).then(function (response) {
				$('.ButtonContinue .imgLoading').hide();

				if (response.status !== "success") {
					return resetFormRaiseError(response.message);
				}

				resetFormRaiseSuccess('Success');
				$('div.alert-danger').html('').hide();
				setTimeout(function () {
					window.location.href=urlOnSuccess;
				}, 2000);
				return;
			}).fail(function (error) {
				console.log(error);
				$('.ButtonContinue .imgLoading').hide();
				return resetFormRaiseError('There was an error. Try again later!');
			});
		}
		$(function () {
			$("input[name=first_name").focus();
		});
	`
}
