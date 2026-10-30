# Proposal: Extract Page Scripts into Embedded `.js` Files

**Date:** 2026-10-30
**Status:** Proposed
**Scope:** `internal/ui/*`

## Motivation

Every built-in page keeps its JavaScript inside Go string literals
(`LoginScripts`, `LoginPasswordlessScripts`, `LoginCodeVerifyScripts`,
`RegisterPasswordlessScripts`, `RegisterUsernameAndPasswordScripts`,
`RegisterAuthKnightScripts`, `RegisterCodeVerifyScripts`,
`PasswordRestoreScripts`, `PasswordResetScripts`, `LogoutScripts`, plus the
`ensureAssetsJS`/`pageHelpersJS` shared runtime in
`internal/ui/shared/page_helpers.go`).

Inside string literals the JS is dead text: no editor highlighting, no
linting, no type checking. This has already produced real defects:

- `LoginExtras.RememberMe` was added to content but never populated — the
  opt-in checkbox silently never rendered;
- the magic-link `payload.Remember` flag was written but never forwarded —
  a dead field;
- the jQuery→vanilla migration (v0.40.0) happened entirely inside string
  literals with no syntax checking.

The legacy blueprint controllers used `app.html`/`app.js`/`app.css` files
per page and had none of these problems. The old split was abandoned when
pages became dynamically composed (branding, alternatives, multi-method
mount paths), but only the **HTML** actually needs composition — the
JavaScript is a static payload with a handful of URL substitutions.

## Proposed Changes

### 1. One `.js` file per script function, colocated with `content.go`

Each page package gains `<name>.js` files next to the Go file that owns
them, loaded via `//go:embed`:

```
internal/ui/page_login/login_passwordless.js
internal/ui/page_login/login_password.js
internal/ui/page_login_code_verify/login_code_verify.js
internal/ui/page_register/register_passwordless.js
internal/ui/page_register/register_username_password.js
internal/ui/page_register/register_authknight.js
internal/ui/page_register_code_verify/register_code_verify.js
internal/ui/page_password_restore/password_restore.js
internal/ui/page_password_reset/password_reset.js
internal/ui/page_logout/logout.js
internal/ui/shared/ensure_assets.js
internal/ui/shared/page_helpers.js
```

### 2. URL injection via token placeholders

Scripts currently interpolate endpoint URLs with Go string concatenation.
In the `.js` files they become `__TOKEN__` placeholders the function
replaces at render time:

```go
//go:embed login_password.js
var loginPasswordJS string

func LoginScripts(urlApiLogin, urlOnSuccess string) string {
    s := loginPasswordJS
    s = strings.ReplaceAll(s, "__URL_API_LOGIN__", urlApiLogin)
    s = strings.ReplaceAll(s, "__URL_ON_SUCCESS__", urlOnSuccess)
    return s
}
```

The JS keeps readable declarations (`var urlApiLogin = "__URL_API_LOGIN__";`)
so the file is still syntactically valid standalone JS.

### 3. CDN placeholders unchanged

`ensure_assets.js`/`page_helpers.js` keep their existing `BOOTSTRAP_CSS`,
`NOTIFLIX_JS`, … uppercase placeholders — the substitution contract already
exists and callers/tests stay identical.

### 4. No signature or output changes

Function names, signatures, and emitted markup stay byte-identical (modulo
incidental whitespace), so `page_*_test.go` assertions and `PageRender`
consumers are unaffected. `//go:embed` adds no dependency.

## Benefits

- JS gets editor/CI linting — the highest-defect surface in the package;
- pages keep Go-side HTML composition (`LoginExtras`, alternatives,
  registration toggles) — no templating layer added;
- diffs to page behavior review as real JS diffs;
- scripts can later gain `npx eslint`/CI checking with zero packaging work.

## Non-Goals

- HTML content builders stay in Go (`hb` composition is where the dynamic
  part lives);
- no template engine introduced;
- no behavioural change to any emitted page.
