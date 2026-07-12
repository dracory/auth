# Impersonation Feature — Implementation Plan

**Last Updated:** 2026-07-12 (rev 2 — incorporated review feedback)

---

## Overview

Impersonation allows an administrator to "log in as" another user — creating a new auth token for the target user and browsing the application from their perspective. The admin can later "exit" impersonation to return to their original session.

This document describes the full design for integrating impersonation into `dracory/auth` as a first-class, opt-in feature.

---

## Goals

- **Seamless token swap**: Admin gets a real auth token for the target user — existing middleware and application code work unchanged.
- **Reversible**: The admin's original auth token is preserved so they can return to their own session.
- **Secure**: Permission check via callback, rate-limited, CSRF-protected, audit-logged.
- **Opt-in**: Disabled by default. Consumers must explicitly enable it and provide a permission callback.
- **Implementation-agnostic**: Uses the existing callback pattern — no assumptions about the database or session store.

---

## Non-Goals

- UI pages for impersonation (the admin panel / user list is the consumer's responsibility).
- Role-based access control (the consumer decides who can impersonate via the callback).
- Impersonation chains (admin → moderator → user). Only one level of impersonation is supported.

---

## Architecture

### Current Flow (from user's existing projects)

```
Admin (authenticated) → selects target user_id
  → new session created with target user_id
  → auth cookie overwritten with new session key
  → admin browses as target user
  → (no exit mechanism — admin must log in again)
```

### Proposed Flow

```
Admin (authenticated) → POST /auth/api/impersonate/start {user_id}
  → handler resolves adminUserID from current auth token (via FuncUserFindByAuthToken)
  → chain guard: check if current token already has an imp: mapping → reject 400 if so
  → FuncCanImpersonate(adminUserID, targetUserID) is called
  → if allowed:
      → generate new auth token via str.RandomFromGamma(32, "BCDFGHJKLMNPQRSTVXYZ")
      → save admin's current auth token via FuncTemporaryKeySet (keyed by imp: + new token)
      → call FuncUserStoreAuthToken(ctx, newToken, targetUserID, options) to create a session for the target
      → set auth cookie to newToken (if UseCookies)
      → call FuncImpersonationStart hook (audit)
      → RecordImpersonationStart observability hook
      → RecordSessionCreated(targetUserID) observability hook
      → return success JSON with token (always, matching api_login pattern)

Admin (browsing as target) → POST /auth/api/impersonate/stop
  → handler resolves current auth token (the impersonation token)
  → retrieve saved original admin token via FuncTemporaryKeyGet (keyed by imp: + current token)
  → if not found → return 400 "not currently impersonating"
  → resolve targetUserID from current token via FuncUserFindByAuthToken(currentToken)
  → resolve adminUserID from original token via FuncUserFindByAuthToken(originalToken)
  → set auth cookie back to original admin token
  → optionally invalidate impersonation token via FuncUserLogout(ctx, targetUserID, options)
  → delete the temporary key (see Temp Key Deletion contract below)
  → call FuncImpersonationStop(ctx, adminUserID, targetUserID) if configured
  → RecordImpersonationStop(adminUserID, targetUserID) observability hook
  → return success JSON
```

---

## Detailed Design

### 1. New Config Fields

Added to both `ConfigUsernameAndPassword` and `ConfigPasswordless` (in the shared section):

```go
// Impersonation
EnableImpersonation       bool
FuncCanImpersonate        func(ctx context.Context, adminUserID string, targetUserID string) (bool, error)
FuncImpersonationStart    func(ctx context.Context, adminUserID string, targetUserID string) error  // optional
FuncImpersonationStop     func(ctx context.Context, adminUserID string, targetUserID string) error  // optional
```

**Validation:**
- If `EnableImpersonation` is `true` and `FuncCanImpersonate` is `nil`, return an error from the constructor.
- `FuncImpersonationStart` and `FuncImpersonationStop` are optional (default to no-op).

### 2. New Struct Fields on `authImplementation`

```go
// ===== START: impersonation
enableImpersonation    bool
funcCanImpersonate     func(ctx context.Context, adminUserID string, targetUserID string) (bool, error)
funcImpersonationStart func(ctx context.Context, adminUserID string, targetUserID string) error
funcImpersonationStop  func(ctx context.Context, adminUserID string, targetUserID string) error
// ===== END: impersonation
```

### 3. New Context Keys

**Important — pre-existing context key inconsistency:** The codebase currently has TWO separate `AuthenticatedUserID` types:
- `auth.AuthenticatedUserID{}` in `constants.go:15` — used by `GetCurrentUserID()` in `auth_implementation.go:134`
- `types.AuthenticatedUserID{}` in `types/constants.go:3` — used by all middlewares (`api_auth_or_error_middleware.go:42`, `web_auth_or_redirect_middleware.go:45`, `web_append_user_id_if_exists_middleware.go:38`)

These are **different Go types** and thus **different context keys**. This is a pre-existing bug: middlewares set `types.AuthenticatedUserID{}` but `GetCurrentUserID` reads `auth.AuthenticatedUserID{}`, so the value is never found. The impersonation feature must not repeat this mistake.

**Decision:** Define all new impersonation context keys in `types/constants.go` only (where the middlewares look for them). Use `types.ImpersonatorUserID{}` and `types.IsImpersonating{}` consistently everywhere. Do NOT duplicate them in `constants.go`.

```go
// In types/constants.go ONLY:
type ImpersonatorUserID struct{}  // the real admin's user ID (present only during impersonation)
type IsImpersonating struct{}     // boolean flag (present only during impersonation)
```

**Bug fix (recommended):** Also consolidate `AuthenticatedUserID` to a single definition in `types/constants.go` and update `GetCurrentUserID` to use `types.AuthenticatedUserID{}`. This is a pre-existing bug that should be fixed alongside the impersonation work.

These are set in the request context by the impersonation-aware middleware so that application code can detect impersonation and access the real admin's identity.

### 4. New Interface Methods on `AuthSharedInterface`

```go
// Impersonation
IsImpersonationEnabled() bool

GetFuncCanImpersonate() func(ctx context.Context, adminUserID string, targetUserID string) (bool, error)
SetFuncCanImpersonate(fn func(ctx context.Context, adminUserID string, targetUserID string) (bool, error))

GetFuncImpersonationStart() func(ctx context.Context, adminUserID string, targetUserID string) error
SetFuncImpersonationStart(fn func(ctx context.Context, adminUserID string, targetUserID string) error)

GetFuncImpersonationStop() func(ctx context.Context, adminUserID string, targetUserID string) error
SetFuncImpersonationStop(fn func(ctx context.Context, adminUserID string, targetUserID string) error)

// Link helpers (must be on the interface, not just on the struct)
LinkApiImpersonateStart() string
LinkApiImpersonateStop() string
```

### 5. New API Endpoints

#### `POST /auth/api/impersonate/start`

**Path constant:** `PathApiImpersonateStart = "api/impersonate/start"`

**Request:**
```json
{
  "user_id": "target-user-id"
}
```

**Handler logic:**
1. Resolve the current auth token from the request (cookie or header) via `utils.AuthTokenRetrieve`.
2. Resolve `adminUserID` from the auth token via `FuncUserFindByAuthToken(ctx, authToken, options)`. If token is empty or user not found → return `401 Unauthorized` JSON.
3. Read `user_id` from the request body/form.
4. Validate `user_id` is not empty → `400 Bad Request` if missing.
5. **Chain guard:** Check if the current auth token already has an impersonation mapping by calling `FuncTemporaryKeyGet(key=impersonationKeyPrefix+authToken)`. If a mapping is found → return `400 Bad Request` ("already impersonating — stop first").
6. Call `FuncCanImpersonate(ctx, adminUserID, targetUserID)`. If not allowed → return `403 Forbidden` JSON.
7. Generate a new auth token via `str.RandomFromGamma(32, "BCDFGHJKLMNPQRSTVXYZ")` (same pattern as `api_authenticate_via_username.go:166`).
8. Save the admin's current auth token via `FuncTemporaryKeySet(key=impersonationKeyPrefix+newToken, value=authToken, expiresSeconds=int(DefaultAuthTokenExpiration.Seconds()))`.
9. Call `FuncUserStoreAuthToken(ctx, newToken, targetUserID, options)` to create a session for the target user.
10. Set the auth cookie to `newToken` (if `UseCookies` is true).
11. Call `FuncImpersonationStart(ctx, adminUserID, targetUserID)` if configured.
12. Call `RecordImpersonationStart(adminUserID, targetUserID)` on observability hooks.
13. Call `RecordSessionCreated(targetUserID)` on observability hooks (consistent with login flow in `login_with_username_and_password.go:107`).
14. Return success JSON (always include token, matching `api_login.go:72-74` pattern):
```json
{
  "status": "success",
  "message": "impersonation started",
  "data": {
    "token": "new-token"
  }
}
```

**Authentication enforcement:** The start endpoint does NOT rely on `ApiAuthOrErrorMiddleware` wrapping (which is commented out in the interface). Instead, the handler itself resolves the auth token and user ID directly, returning `401` if unauthenticated. This matches the pattern where `buildAPIRoutes` only wraps CSRF and rate-limiting, not auth middleware.

**Security:**
- CSRF protected (same as other POST endpoints).
- Rate limited (same as other API endpoints).
- Chain guard prevents nested impersonation.

#### `POST /auth/api/impersonate/stop`

**Path constant:** `PathApiImpersonateStop = "api/impersonate/stop"`

**Request:** No body required.

**Handler logic:**
1. Extract the current auth token from the request (cookie or header) via `utils.AuthTokenRetrieve`.
2. If `currentAuthToken` is empty → return `400 Bad Request` ("not currently impersonating"). This guards against looking up `imp:` with no token suffix.
3. Look up the saved original admin token via `FuncTemporaryKeyGet(key=impersonationKeyPrefix+currentAuthToken)`.
4. If not found → return `400 Bad Request` ("not currently impersonating").
5. Resolve `targetUserID` from the current impersonation token via `FuncUserFindByAuthToken(ctx, currentAuthToken, options)`.
6. Resolve `adminUserID` from the original admin token via `FuncUserFindByAuthToken(ctx, originalAdminToken, options)`. (The temp value stores the original **token**, not the admin user ID — so we must resolve it.)
7. Set the auth cookie back to the original admin token (if `UseCookies` is true).
8. Optionally invalidate the impersonation token via `FuncUserLogout(ctx, targetUserID, options)`.
9. Delete the temporary key (see **Temp Key Deletion** contract below).
10. Call `FuncImpersonationStop(ctx, adminUserID, targetUserID)` if configured.
11. Call `RecordImpersonationStop(adminUserID, targetUserID)` on observability hooks.
12. Return success JSON:
```json
{
  "status": "success",
  "message": "impersonation stopped"
}
```

**Security:**
- CSRF protected.
- Rate limited.

### 6. Temp Key Deletion Contract

The plan originally said "Delete via `FuncTemporaryKeySet(..., value="", expiresSeconds=0)`" but `expiresSeconds=0` may mean "no expiry" (i.e. permanent) in some store implementations, which would leave a stale key forever.

**Decision:** Use `FuncTemporaryKeySet(key=impersonationKeyPrefix+currentAuthToken, value="", expiresSeconds=1)` — set expiry to 1 second (effectively immediate expiration). This is safe across all store implementations.

**Future enhancement (out of scope):** Consider adding a `FuncTemporaryKeyDelete(key string) error` callback to the interface for explicit deletion. For now, the 1-second expiry approach works with the existing callback set.

### 7. New Constants

In `constants.go`:

```go
PathApiImpersonateStart string = "api/impersonate/start"
PathApiImpersonateStop  string = "api/impersonate/stop"

impersonationKeyPrefix = "imp:"  // prefix for temporary keys that store original admin tokens
```

### 8. New Link Helpers

In `internal/links/`:

```go
func ApiImpersonateStart(endpoint string) string  // returns endpoint + "/api/impersonate/start"
func ApiImpersonateStop(endpoint string) string   // returns endpoint + "/api/impersonate/stop"
```

### 9. Router Integration

**Two changes required in `router.go`:**

**9a. Path detection in `AuthHandler` (lines 41-69):** The `if/else` chain that matches URI suffixes to path constants must include the new impersonation paths, otherwise requests fall through to `notFoundHandler` (which redirects to login):

```go
} else if strings.HasSuffix(uri, PathApiImpersonateStart) {
    path = PathApiImpersonateStart
} else if strings.HasSuffix(uri, PathApiImpersonateStop) {
    path = PathApiImpersonateStop
}
```

**9b. Route registration in `buildAPIRoutes`:** Add the new routes to the `apiRoutes` slice, but **only when `enableImpersonation` is `true`**. Since `buildAPIRoutes` is called unconditionally, the guard must be inside it:

```go
if a.enableImpersonation {
    apiRoutes = append(apiRoutes, []struct{...}{
        {PathApiImpersonateStart, "impersonate_start", a.apiImpersonateStart, true},
        {PathApiImpersonateStop,  "impersonate_stop",  a.apiImpersonateStop,  true},
    }...)
}
```

### 10. Observability Hooks

Add to `ObservabilityHooks` interface:

```go
RecordImpersonationStart(adminUserID string, targetUserID string)
RecordImpersonationStop(adminUserID string, targetUserID string)
```

Update `NoopObservabilityHooks`:

```go
func (NoopObservabilityHooks) RecordImpersonationStart(string, string) {}
func (NoopObservabilityHooks) RecordImpersonationStop(string, string)  {}
```

### 11. Impersonation-Aware Middleware (Optional Enhancement)

An optional middleware `ImpersonationAwareMiddleware` that:

1. Runs after the standard auth middleware (which sets `AuthenticatedUserID`).
2. Checks if the current auth token has a saved original admin token (via `FuncTemporaryKeyGet`).
3. If yes:
   - Sets `ImpersonatorUserID` in context to the admin's user ID (resolved from the original token).
   - Sets `IsImpersonating` in context to `true`.
4. Passes the request to the next handler.

This middleware is **optional** — consumers who don't need impersonation context awareness can skip it. It's primarily useful for:
- Audit logging in application handlers.
- Showing a "You are impersonating user X. Click here to exit." banner.
- Restricting certain actions during impersonation (e.g., password changes, deletions).

### 12. Token Storage Strategy

The key insight: we use the existing `FuncTemporaryKeySet` / `FuncTemporaryKeyGet` callbacks to store the mapping:

```
Key:   "imp:<impersonationToken>"
Value: <originalAdminAuthToken>
```

This means:
- No new callbacks are needed for storing the original token.
- The temporary key store already has expiration support.
- When the impersonation token expires, the mapping also expires (cleanup is automatic).

### 13. Helper Functions

In `global_helpers.go` (or a new `impersonation_helpers.go`):

```go
// IsImpersonating checks if the current request is in an impersonation session.
func IsImpersonating(r *http.Request) bool

// GetImpersonatorUserID returns the real admin's user ID if impersonating, empty string otherwise.
func GetImpersonatorUserID(r *http.Request) string
```

---

## File Changes Summary

| File | Change |
|------|--------|
| `types/auth_interfaces.go` | Add impersonation methods + link helpers to `AuthSharedInterface` |
| `types/config_username_and_password.go` | Add impersonation config fields |
| `types/config_passwordless.go` | Add impersonation config fields |
| `types/observability.go` | Add `RecordImpersonationStart` / `RecordImpersonationStop` |
| `types/constants.go` | Add `ImpersonatorUserID` and `IsImpersonating` context keys (ONLY here, not in `constants.go`) |
| `types/messages.go` | Add `MsgImpersonationStarted`, `MsgImpersonationStopped`, `MsgNotImpersonating`, `MsgAlreadyImpersonating` |
| `constants.go` | Add `PathApiImpersonateStart`, `PathApiImpersonateStop`, `impersonationKeyPrefix` |
| `auth_implementation.go` | Add impersonation struct fields, getters/setters, `IsImpersonationEnabled()`, link helper methods |
| `auth_implementation_api.go` | Add `apiImpersonateStart` / `apiImpersonateStop` thin wrapper methods (matching existing pattern) |
| `new_username_and_password_auth.go` | Wire impersonation config fields, add validation |
| `new_passwordless_auth.go` | Wire impersonation config fields, add validation |
| `router.go` | Add path detection in `AuthHandler` if/else chain + conditional route registration in `buildAPIRoutes` |
| `internal/links/links.go` | Add `ApiImpersonateStart` / `ApiImpersonateStop` link helpers |
| `internal/api/api_impersonate_start/` | New package — `dependencies.go`, `api_impersonate_start.go`, `api_impersonate_start_test.go` (follows existing `Dependencies` + `WithAuth` pattern) |
| `internal/api/api_impersonate_stop/` | New package — `dependencies.go`, `api_impersonate_stop.go`, `api_impersonate_stop_test.go` (follows existing `Dependencies` + `WithAuth` pattern) |
| `global_helpers.go` | Add `IsImpersonating` / `GetImpersonatorUserID` helpers |
| `middlewares/impersonation_aware_middleware.go` | New — optional middleware for impersonation context |
| `middlewares/impersonation_aware_middleware_test.go` | New — tests for the middleware |

---

## Testing Plan

### Unit Tests

- **`api_impersonate_start_test.go`**:
  - Happy path: admin impersonates target user successfully.
  - No auth token / unauthenticated → 401.
  - Missing `user_id` → 400.
  - `FuncCanImpersonate` returns false → 403.
  - `FuncCanImpersonate` returns error → 500.
  - `FuncUserStoreAuthToken` fails → 500.
  - `FuncTemporaryKeySet` fails → 500.
  - Chain guard: already impersonating → 400.
  - Impersonation disabled → 404 (route not registered).
  - CSRF token missing → 403.
  - Rate limit exceeded → 429.
  - Token is always returned in JSON (not conditional on `UseLocalStorage`).
  - `RecordSessionCreated` is called with `targetUserID`.

- **`api_impersonate_stop_test.go`**:
  - Happy path: stop impersonation, original token restored, both `adminUserID` and `targetUserID` resolved correctly.
  - Empty auth token (no cookie/header) → 400 "not currently impersonating".
  - Not currently impersonating (token has no imp: mapping) → 400.
  - `FuncTemporaryKeyGet` fails → 500.
  - `FuncUserFindByAuthToken` fails for original token → 500.
  - CSRF token missing → 403.
  - Rate limit exceeded → 429.

- **`impersonation_aware_middleware_test.go`**:
  - Impersonation active → context has `ImpersonatorUserID` and `IsImpersonating`.
  - Not impersonating → context does not have impersonation keys.
  - No auth token → passes through without impersonation context.

- **`global_helpers_test.go`** (additions):
  - `IsImpersonating` returns true/false correctly.
  - `GetImpersonatorUserID` returns correct ID or empty string.

- **`auth_implementation_api_test.go`** (additions):
  - Smoke test: `apiImpersonateStart` wrapper calls the internal package correctly.
  - Smoke test: `apiImpersonateStop` wrapper calls the internal package correctly.

- **`links_test.go`** (new or additions to existing link tests):
  - `ApiImpersonateStart` returns correct URL.
  - `ApiImpersonateStop` returns correct URL.

- **`new_username_and_password_auth_test.go`** (additions):
  - `EnableImpersonation=true` with `FuncCanImpersonate=nil` → constructor error.
  - `EnableImpersonation=true` with `FuncCanImpersonate` set → constructor succeeds.
  - `EnableImpersonation=false` → `IsImpersonationEnabled()` returns false.

- **`new_passwordless_auth_test.go`** (additions):
  - Same as above for passwordless config.

- **`observability_test.go`** (additions):
  - Verify `RecordImpersonationStart` / `RecordImpersonationStop` are called.
  - Verify `NoopObservabilityHooks` implements the new methods.

### Integration Tests

- Full flow: admin logs in → starts impersonation → browses as target → stops impersonation → back as admin.
- Impersonation with cookies vs localStorage.
- Impersonation token expiration cleanup.

---

## Security Considerations

1. **Permission check is mandatory**: `FuncCanImpersonate` must be provided when impersonation is enabled. The library never assumes any user can impersonate.
2. **CSRF protection**: Both start and stop endpoints are CSRF-protected (same as other POST endpoints).
3. **Rate limiting**: Both endpoints are rate-limited to prevent abuse.
4. **Token expiration**: The impersonation mapping expires with `DefaultAuthTokenExpiration` (2 hours). If the admin doesn't exit impersonation within that window, the original token mapping is lost and they must log in again.
5. **No privilege escalation**: The impersonated session is a regular user session — it has exactly the permissions of the target user, no more.
6. **Audit trail**: `FuncImpersonationStart` / `FuncImpersonationStop` hooks and observability hooks provide a complete audit trail.
7. **One level only**: The design does not support impersonation chains. The start handler enforces this via a chain guard — it checks if the current auth token already has an `imp:` mapping in the temporary key store. If so, it rejects with `400 Bad Request` ("already impersonating — stop first").

---

## Open Questions

1. **Should the impersonation token be distinguishable from a regular token?** Currently no — it's a regular token stored via `FuncUserStoreAuthToken`. This means the session store can't tell impersonation sessions apart from real ones. If audit logging at the session level is needed, we could add a `FuncUserStoreAuthToken` variant that accepts metadata. For now, the audit hooks cover this.

2. **Should we invalidate the impersonation token on stop?** The plan calls `FuncUserLogout` for the target user on stop. This is optional — if the consumer wants the impersonation session to remain valid (e.g., for audit), they can skip this. We could make this configurable.

3. **Should there be a web page endpoint (not just API)?** The current design only has API endpoints. A web page redirect flow (like the user's current controller) could be added if needed, but it's essentially a thin wrapper around the API endpoint.

---

## Internal API Package Structure

The new `internal/api/api_impersonate_start/` and `internal/api/api_impersonate_stop/` packages must follow the existing pattern used by `api_login/`, `api_register/`, etc.:

- **`dependencies.go`** — A `Dependencies` struct that aggregates all required callbacks (decoupled from `authImplementation` to avoid import cycles).
- **`api_impersonate_start.go`** (or `_stop.go`) — The core handler function that takes `Dependencies`, plus a `WithAuth` convenience wrapper that accepts `types.AuthSharedInterface` and wires the dependencies.
- **`api_impersonate_start_test.go`** (or `_stop_test.go`) — Unit tests.

Example structure for `api_impersonate_start/`:

```go
// dependencies.go
type Dependencies struct {
    CanImpersonate        func(ctx context.Context, adminUserID, targetUserID string) (bool, error)
    UserFindByAuthToken   func(ctx context.Context, token string, options types.UserAuthOptions) (string, error)
    UserStoreAuthToken    func(ctx context.Context, token, userID string, options types.UserAuthOptions) error
    TemporaryKeyGet       func(key string) (string, error)
    TemporaryKeySet       func(key string, value string, expiresSeconds int) error
    UseCookies            bool
    SetAuthCookie         func(w http.ResponseWriter, r *http.Request, token string)
    ObservabilityHooks    types.ObservabilityHooks
    ImpersonationStart    func(ctx context.Context, adminUserID, targetUserID string) error  // optional
}

// api_impersonate_start.go
func ApiImpersonateStart(w http.ResponseWriter, r *http.Request, deps Dependencies) { ... }
func ApiImpersonateStartWithAuth(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) { ... }
```

The thin wrapper in `auth_implementation_api.go` then just calls:
```go
func (a authImplementation) apiImpersonateStart(w http.ResponseWriter, r *http.Request) {
    api_impersonate_start.ApiImpersonateStartWithAuth(w, r, &a)
}
```

---

## Implementation Order

1. **Types & config** — Add context keys (`types/constants.go`), config fields (both config structs), interface methods (`types/auth_interfaces.go`), observability hooks (`types/observability.go`), message constants (`types/messages.go`).
2. **Struct & getters/setters** — Add fields to `authImplementation` (`auth_implementation.go`), implement all new interface methods including `IsImpersonationEnabled()` and link helpers.
3. **API wrappers** — Add `apiImpersonateStart` / `apiImpersonateStop` thin wrappers in `auth_implementation_api.go`.
4. **Constructors** — Wire config in both `NewUsernameAndPasswordAuth` and `NewPasswordlessAuth`, add validation (`EnableImpersonation=true` requires `FuncCanImpersonate`).
5. **Link helpers** — Add `ApiImpersonateStart` / `ApiImpersonateStop` to `internal/links/links.go`.
6. **Constants** — Add `PathApiImpersonateStart`, `PathApiImpersonateStop`, `impersonationKeyPrefix` to `constants.go`.
7. **API handlers** — Create `internal/api/api_impersonate_start/` and `internal/api/api_impersonate_stop/` packages with `Dependencies` struct + `WithAuth` wrapper.
8. **Router** — Add path detection in `AuthHandler` if/else chain + conditional route registration in `buildAPIRoutes`.
9. **Global helpers** — Add `IsImpersonating` / `GetImpersonatorUserID`.
10. **Middleware** — Create `ImpersonationAwareMiddleware`.
11. **Tests** — Write all unit and integration tests per the testing plan (TDD — write tests before or alongside each step above).
12. **Documentation** — Update README, add examples.
