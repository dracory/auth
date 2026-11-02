
		var urlApiLogin = "__URL_API_LOGIN__";
		var urlOnSuccess = "__URL_ON_SUCCESS__";

		function loginFormValidate() {
			var email = inputVal('email');

			if (email === '') {
				return loginFormRaiseError('Email is required');
			}

			spinner(true);

			var remember = document.querySelector('input[name=remember]');
			postForm(urlApiLogin, {"email": email, "remember": remember && remember.checked ? "true" : ""}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return loginFormRaiseError(response.message);
				}

				loginFormRaiseSuccess('Success');
				// Hand the nonce (and email) to the code-verify page via
				// sessionStorage — query params would leak into browser
				// history, access logs, and Referer headers.
				if (response.data && response.data.nonce) {
					sessionStorage.setItem('auth_login_nonce', response.data.nonce);
					sessionStorage.setItem('auth_login_email', email);
					sessionStorage.setItem('auth_login_remember', remember && remember.checked ? "true" : "");
				}
				setTimeout(function () {
					authTo(urlOnSuccess);
				}, 2000);
			}).catch(function (error) {
				console.log(error);
				spinner(false);
				return loginFormRaiseError('There was an error. Try again later!');
			});
		}
		onReady(function () {
			var el = document.querySelector('input[name=email]');
			if (el) { el.focus(); }
		});
	