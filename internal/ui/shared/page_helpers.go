package shared

import _ "embed"

// ensureAssetsJS injects the pages' CDN dependencies (Bootstrap CSS,
// bootstrap-icons, Notiflix) only when the current document does not
// already have them — dedup happens client-side because the package
// cannot see what a host layout already rendered. Runs first, before
// pageHelpersJS and the page script. A host that bundles or otherwise
// manages its own assets can set window.AUTH_ASSETS_PROVIDED = true (or
// <html data-auth-assets="provided">) to skip the check entirely.
//
//go:embed ensure_assets.js
var ensureAssetsJS string

// pageHelpersJS is the shared vanilla-JS runtime for the built-in pages —
// injected once ahead of each page's own script so every page stays free
// of duplicated helpers.
//
// authStore/authTo replicate the storage format of the web.js client
// helper — obfuscated value + "&&expires" record — so a host app running
// web.js reads exactly what these pages write.
//
//go:embed page_helpers.js
var pageHelpersJS string
