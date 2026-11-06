// Legal-link modal runtime: delegated click handler for links rendered
// with data-legal-modal (LegalLink.Target == LinkTargetModal). Opens the
// destination in an iframe inside a lightweight overlay so users can read
// the terms without leaving the auth form. The framed page must permit
// framing — the host's X-Frame-Options / CSP frame-ancestors can block it.
(function () {
	function openLegalModal(url, title) {
		var overlay = document.createElement('div');
		overlay.style.cssText = 'position:fixed;inset:0;background:rgba(0,0,0,.6);z-index:9999;display:flex;align-items:center;justify-content:center;padding:1rem;';

		var box = document.createElement('div');
		box.style.cssText = 'background:#fff;color:#000;border-radius:.5rem;width:100%;max-width:900px;height:80vh;display:flex;flex-direction:column;overflow:hidden;box-shadow:0 10px 40px rgba(0,0,0,.4);';

		var head = document.createElement('div');
		head.style.cssText = 'display:flex;align-items:center;justify-content:space-between;padding:.75rem 1rem;border-bottom:1px solid #ddd;font-weight:600;';
		var headText = document.createElement('span');
		headText.textContent = title || 'Legal';
		var close = document.createElement('button');
		close.type = 'button';
		close.textContent = '\u00d7';
		close.style.cssText = 'border:0;background:none;font-size:1.5rem;line-height:1;cursor:pointer;';
		close.onclick = function () { overlay.remove(); };
		head.appendChild(headText);
		head.appendChild(close);

		var frame = document.createElement('iframe');
		frame.src = url;
		frame.style.cssText = 'flex:1;border:0;width:100%;';

		box.appendChild(head);
		box.appendChild(frame);
		overlay.appendChild(box);

		overlay.addEventListener('click', function (e) {
			if (e.target === overlay) { overlay.remove(); }
		});
		document.addEventListener('keydown', function esc(e) {
			if (e.key === 'Escape') {
				overlay.remove();
				document.removeEventListener('keydown', esc);
			}
		});

		document.body.appendChild(overlay);
	}

	document.addEventListener('click', function (e) {
		var el = e.target && e.target.closest ? e.target.closest('[data-legal-modal]') : null;
		if (!el) { return; }
		e.preventDefault();
		openLegalModal(el.getAttribute('href'), el.textContent.trim());
	});
})();
