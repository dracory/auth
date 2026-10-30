# Changelog

## v0.40.0 — 2026-10-30

### Breaking
- `types.SessionStoreInterface` now requires `NewSession() SessionInterface`
  (store mints its concrete session type). Host adapters (e.g. a
  sessionstore adapter) must add the method before upgrading.
- Code verification requires `nonce` (and `email` for login verify) —
  pre-upgrade in-flight codes will not verify.
- `ConfigShared.FuncLayout` is now a full-document callback
  `func(w, r, title, body, options LayoutOptions) string` matching the
  blogadmin/shopadmin contract, replacing `func(content string) string`
  which only wrapped the inner card. Returning `""` falls back to the
  package's standalone document. `GetLayout`/`SetLayout` and
  `PageOptions.Layout` changed accordingly, and `PageRender` now takes the
  `*http.Request`.
- Built-in pages are restyled (v3): page title + muted subtitle above a
  centered `shadow-sm` card, unified footer links — `card-header`
  blocks, `btn-info`/`btn-warning` footer buttons, and `btn-success`
  primaries are gone.
- jQuery, vendored Bootstrap JS/CSS, and web.js are no longer loaded —
  `uncdn` dependency removed from `go.mod`. Page scripts are vanilla JS
  (`fetch`/`URLSearchParams`); assets come from CDN links.

### Added
- `BrandName`/`BrandTagline`/`BrandLogoURL`/`LoginAlternatives` config on
  built-in pages, with an "or" divider + outlined sibling-method buttons.
- `PageLoginWithAPI` — method pages post to their own mount path so
  secondary methods (`/login-otp`, `/login-magiclink`) work under
  `NewMultiMethodAuth`; registration routes are also mounted from the
  password impl via `registrationImpl()`.
- Password restore/reset ajax mounted at `/api/restore-password` and
  `/api/reset-password` (matching the page scripts).
- Segmented 6-digit OTP input: auto-advance, backspace nav, paste
  distribution, auto-submit, `one-time-code` autocomplete.
- `authStore`/`authTo` — byte-compatible replica of web.js credential
  storage (obfuscation key + `&&expires`), so hosts running web.js still
  read the auth credentials.
- `ensureAssetsJS` — client-side loader that injects CDN assets
  (Bootstrap CSS, bootstrap-icons, Notiflix) only when the document
  doesn't already have them; host-provided versions always win. Opt out
  entirely via `window.AUTH_ASSETS_PROVIDED`/`data-auth-assets`.
- Notiflix toasts for form feedback (`Notify.failure`/`success`), with
  graceful fallback to inline alert divs when the library is absent.

### Added (earlier in this cycle)
- Nonce-bound verification codes (logincode:/registercode:), attempt
  limiting, per-email send throttling, constant-time compare.
- Password anti-enumeration via optional `FuncUserExists` + dummy bcrypt.
- `SessionInterface`/`SessionStoreInterface` + `MemSessionStore` +
  `FuncSessionCreate` (store-derived by default).
- Magic-link send/callback (`BindIP` opt-in, single-use, safe-return).
- `FuncEmailEnqueue` queue-safe delivery + `ResolveDeliveryKey`/
  `ConsumeDeliveryKey` worker helpers.
- `types.Method` + `NewMultiMethodAuth` multi-method mux.
- `LoginAlternatives` fragment + per-route `SetRateLimits`.
- `PostAuthHooks` pipeline + `IsSafeBackURL`.
- Remember-me: `IssueRememberSession`/`IssueAuthSession`,
  `RememberMeMiddleware`, logout cleanup.
