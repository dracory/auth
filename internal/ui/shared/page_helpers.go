package shared

// pageHelpersJS is the shared vanilla-JS runtime for the built-in pages —
// injected once ahead of each page's own script so every page stays free
// of duplicated helpers.
//
// authStore/authTo replicate the storage format of the web.js client
// helper — obfuscated value + "&&expires" record — so a host app running
// web.js reads exactly what these pages write.
const pageHelpersJS = `
	// alertShow prefers Notiflix toasts (coursethread pattern); when the
	// library is absent it falls back to the card's inline alert divs.
	function alertShow(cls, message) {
		if (window.Notiflix && Notiflix.Notify) {
			if (cls === 'div.alert-danger') {
				Notiflix.Notify.failure(message);
			} else {
				Notiflix.Notify.success(message);
			}
			return false;
		}
		var other = cls === 'div.alert-danger' ? 'div.alert-success' : 'div.alert-danger';
		var o = document.querySelector(other);
		o.innerHTML = ''; o.style.display = 'none';
		var el = document.querySelector(cls);
		el.innerHTML = message; el.style.display = 'block';
		setTimeout(function () { el.innerHTML = ''; el.style.display = 'none'; }, 10000);
		return false;
	}
	function loginFormRaiseError(error) { return alertShow('div.alert-danger', error); }
	function loginFormRaiseSuccess(success) { return alertShow('div.alert-success', success); }
	function registerFormRaiseError(error) { return alertShow('div.alert-danger', error); }
	function registerFormRaiseSuccess(success) { return alertShow('div.alert-success', success); }
	function resetFormRaiseError(error) { return alertShow('div.alert-danger', error); }
	function resetFormRaiseSuccess(success) { return alertShow('div.alert-success', success); }
	function passwordRestoreFormRaiseError(error) { return alertShow('div.alert-danger', error); }
	function passwordRestoreFormRaiseSuccess(success) { return alertShow('div.alert-success', success); }
	function logoutFormRaiseError(error) { return alertShow('div.alert-danger', error); }
	function logoutFormRaiseSuccess(success) { return alertShow('div.alert-success', success); }

	// postForm posts url-encoded form data and parses the JSON response,
	// mirroring the $.post behavior the pages previously relied on.
	function postForm(url, data) {
		return fetch(url, {
			method: 'POST',
			headers: { 'Content-Type': 'application/x-www-form-urlencoded;charset=UTF-8' },
			body: new URLSearchParams(data)
		}).then(function (response) { return response.json(); });
	}
	function inputVal(name) {
		var el = document.querySelector('input[name=' + name + ']');
		return el ? el.value.trim() : '';
	}
	function spinner(show) {
		var el = document.querySelector('.ImgLoading');
		if (el) { el.style.display = show ? 'inline-block' : 'none'; }
	}
	function onReady(fn) {
		if (document.readyState === 'loading') {
			document.addEventListener('DOMContentLoaded', fn);
		} else {
			fn();
		}
	}

	var AUTH_REGISTRY_KEY = '8pVbaKePV3beCUZYbKSfujzucbcD3eqyJvCAUgQL8PbYe3VmAMSKC9esx8jV8M7KegPsxkDTpUKvu2UenQyPPjsDf92XnjtZh5GJRz8bQHZngNGKenKZHDD8';
	function authStore(name, value) {
		var src = JSON.stringify(value);
		var enc = [];
		for (var i = 0; i < src.length; i++) {
			enc.push(src.charCodeAt(i) + AUTH_REGISTRY_KEY.charCodeAt(i % AUTH_REGISTRY_KEY.length));
		}
		var now = new Date(); now.setTime(now.getTime() + 6e10);
		localStorage.setItem(name + AUTH_REGISTRY_KEY, JSON.stringify(enc));
		localStorage.setItem(name + AUTH_REGISTRY_KEY + '&&expires', now.toString());
	}
	function authTo(url) {
		if (/^https?:/i.test(url)) { window.location.href = url; return; }
		var root = window.location.protocol + '//' + window.location.hostname +
			(window.location.port ? ':' + window.location.port : '');
		window.location.href = root + '/' + url.replace(/^\/+|\/+$/g, '');
	}
`
