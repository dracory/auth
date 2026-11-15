# e2e/browser

Browser end-to-end tests that drive the auth UI pages in a real headless
Chrome via [chromedp](https://github.com/chromedp/chromedp) (CDP). Same
harness as `blogadmin/e2e` — chromedp has no `leakless.exe` helper, so
it avoids the Windows Defender false positive.

This package complements the HTTP-level flow tests in `e2e/`: those
exercise the API end to end via `httptest.NewRecorder` (no browser),
while these prove the server-rendered pages and their embedded JS
actually work in a browser — form validation, `fetch` wiring, token
storage, and redirects.

## How it works

- `TestMain` mounts `a.Router()` on `httptest.NewServer` (in-process,
  random port). The config's `Endpoint` is set to the live server URL so
  all `Link*` helpers and the page JS post back to the same origin.
  `UrlRedirectOnSuccess` points at a stub `/dashboard` page on the same
  mux.
- `testutils.NewUsernameAndPasswordConfigForTest` provides stubbed
  backend funcs; `FuncUserLogin` accepts `jane@example.com` /
  `long-enough-password` and rejects everything else.
- One shared browser via `chromedp.NewExecAllocator`; each test opens a
  fresh tab with a 30s deadline (`newTab`).
- Chrome binary resolution order: `CHROME_BIN` env → Windows Playwright
  `chromium_headless_shell` → WSL-side Windows shell → Linux Playwright
  shell → chromedp's default lookup.

## Run

```bash
go test -v ./e2e/browser/
```

## Gotchas learned

- Errors surface as **Notiflix toasts**, not the inline `div.alert-*`
  elements — the pages load Notiflix from the CDN and `alertShow`
  prefers it when present. Assert on `document.body.innerText` (covers
  both the toast and the fallback alert divs), never on the alert divs
  alone.
- `password-reset` without a `?t=` token renders an error card — the
  password fields only exist for a valid token.
- Successful login redirects after a **hardcoded 2s** `setTimeout` in
  `login_password.js` — sleep ≥3s before asserting the location.
- `document.body.innerText` applies CSS `text-transform` — the field
  labels use `text-uppercase`, so assert "EMAIL"/"FIRST NAME", not the
  source "Email"/"First name".
- Prefer `chromedp.WaitReady` over `WaitVisible`.
- chromedp v0.23 rewrote the API; this suite pins the stable v0.14 line
  with matching `cdproto`/`go-json-experiment` versions.
