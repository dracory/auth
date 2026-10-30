
		var urlApiLogout = "__URL_API_LOGOUT__";
		var urlOnSuccess = "__URL_ON_SUCCESS__";

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
	