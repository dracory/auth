
	(function () {
		if (window.AUTH_ASSETS_PROVIDED ||
			document.documentElement.getAttribute('data-auth-assets') === 'provided') {
			return;
		}
		function present(tag, attr, key) {
			return !!document.querySelector(tag + '[' + attr + '*="' + key + '"]');
		}
		var styles = [
			['bootstrap.min.css', 'BOOTSTRAP_CSS'],
			['bootstrap-icons', 'BOOTSTRAP_ICONS'],
			['notiflix', 'NOTIFLIX_CSS']
		];
		styles.forEach(function (s) {
			if (!present('link', 'href', s[0])) {
				var l = document.createElement('link');
				l.rel = 'stylesheet';
				l.href = s[1];
				document.head.appendChild(l);
			}
		});
		if (!present('script', 'src', 'notiflix')) {
			var sc = document.createElement('script');
			sc.src = 'NOTIFLIX_JS';
			document.head.appendChild(sc);
		}
	})();
