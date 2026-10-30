/* dracory/auth docs — interactive documentation behaviors.
   Self-contained (no deps beyond Bootstrap's bundle for the navbar). */
(function () {
  'use strict';

  /* When served through html-preview.github.io the proxy rewrites relative
     links to raw.githubusercontent.com URLs — which the browser shows as
     plain text. Intercept internal docs links and keep them inside the
     previewer, resolving each target against the ?url= source. Local/file
     serving is unaffected (links stay relative). */
  var isPreview = location.hostname === 'html-preview.github.io';
  var docsRoot = null;
  if (isPreview) {
    var src = new URLSearchParams(location.search).get('url');
    if (src) {
      var srcBlob = src.replace('raw.githubusercontent.com/dracory/auth/', 'github.com/dracory/auth/blob/');
      // directory of the current doc, normalized to the docs/ root so
      // docs-root-relative NAV paths resolve correctly from any page
      docsRoot = srcBlob.replace(/[^/]+$/, '');
      docsRoot = docsRoot.replace(/peers\/$/, '');
    }
  }

  // docsHref maps a docs-root-relative path to either a local relative link
  // or an html-preview ?url= link (the proxy only rewrites anchors present
  // at load time — dynamically created links need the preview URL directly).
  function docsHref(rel) {
    if (docsRoot) {
      return '/?url=' + encodeURIComponent(new URL(rel, docsRoot).href);
    }
    return prefix + rel;
  }

  if (isPreview && docsRoot) {
    document.addEventListener('click', function (e) {
      var a = e.target.closest('a[href]');
      if (!a) return;
      var href = a.getAttribute('href');
      if (!href || href.charAt(0) === '#') return;
      var target;
      if (/^https?:\/\//.test(href)) {
        // already-rewritten absolute raw/github link — keep it only if it
        // points back into this docs tree
        var m = href.match(/(?:raw\.githubusercontent\.com\/dracory\/auth|github\.com\/dracory\/auth\/blob)\/([^/]+\/docs\/.+?\.html)(?:[?#].*)?$/);
        if (!m) return;
        target = 'https://github.com/dracory/auth/blob/' + m[1];
      } else {
        // relative docs link — resolve against the docs root
        var resolved = new URL(href, docsRoot);
        if (!/github\.com\/dracory\/auth\/blob\//.test(resolved.href)) return;
        target = resolved.href;
      }
      e.preventDefault();
      location.href = '/?url=' + encodeURIComponent(target);
    });
  }

  // On html-preview the location is /?url=... — the doc path comes from the
  // source URL, not location.pathname.
  var docPath = isPreview && src ? src : location.pathname;
  var docSegs = docPath.split('/').filter(Boolean);
  var inPeers = docSegs.slice(-2, -1)[0] === 'peers';
  var prefix = inPeers ? '../' : '';
  var current = docSegs[docSegs.length - 1] || 'index.html';
  var currentKey = (inPeers ? 'peers/' : '') + current;

  var NAV = [
    {
      section: 'Getting Started',
      pages: [
        ['index.html', 'Home'],
        ['overview.html', 'Overview'],
        ['project-structure.html', 'Project Structure'],
      ],
    },
    {
      section: 'Features',
      pages: [
        ['authknight.html', 'AuthKnight'],
        ['impersonation.html', 'Impersonation'],
        ['rate_limit.html', 'Rate Limiting'],
        ['ui-patterns.html', 'UI Patterns'],
      ],
    },
    {
      section: 'Comparisons',
      collapsible: true,
      pages: [
        ['peers/index.html', 'Overview'],
        ['peers/auth0.html', 'vs Auth0'],
        ['peers/authelia.html', 'vs Authelia'],
        ['peers/authentik.html', 'vs Authentik'],
        ['peers/better-auth.html', 'vs Better Auth'],
        ['peers/clerk.html', 'vs Clerk'],
        ['peers/firebase-auth.html', 'vs Firebase Auth'],
        ['peers/hanko.html', 'vs Hanko'],
        ['peers/keycloak.html', 'vs Keycloak'],
        ['peers/ory-kratos.html', 'vs Ory Kratos'],
        ['peers/supertokens.html', 'vs SuperTokens'],
        ['peers/workos.html', 'vs WorkOS'],
        ['peers/zitadel.html', 'vs Zitadel'],
      ],
    },
  ];

  /* --- Sidebar ----------------------------------------------------------- */
  var navHost = document.getElementById('docsNav');
  if (navHost) {
    var html = '<div class="docs-nav-search">' +
      '<input id="docsSearch" type="search" class="form-control form-control-sm" placeholder="Filter pages…" aria-label="Filter pages">' +
      '</div><div class="docs-nav-empty">No pages match.</div>';

    NAV.forEach(function (group, gi) {
      var anyActive = group.pages.some(function (p) { return p[0] === currentKey; });
      var collapsed = group.collapsible && !anyActive;
      html += '<div class="docs-nav-section">' +
        (group.collapsible
          ? '<button class="docs-nav-group-toggle' + (collapsed ? ' collapsed' : '') + '" data-group="' + gi + '" type="button">' +
            '<span>' + group.section + '</span><span class="chev">▾</span></button>'
          : group.section) +
        '</div>';
      html += '<div class="docs-nav-group" data-group="' + gi + '"' + (collapsed ? ' style="display:none"' : '') + '>';
      group.pages.forEach(function (p) {
        var active = p[0] === currentKey ? ' active' : '';
        html += '<a class="docs-nav-link' + active + '" href="' + docsHref(p[0]) + '" data-name="' + p[1].toLowerCase() + '">' + p[1] + '</a>';
      });
      html += '</div>';
    });
    navHost.innerHTML = html;

    // collapsible groups
    navHost.querySelectorAll('.docs-nav-group-toggle').forEach(function (btn) {
      btn.addEventListener('click', function () {
        var g = navHost.querySelector('.docs-nav-group[data-group="' + btn.dataset.group + '"]');
        var open = g.style.display !== 'none';
        g.style.display = open ? 'none' : '';
        btn.classList.toggle('collapsed', open);
      });
    });

    // live filter
    var search = document.getElementById('docsSearch');
    var emptyMsg = navHost.querySelector('.docs-nav-empty');
    search.addEventListener('input', function () {
      var q = search.value.trim().toLowerCase();
      var visible = 0;
      navHost.querySelectorAll('.docs-nav-link').forEach(function (a) {
        var show = !q || a.dataset.name.indexOf(q) !== -1;
        a.style.display = show ? '' : 'none';
        if (show) visible++;
      });
      navHost.querySelectorAll('.docs-nav-group').forEach(function (g) {
        if (q) g.style.display = '';
      });
      emptyMsg.style.display = visible ? 'none' : 'block';
    });
  }

  /* --- Mobile sidebar toggle ---------------------------------------------- */
  var sidebar = document.getElementById('docsSidebar');
  var navbarInner = document.querySelector('nav .container-fluid');
  if (sidebar && navbarInner && window.matchMedia('(max-width: 991.98px)').matches) {
    var btn = document.createElement('button');
    btn.className = 'btn btn-outline-secondary btn-sm docs-nav-toggle me-2';
    btn.type = 'button';
    btn.innerHTML = '☰';
    btn.setAttribute('aria-label', 'Toggle docs navigation');
    navbarInner.insertBefore(btn, navbarInner.firstChild);
    var backdrop = document.createElement('div');
    backdrop.className = 'docs-backdrop';
    document.body.appendChild(backdrop);
    function closeNav() { sidebar.classList.remove('open'); backdrop.classList.remove('show'); }
    btn.addEventListener('click', function () {
      var open = sidebar.classList.toggle('open');
      backdrop.classList.toggle('show', open);
    });
    backdrop.addEventListener('click', closeNav);
    sidebar.addEventListener('click', function (e) {
      if (e.target.closest('a')) closeNav();
    });
  }

  /* --- On this page (right TOC + scrollspy) -------------------------------- */
  var main = document.querySelector('.docs-main');
  var tocHost = document.getElementById('docsToc');
  if (main && tocHost) {
    var headings = main.querySelectorAll('h2, h3');
    if (headings.length > 1) {
      var toc = '<div class="docs-toc-title">On this page</div>';
      var links = [];
      headings.forEach(function (h, i) {
        if (!h.id) {
          h.id = h.textContent.trim().toLowerCase()
            .replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '') || 'section-' + i;
        }
        toc += '<a href="#' + h.id + '" class="lvl-' + h.tagName.slice(1) + '" data-target="' + h.id + '">' +
          h.textContent.trim() + '</a>';
        links.push({ h: h });
      });
      tocHost.innerHTML = toc;
      var tocLinks = tocHost.querySelectorAll('a');
      var spy = function () {
        var best = null;
        headings.forEach(function (h) {
          if (h.getBoundingClientRect().top <= 120) best = h.id;
        });
        tocLinks.forEach(function (a) {
          a.classList.toggle('active', a.dataset.target === best);
        });
      };
      document.addEventListener('scroll', spy, { passive: true });
      spy();
    }
  }

  /* --- Heading anchor links ------------------------------------------------- */
  if (main) {
    main.querySelectorAll('h1[id], h2[id], h3[id], h4[id]').forEach(function (h) {
      var a = document.createElement('a');
      a.className = 'anchor-link';
      a.href = '#' + h.id;
      a.textContent = '#';
      a.setAttribute('aria-label', 'Link to this section');
      h.appendChild(a);
    });
    // ensure ids exist on all h2/h3 even without TOC (single heading pages)
    main.querySelectorAll('h2:not([id]), h3:not([id])').forEach(function (h, i) {
      h.id = h.textContent.trim().toLowerCase()
        .replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '') || 'section-' + i;
      var a = document.createElement('a');
      a.className = 'anchor-link';
      a.href = '#' + h.id;
      a.textContent = '#';
      h.appendChild(a);
    });
  }

  /* --- Copy buttons on code blocks ------------------------------------------ */
  document.querySelectorAll('.docs-main pre, .doc pre').forEach(function (pre) {
    var wrap = document.createElement('div');
    wrap.className = 'codeblock';
    pre.parentNode.insertBefore(wrap, pre);
    wrap.appendChild(pre);
    var btn = document.createElement('button');
    btn.type = 'button';
    btn.className = 'copy-btn';
    btn.textContent = 'copy';
    btn.addEventListener('click', function () {
      var text = pre.innerText;
      (navigator.clipboard ? navigator.clipboard.writeText(text) : Promise.reject())
        .then(function () {
          btn.textContent = 'copied!';
          btn.classList.add('copied');
          setTimeout(function () { btn.textContent = 'copy'; btn.classList.remove('copied'); }, 1600);
        })
        .catch(function () {
          var r = document.createRange(); r.selectNodeContents(pre);
          var s = getSelection(); s.removeAllRanges(); s.addRange(r);
        });
    });
    wrap.appendChild(btn);
  });
})();
