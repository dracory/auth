
		var urlApiRegister = "__URL_API_REGISTER__";
		var urlOnSuccess = "__URL_ON_SUCCESS__";

		function registerFormValidate() {
			var first_name = inputVal('first_name');
			var last_name = inputVal('last_name');
			var email = inputVal('email');

			if (first_name === '') {
				return registerFormRaiseError('First name is required');
			}
			if (last_name === '') {
				return registerFormRaiseError('Last name is required');
			}
			if (email === '') {
				return registerFormRaiseError('Email is required');
			}

			spinner(true);

			postForm(urlApiRegister, {"first_name": first_name, "last_name": last_name, "email": email}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return registerFormRaiseError(response.message);
				}

				registerFormRaiseSuccess('Success');
				// Hand the nonce to the code-verify page via sessionStorage —
				// query params would leak into history, logs, and Referer.
				if (response.data && response.data.nonce) {
					sessionStorage.setItem('auth_register_nonce', response.data.nonce);
				}
				setTimeout(function () {
					window.location.href = urlOnSuccess;
				}, 100);
			}).catch(function (error) {
				console.log(error);
				spinner(false);
				return registerFormRaiseError('There was an error. Try again later!');
			});
		}
		onReady(function () {
			var el = document.querySelector('input[name=first_name]');
			if (el) { el.focus(); }
		});
	