# Impersonation

This document describes the **opt-in** impersonation feature in `dracory/auth`. It allows an authenticated administrator to temporarily act on behalf of another user ("login as"), then exit impersonation and return to their own session.

---

## Table of Contents

1. [Overview](#overview)
2. [Security Model](#security-model)
3. [Configuration](#configuration)
4. [API Endpoints](#api-endpoints)
5. [How It Works](#how-it-works)
6. [Impersonation-Aware Middleware](#impersonation-aware-middleware)
7. [Helper Functions](#helper-functions)
8. [Audit and Observability](#audit-and-observability)
9. [Complete Example](#complete-example)
10. [Testing](#testing)

---

## Overview

Impersonation is useful for:

- Customer support agents investigating user-reported issues.
- Administrators performing actions on behalf of another user.
- Debugging permission-specific behavior from a target user's perspective.

The feature is **disabled by default** and requires the consumer to explicitly enable it and provide a permission callback.

---

## Security Model

- **Opt-in only**: Routes are not registered unless `EnableImpersonation` is `true`.
- **Permission check required**: `FuncCanImpersonate` is mandatory when enabled. The library never decides who can impersonate.
- **One-level only**: Nested impersonation is rejected by the chain guard.
- **CSRF protected**: Both endpoints are wrapped with CSRF middleware (same as other POST endpoints).
- **Rate limited**: Both endpoints participate in the standard per-IP rate limiting.
- **Session bound**: The admin's original token is stored under a temporary key with the same expiry as a normal session.
- **Audit trail**: Optional callbacks and observability hooks capture start/stop events.

---

## Configuration

Add the following fields to either `types.ConfigUsernameAndPassword` or `types.ConfigPasswordless`:

```go
EnableImpersonation: true,
FuncCanImpersonate: func(ctx context.Context, adminUserID, targetUserID string) (bool, error) {
    // Implement your permission logic here.
    // Example: only "superadmin" role can impersonate.
    return userHasRole(adminUserID, "superadmin"), nil
},
FuncImpersonationStart: func(ctx context.Context, adminUserID, targetUserID string) error {
    // Optional: audit log, metrics, notifications
    slog.Info("impersonation started", "admin", adminUserID, "target", targetUserID)
    return nil
},
FuncImpersonationStop: func(ctx context.Context, adminUserID, targetUserID string) error {
    // Optional: audit log, metrics, notifications
    slog.Info("impersonation stopped", "admin", adminUserID, "target", targetUserID)
    return nil
},
```

Validation rule:

- If `EnableImpersonation` is `true` and `FuncCanImpersonate` is `nil`, the constructor returns an error.

---

## API Endpoints

The endpoints are registered only when impersonation is enabled.

### `POST /auth/api/impersonate/start`

Starts an impersonation session for the supplied target user.

**Request (form or JSON body):**

```json
{
  "user_id": "target-user-id"
}
```

**Behavior:**

1. Extracts the current auth token from the request (cookie or `Authorization` header).
2. Resolves `adminUserID` from the auth token via `FuncUserFindByAuthToken`. Returns `401` if token is empty or user not found.
3. Reads `user_id` from the request body/form. Returns `400` if missing.
4. **Chain guard**: Checks if the current token already has an `imp:` mapping via `FuncTemporaryKeyGet`. Returns `400` if already impersonating.
5. Calls `FuncCanImpersonate(adminUserID, targetUserID)`. Returns `403` if not allowed, `500` on error.
6. Generates a new 32-character auth token.
7. Stores the mapping `imp:<newToken> -> <adminToken>` via `FuncTemporaryKeySet` with 2-hour TTL.
8. Persists the new token via `FuncUserStoreAuthToken`. On failure, cleans up the temp key and returns `500`.
9. Sets the auth cookie to the new token (when `UseCookies` is true).
10. Calls optional `FuncImpersonationStart` callback.
11. Calls `RecordImpersonationStart` and `RecordSessionCreated` on observability hooks.
12. Returns success JSON with the new token.

**Success response:**

```json
{
  "status": "success",
  "message": "impersonation started",
  "data": {
    "token": "NEW-AUTH-TOKEN"
  }
}
```

**Error responses:**

- `401 Unauthorized` — missing or invalid admin auth token.
- `400 Bad Request` — missing `user_id` or already impersonating (chain guard).
- `403 Forbidden` — `FuncCanImpersonate` returned `false`.
- `500 Internal Server Error` — token generation, temp key storage, or `FuncUserStoreAuthToken` failure.
- `403` / `429` — CSRF or rate-limit failures from the standard middleware wrappers.

### `POST /auth/api/impersonate/stop`

Exits the current impersonation session and restores the admin's original token.

**Request:** no body required.

**Behavior:**

1. Extracts the current auth token from the request (cookie or `Authorization` header). Returns `400` if empty.
2. Looks up `imp:<currentToken>` via `FuncTemporaryKeyGet` to recover the original admin token. Returns `400` if no mapping found (not currently impersonating), `500` on lookup error.
3. Resolves `targetUserID` from the current impersonation token via `FuncUserFindByAuthToken`. Returns `500` on error or empty result.
4. Resolves `adminUserID` from the original admin token via `FuncUserFindByAuthToken`. Returns `500` on error or empty result.
5. Calls `FuncUserLogout` for the target user (if configured).
6. Restores the auth cookie to the original admin token (when `UseCookies` is true).
7. Deletes the temporary mapping by setting it to empty with 1-second expiry.
8. Calls optional `FuncImpersonationStop` callback.
9. Calls `RecordImpersonationStop` on observability hooks.
10. Returns success JSON.

**Success response:**

```json
{
  "status": "success",
  "message": "impersonation stopped"
}
```

**Error responses:**

- `400 Bad Request` — empty token or no `imp:` mapping (not currently impersonating).
- `500 Internal Server Error` — temporary key or user lookup failure.
- `403` / `429` — CSRF or rate-limit failures.

---

## How It Works

### Token Mapping

The library stores a single temporary key during an impersonation session:

```
Key:   imp:<impersonationToken>
Value: <originalAdminToken>
TTL:   DefaultAuthTokenExpiration (2 hours by default)
```

This means:

- No schema changes are required in your user store.
- The mapping is automatically cleaned up when the temporary key expires.
- The original admin token is never exposed to the target session.

### Chain Guard

When the admin is already impersonating, their current auth token has an `imp:` mapping. The start handler detects this and rejects a second impersonation attempt with `400 Bad Request` (`MsgAlreadyImpersonating`). This prevents chains such as admin → moderator → user.

### Cookie vs LocalStorage

When `UseCookies` is `true`, the start handler overwrites the auth cookie with the new target token, and the stop handler restores the original admin cookie.

When `UseLocalStorage` is `true` (i.e., `UseCookies` is `false`), the client is responsible for token management. The API response always includes `"token"` in the success payload, and the client must swap the stored token.

---

## Impersonation-Aware Middleware

The library provides an optional middleware that enriches the request context with impersonation metadata:

```go
mux.Handle("/dashboard", auth.WebAuthOrRedirectMiddleware(
    middlewares.ImpersonationAwareMiddleware(dashboardHandler, authInstance),
))
```

If the current session is an impersonation session, the middleware sets:

- `types.ImpersonatorUserID{}` — the real admin's user ID.
- `types.IsImpersonating{}` — `true`.

Use the helper functions to read these values:

```go
if auth.IsImpersonating(r) {
    adminID := auth.GetImpersonatorUserID(r)
    fmt.Fprintf(w, "You are impersonating user %s. Admin: %s", userID, adminID)
}
```

The middleware is **optional**. Standard authentication continues to work normally without it.

---

## Helper Functions

```go
// IsImpersonating reports whether the request context is in an impersonation session.
func IsImpersonating(r *http.Request) bool

// GetImpersonatorUserID returns the real admin's user ID, or empty string if not impersonating.
func GetImpersonatorUserID(r *http.Request) string
```

These read from the request context, which is populated by `ImpersonationAwareMiddleware`. The functions are safe to call even when the middleware is not installed — they return `false` and `""` respectively.

---

## Audit and Observability

### Optional Callbacks

- `FuncImpersonationStart` — called after a successful start.
- `FuncImpersonationStop` — called after a successful stop.

Use these for audit logs, external SIEM integrations, or application-specific notifications.

### Observability Hooks

If you set `ObservabilityHooks` in the config, the following methods are also called:

```go
RecordImpersonationStart(adminUserID string, targetUserID string)
RecordImpersonationStop(adminUserID string, targetUserID string)
RecordSessionCreated(targetUserID string)  // called for the new target session
```

See `types/observability.go` for the full interface.

---

## Complete Example

```go
package main

import (
    "context"
    "log/slog"
    "net/http"

    "github.com/dracory/auth"
    "github.com/dracory/auth/middlewares"
    "github.com/dracory/auth/types"
)

func canImpersonate(ctx context.Context, adminUserID, targetUserID string) (bool, error) {
    // Example: allow only superadmins, and never impersonate another admin.
    if isSuperAdmin(adminUserID) && !isAdmin(targetUserID) {
        return true, nil
    }
    return false, nil
}

func main() {
    authInstance, err := auth.NewUsernameAndPasswordAuth(types.ConfigUsernameAndPassword{
        Endpoint:             "/auth",
        UrlRedirectOnSuccess: "/dashboard",
        UseCookies:           true,
        EnableImpersonation:  true,
        FuncCanImpersonate:   canImpersonate,
        FuncImpersonationStart: func(ctx context.Context, admin, target string) error {
            slog.Info("audit: impersonation start", "admin", admin, "target", target)
            return nil
        },
        FuncImpersonationStop: func(ctx context.Context, admin, target string) error {
            slog.Info("audit: impersonation stop", "admin", admin, "target", target)
            return nil
        },
        // ... other required callbacks (FuncUserLogin, FuncUserStoreAuthToken, etc.)
    })
    if err != nil {
        panic(err)
    }

    mux := http.NewServeMux()
    mux.HandleFunc("/auth/", authInstance.Router().ServeHTTP)

    // Protected dashboard with impersonation awareness
    mux.Handle("/dashboard",
        authInstance.WebAuthOrRedirectMiddleware(
            middlewares.ImpersonationAwareMiddleware(http.HandlerFunc(dashboardHandler), authInstance),
        ),
    )

    http.ListenAndServe(":8080", mux)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
    userID := auth.GetCurrentUserID(r)
    if auth.IsImpersonating(r) {
        adminID := auth.GetImpersonatorUserID(r)
        w.Write([]byte("Viewing as user " + userID + ". Impersonating admin: " + adminID))
        return
    }
    w.Write([]byte("Welcome, user " + userID))
}
```

---

## Testing

The feature ships with dedicated unit tests:

- `internal/api/api_impersonate_start/api_impersonate_start_test.go`
- `internal/api/api_impersonate_stop/api_impersonate_stop_test.go`
- `middlewares/impersonation_aware_middleware_test.go`
- `impersonation_helpers_test.go`

Run them together with the rest of the suite:

```sh
go test ./...
```

For coverage:

```sh
go test -cover ./...
```

---

## See Also

- `docs/impersonation_plan.md` — design and implementation plan.
- `types/auth_interfaces.go` — `AuthSharedInterface` additions.
- `types/observability.go` — observability hooks.
- `types/constants.go` — `ImpersonationKeyPrefix` and context key types.
- `types/messages.go` — impersonation message constants.
