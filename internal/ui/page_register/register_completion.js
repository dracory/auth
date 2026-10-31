
		const urlApiRegister = "__URL_API_REGISTER__";
		const urlOnSuccess = "__URL_ON_SUCCESS__";

		function registerFormValidate() {
			const first_name = inputVal('first_name');
			const last_name = inputVal('last_name');

			if (first_name === '') {
				return registerFormRaiseError('First name is required');
			}
			if (last_name === '') {
				return registerFormRaiseError('Last name is required');
			}

			spinner(true);

			postForm(urlApiRegister, {"first_name": first_name, "last_name": last_name}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return registerFormRaiseError(response.message);
				}

				registerFormRaiseSuccess('Success');
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
			const el = document.querySelector('input[name=first_name]');
			if (el) { el.focus(); }
		});
