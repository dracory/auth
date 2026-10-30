# Changelog

## Unreleased

### Added
- Remember-me wired end-to-end: `SetRememberMeEnabled`/`SetRememberMeDays`/
  `GetSecureCookies` on the implementation populate `core.PostAuthDeps`
  (previously never set, so `PostAuthHooks.IssueRememberSession` could not
  fire). Opt-in checkbox on the built-in login pages posts `remember=true`;
  the OTP flow carries the choice through sessionStorage to code-verify.
- `middlewares.RememberMeMiddleware` promoted to the public `middlewares`
  package and upgraded: grace-window rotation (consumed remember sessions
  are shortened to `RotationGrace`, never extended), optional `UserIsActive`
  revocation, `Enabled` master switch (misuse checks still run when off),
  same-request Cookie-header rewrite so downstream auth middleware sees the
  minted session immediately. `SessionUpdater` optional interface lets
  stores shorten sessions in place; `utils.MemSessionStore` implements it.
- Page JavaScript extracted from Go string literals into `//go:embed`-ed
  `.js` files under `internal/ui/**` (see
  `dev/proposals/2026-10-30-extract-page-scripts-to-js-files.md`), with a
  `js-syntax-check` CI job running `node --check` over all of them.
- `MultiMethodAuth` now mounts `GET /logout` and `POST /api/logout`
  (previously absent — the routes 404'd under a multi-method mux).

### Fixed
- `altAuthImpl` (the multi-method alternatives decorator) hid the wrapped
  impl's capability interfaces, so `AuthPasswordInterface` assertions in
  `ApiLoginWithAuth`/`ApiPasswordRestoreWithAuth`/`ApiRegisterWithAuth`
  failed with 500. Added `types.UnwrapAuth` and `altAuthImpl.Unwrap`; the
  handlers unwrap before asserting.
- `authStore` in the shared page helpers threw on `undefined` values
  (`response.data.user` is never returned by the login API), so a
  successful login surfaced the generic error toast. It now skips
  undefined values.
- Login code verification: the segmented input was hardcoded to 6
  digit-only boxes while issued codes are 8 uppercase consonants, so a
  real code could never be entered. Now 8 alphanumeric boxes
  (auto-uppercased), matching `utils.LoginCodeLength`.

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
