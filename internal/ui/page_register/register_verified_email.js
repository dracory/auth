
		var urlApiRegister = "__URL_API_REGISTER__";
		var urlOnSuccess = "__URL_ON_SUCCESS__";

		function registerFormValidate() {
			var first_name = inputVal('first_name');
			var last_name = inputVal('last_name');
			var email = inputVal('email');
			var verified_key = inputVal('verified_key');

			if (first_name === '') {
				return registerFormRaiseError('First name is required');
			}
			if (last_name === '') {
				return registerFormRaiseError('Last name is required');
			}

			spinner(true);

			postForm(urlApiRegister, {"first_name": first_name, "last_name": last_name, "email": email, "verified_key": verified_key}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return registerFormRaiseError(response.message);
				}

				registerFormRaiseSuccess('Success');
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
			// Strip the verified_key bearer token from the address bar so it
			// does not linger in browser history (the form posts the hidden
			// field, not the URL).
			if (window.history && history.replaceState && /[?&]verified_key=/.test(location.search)) {
				history.replaceState(null, '', location.pathname);
			}
			var el = document.querySelector('input[name=first_name]');
			if (el) { el.focus(); }
		});
