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

		function logoutFormValidate() {
			spinner(true);

			postForm(urlApiLogout, {}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return logoutFormRaiseError(response.message);
				}

				$$.setAuthToken(response.data.token);
				$$.setAuthUser(response.data.user);
				logoutFormRaiseSuccess('Success');
				setTimeout(function () {
					$$.to(urlOnSuccess);
				}, 2000);
			}).catch(function (error) {
				console.log(error);
				spinner(false);
				return logoutFormRaiseError('There was an error. Try again later!');
			});
		}
	`
}
