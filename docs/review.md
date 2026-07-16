# Code Review: AuthKnight Integration

**Date:** 2026-07-16

---

## High Severity

### 1. ✅ `PageRender` sets `WriteHeader` before `Content-Type` header

**File:** `internal/ui/shared/webpage.go:87-88`

```go
w.WriteHeader(http.StatusOK)
w.Header().Set("Content-Type", "text/html")
```

`WriteHeader` flushes the status line and all accumulated headers. Setting `Content-Type` **after** `WriteHeader` has no effect — the header won't be sent to the client. Go will fall back to content sniffing via `Write`, which may produce wrong results.

**Fix:** Swap the two lines.

---

### 2. ✅ AuthKnight register API bypasses `EnableRegistration` check

**File:** `internal/api/api_register/api_register.go:29-113`

The AuthKnight branch in `ApiRegister` never checks if registration is enabled. The page route is gated by `enableRegistration` in `router.go:118-121`, but the API route is always registered in `buildAPIRoutes`. A user can POST directly to `/auth/api/register` with an `ak_key` and register even when `EnableRegistration: false`.

**Fix:** Gate `PathApiRegister` and `PathApiRegisterCodeVerify` API routes behind `enableRegistration` in `buildAPIRoutes` (same pattern as page routes).

---

### 3. ✅ Temporary key not deleted after AuthKnight registration

**File:** `internal/api/api_register/api_register.go:42`

The `ak_key` is read from the temp key store but never deleted. The same key can be reused to register multiple accounts.

**Fix:** Delete the temp key after successful retrieval, or at minimum after successful registration.

---

### 4. ✅ `LinkAuthKnightRedirect` ignores `X-Forwarded-Proto`

**File:** `auth_implementation.go:657-664`

```go
scheme := "http"
if r.TLS != nil {
    scheme = "https"
}
```

When behind a reverse proxy (the common production setup), TLS is terminated at the proxy and `r.TLS` is `nil`. This produces `http://` callback URLs instead of `https://`, which AuthKnight will reject or redirect incorrectly.

**Fix:** Also check `X-Forwarded-Proto` header:

```go
if xfp := r.Header.Get("X-Forwarded-Proto"); xfp != "" {
    scheme = xfp
}
```

---

## Medium Severity

### 5. Excessive debug `fmt.Println`/`fmt.Printf` in production code

**File:** `internal/api/api_authknight_callback/api_authknight_callback.go` (lines 33-43, 51, 63, 69, 71-73, 84, 90, 97, 99, 101, 104, 112, 124, 126, 136, 141, 144, 152, 154, 174-175, 194, 196, 201, 209, 223, 231, 239, 247, 250, 254)

**File:** `router.go:51,93`

~35 `fmt.Println`/`fmt.Printf` statements in the callback handler alone, printing sensitive data (emails, tokens, request URIs, query params) to stdout. These should be removed or converted to structured logging via `slog`.

---

### 6. ✅ `verifyOnceToken` doesn't check HTTP status code

**File:** `internal/api/api_authknight_callback/api_authknight_callback.go:231-245`

The response status code is printed (line 231) but never validated. If AuthKnight returns a 500/403/429 with a non-JSON body (e.g., an HTML error page), the code proceeds to `json.Unmarshal` which will fail with a confusing error.

**Fix:** Check `resp.StatusCode != http.StatusOK` before parsing:

```go
if resp.StatusCode != http.StatusOK {
    return "", fmt.Errorf("authknight API returned status %d", resp.StatusCode)
}
```

---

### 7. ✅ ~~AuthKnight base URL not configurable per instance~~ (not an issue — URL is intentionally hardcoded)

**File:** `internal/api/api_authknight_callback/api_authknight_callback.go:183`

`ApiAuthKnightCallbackWithAuth` hardcodes `AuthKnightDefaultBaseURL`. The `ConfigAuthKnight` struct has no `BaseURL` field. Users cannot point to a staging/self-hosted AuthKnight instance.

**Fix:** Add a `BaseURL` field to `ConfigAuthKnight`.

---

### 8. ✅ ~~Auth token returned in JSON body even when cookies are used~~ (not an issue — consistent with existing flows)

**File:** `internal/api/api_register/api_register.go:109-111`

```go
api.Respond(w, r, api.SuccessWithData(types.MsgRegistrationSuccess, map[string]any{
    "token": token,
}))
```

When `UseCookies: true`, the token is already set as an HttpOnly cookie. Also returning it in the JSON response body exposes it to JavaScript, defeating the purpose of HttpOnly cookies.

**Fix:** Omit the token from the response when cookies are used.

---

## Low Severity

### 9. ✅ Fragile type assertions in `ApiRegisterWithAuth`

**File:** `internal/api/api_register/api_register.go:176-194`

The code uses local inline interfaces (`authKnightChecker`, `authKnightAccessor`) instead of `types.AuthAuthKnightInterface`. If `GetAuthKnightUserRegister` is renamed, the assertion silently fails and `UserRegister` stays `nil`, causing a runtime error later.

**Fix:** Use `types.AuthAuthKnightInterface` directly:

```go
if akAuth, ok := a.(types.AuthAuthKnightInterface); ok && akAuth.IsAuthKnight() {
    deps.AuthKnight = true
    deps.AuthKnightRegisterDependencies.UserRegister = akAuth.GetAuthKnightUserRegister()
    // ...
}
```

---

### 10. ✅ Redundant query extraction in `page_register.go`

**File:** `internal/ui/page_register/page_register.go:16-18`

`ak_key` is extracted from query twice (line 16 and again inside `getAuthKnightEmail` at line 78). Minor inefficiency.

---

### 11. ✅ ~~`LinkAuthKnightLogin` uses hardcoded `AuthKnightBaseURL` constant~~ (not an issue — URL is intentionally hardcoded)

**File:** `auth_implementation.go:651`

Same issue as #7. The login URL builder uses the package-level constant, not configurable per instance.

---

### 12. ✅ ~~AuthKnight logout flow untested~~ (not an issue — logout is handled by this library, not AuthKnight-specific)

**File:** `new_authknight_auth.go:171-173`

`FuncUserLogout` is validated as required, but the logout flow for AuthKnight mode isn't documented or tested. The logout page doesn't have an AuthKnight redirect guard (which is correct), but the logout API behavior in AuthKnight mode should be verified.

---

## Summary

| # | Severity | Issue |
|---|----------|-------|
| 1 | **High** ✅ | `PageRender` sets `WriteHeader` before `Content-Type` — header lost |
| 2 | **High** ✅ | AuthKnight register API bypasses `EnableRegistration` check |
| 3 | **High** ✅ | Temp key (`ak_key`) not deleted after registration — replayable |
| 4 | **High** ✅ | `LinkAuthKnightRedirect` ignores `X-Forwarded-Proto` — breaks behind proxies |
| 5 | **Medium** | ~35 `fmt.Println` debug statements leaking sensitive data in production |
| 6 | **Medium** ✅ | `verifyOnceToken` doesn't check HTTP status code before JSON parse |
| 7 | **Medium** ✅ | ~~AuthKnight base URL not configurable per instance~~ (intentionally hardcoded) |
| 8 | **Medium** ✅ | ~~Auth token returned in JSON body even when HttpOnly cookies are used~~ (consistent with existing flows) |
| 9 | **Low** ✅ | Fragile inline type assertions instead of `types.AuthAuthKnightInterface` |
| 10 | **Low** ✅ | Redundant `ak_key` query extraction |
| 11 | **Low** ✅ | ~~`LinkAuthKnightLogin` uses hardcoded base URL constant~~ (intentionally hardcoded) |
| 12 | **Low** ✅ | ~~AuthKnight logout flow untested~~ (logout is library-owned, not AuthKnight-specific) |

---

## Priority Fixes

1. **#1** — One-line swap in `webpage.go`
2. **#2** — Add registration check in AuthKnight branch of `api_register.go`
3. **#3** — Delete temp key after registration in `api_register.go`
4. **#4** — Check `X-Forwarded-Proto` in `LinkAuthKnightRedirect`
