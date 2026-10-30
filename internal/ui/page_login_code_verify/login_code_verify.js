
		var urlApiLoginCodeVerify = "__URL_API_LOGIN_CODE_VERIFY__";
		var urlOnSuccess = "__URL_ON_SUCCESS__";

		function loginFormValidate() {
			var verificationCode = inputVal('verification_code');

			if (verificationCode === '') {
				return loginFormRaiseError('Code is required');
			}

			spinner(true);

			// The nonce travels via sessionStorage (set by the login page),
			// not the URL — keeps it out of history, logs, and Referer.
			var nonce = sessionStorage.getItem('auth_login_nonce') || '';
			var email = sessionStorage.getItem('auth_login_email') || '';
			var remember = sessionStorage.getItem('auth_login_remember') || '';

			postForm(urlApiLoginCodeVerify, {"verification_code": verificationCode, "nonce": nonce, "email": email, "remember": remember}).then(function (response) {
				spinner(false);

				if (response.status !== "success") {
					return loginFormRaiseError(response.message);
				}

				authStore('AuthToken', response.data.token);
				authStore('AuthUser', response.data.user);
				sessionStorage.removeItem('auth_login_nonce');
				sessionStorage.removeItem('auth_login_email');
				sessionStorage.removeItem('auth_login_remember');
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
			var digits = Array.prototype.slice.call(document.querySelectorAll('.otp-digit'));
			var hidden = document.querySelector('input[name=verification_code]');

			function syncDigits() {
				var code = '';
				digits.forEach(function (d) { code += d.value; });
				hidden.value = code;
				return code;
			}

			digits.forEach(function (d, i) {
				d.addEventListener('input', function () {
					var v = d.value.replace(/[^0-9A-Za-z]/g, '').toUpperCase();
					d.value = v.slice(-1);
					syncDigits();
					if (v && digits[i + 1]) {
						digits[i + 1].focus();
					}
					if (syncDigits().length === digits.length) {
						loginFormValidate();
					}
				});
				d.addEventListener('keydown', function (e) {
					if (e.key === 'Backspace' && !d.value && digits[i - 1]) {
						digits[i - 1].focus();
					}
				});
				d.addEventListener('paste', function (e) {
					e.preventDefault();
					var pasted = ((e.clipboardData || window.clipboardData).getData('text'))
						.replace(/[^0-9A-Za-z]/g, '').toUpperCase().slice(0, digits.length);
					digits.forEach(function (d2, j) { d2.value = pasted[j] || ''; });
					syncDigits();
					digits[Math.min(pasted.length, digits.length - 1)].focus();
					if (pasted.length === digits.length) {
						loginFormValidate();
					}
				});
			});
			if (digits[0]) { digits[0].focus(); }
		});
	