
		var urlApiRegisterCodeVerify = "__URL_API_REGISTER_CODE_VERIFY__";
		var urlOnSuccess = "__URL_ON_SUCCESS__";

		function registerCodeFormValidate() {
			var verificationCode = inputVal('verification_code');

			if (verificationCode === '') {
				return loginFormRaiseError('Code is required');
			}

			spinner(true);

			// The nonce travels via sessionStorage (set by the register
			// page), not the URL — keeps it out of history, logs, Referer.
			var nonce = sessionStorage.getItem('auth_register_nonce') || '';

			postForm(urlApiRegisterCodeVerify, {"verification_code": verificationCode, "nonce": nonce}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return loginFormRaiseError(response.message);
				}

				authStore('AuthToken', response.data.token);
				authStore('AuthUser', response.data.user);
				sessionStorage.removeItem('auth_register_nonce');
				loginFormRaiseSuccess('Verification successful');
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
			var el = document.querySelector('input[name=verification_code]');
			if (el) { el.focus(); }
		});
	