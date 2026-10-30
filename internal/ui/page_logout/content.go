package page_logout

import (
	"github.com/dracory/auth/internal/ui/shared"
	"github.com/dracory/hb"
)

// LogoutContent builds the HTML for the logout page.
func LogoutContent() string {
	card := shared.Card(
		shared.AlertGroup(),
		shared.PrimaryButton("ButtonLogout", "bi-box-arrow-right", "Logout", "logoutFormValidate()"),
	)

	return hb.NewDiv().Children([]hb.TagInterface{
		shared.PageHeader("Sign out", "Are you sure you want to sign out of your account?"),
		card,
	}).ToHTML()
}

// LogoutScripts builds the JS for the logout page.
func LogoutScripts(urlApiLogout, urlOnSuccess string) string {
	return `
		var urlApiLogout = "` + urlApiLogout + `";
		var urlOnSuccess = "` + urlOnSuccess + `";
		/**
		 * Raises an error message
		 * @param  {String} error
		 * @returns  {Boolean}
		 */
		function logoutFormRaiseError(error) {
			$('div.alert-success').html('').hide();
			$('div.alert-danger').html(error).show();
			setTimeout(function () {
				$('div.alert-danger').html('').hide();
			}, 10000);
			return false;
		}

		function logoutFormRaiseSuccess(success) {
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
		function logoutFormValidate() {
			$('.ButtonLogout .ImgLoading').show();

			var data = {};

			$.post(urlApiLogout, data).then(function (response) {
				$('.ButtonLogout .ImgLoading').hide();

				if (response.status !== "success") {
					return logoutFormRaiseError(response.message);
				}

				$$.setAuthToken(response.data.token);
				$$.setAuthUser(response.data.user);
				logoutFormRaiseSuccess('Success');
				$('div.alert-danger').html('').hide();
				setTimeout(function () {
					$$.to(urlOnSuccess);
				}, 2000);
				return;
			}).fail(function (error) {
				console.log(error);
				$('.ButtonLogout .ImgLoading').hide();
				return logoutFormRaiseError('There was an error. Try again later!');
			});
		}
		$(function () {
			$("#email").focus();
		});
	`
}
