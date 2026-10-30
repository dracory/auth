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

		function logoutFormValidate() {
			spinner(true);

			postForm(urlApiLogout, {}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return logoutFormRaiseError(response.message);
				}

				authStore('AuthToken', response.data.token);
				authStore('AuthUser', response.data.user);
				logoutFormRaiseSuccess('Success');
				setTimeout(function () {
					authTo(urlOnSuccess);
				}, 2000);
			}).catch(function (error) {
				console.log(error);
				spinner(false);
				return logoutFormRaiseError('There was an error. Try again later!');
			});
		}
	`
}
