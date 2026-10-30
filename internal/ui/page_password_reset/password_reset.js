
		var urlApiPasswordReset = "__URL_API_PASSWORD_RESET__";
		var urlOnSuccess = "__URL_ON_SUCCESS__";

		function resetFormValidate() {
			var token = inputVal('token');
			var password = inputVal('password');
			var passwordConfirm = inputVal('password_confirm');

			spinner(true);

			postForm(urlApiPasswordReset, {"password": password, "password_confirm": passwordConfirm, "token": token}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return resetFormRaiseError(response.message);
				}

				resetFormRaiseSuccess('Success');
				setTimeout(function () {
					window.location.href = urlOnSuccess;
				}, 2000);
			}).catch(function (error) {
				console.log(error);
				spinner(false);
				return resetFormRaiseError('There was an error. Try again later!');
			});
		}
		onReady(function () {
			var el = document.querySelector('input[name=password]');
			if (el) { el.focus(); }
		});
	