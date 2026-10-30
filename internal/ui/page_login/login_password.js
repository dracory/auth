
		var urlApiLogin = "__URL_API_LOGIN__";
		var urlOnSuccess = "__URL_ON_SUCCESS__";

		function loginFormValidate() {
			var email = inputVal('email');
			var password = inputVal('password');

			if (email === '') {
				return loginFormRaiseError('Email is required');
			}

			if (password === '') {
				return loginFormRaiseError('Password is required');
			}

			spinner(true);

			var remember = document.querySelector('input[name=remember]');
			postForm(urlApiLogin, {"email": email, "password": password, "remember": remember && remember.checked ? "true" : ""}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return loginFormRaiseError(response.message);
				}

				authStore('AuthToken', response.data.token);
				authStore('AuthUser', response.data.user);
				loginFormRaiseSuccess('Success');
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
	