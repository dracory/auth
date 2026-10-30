
		var urlApiPasswordRestore = "__URL_API_PASSWORD_RESTORE__";
		var urlOnSuccess = "__URL_ON_SUCCESS__";

		function passwordRestoreFormValidate() {
			var first_name = inputVal('first_name');
			var last_name = inputVal('last_name');
			var email = inputVal('email');

			spinner(true);

			postForm(urlApiPasswordRestore, {"first_name": first_name, "last_name": last_name, "email": email}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return passwordRestoreFormRaiseError(response.message);
				}

				passwordRestoreFormRaiseSuccess('Success');
				setTimeout(function () {
					window.location.href = urlOnSuccess;
				}, 2000);
			}).catch(function (error) {
				console.log(error);
				spinner(false);
				return passwordRestoreFormRaiseError('There was an error. Try again later!');
			});
		}
		onReady(function () {
			var el = document.querySelector('input[name=first_name]');
			if (el) { el.focus(); }
		});
	