---
path: modules/core.md
page-type: module
summary: Core module documentation for the root auth package, constructors, and the authImplementation struct.
tags: [module, core, auth, constructors, router, authImplementation]
created: 2025-12-06
updated: 2026-07-16
version: 2.0.0
---

# Core Module

**Package**: `auth` (Root)

The core module provides the main entry points for the library and orchestrates the interaction between the API, UI, and your configuration.

## Key Components

### 1. Factories / Constructors

*   `NewPasswordlessAuth(config types.ConfigPasswordless) (AuthPasswordlessInterface, error)`
*   `NewUsernameAndPasswordAuth(config types.ConfigUsernameAndPassword) (AuthPasswordInterface, error)`
*   `NewAuthKnightAuth(config types.ConfigAuthKnight) (AuthAuthKnightInterface, error)`

These functions validate your configuration and return an initialized auth instance.

### 2. The `authImplementation` Struct

Defined in `auth_implementation.go`. This struct holds:
*   The Router (`*http.ServeMux`)
*   Configuration (endpoint, URLs, flags)
*   All callback functions (user lookup, token storage, email, etc.)
*   Rate limiter instance
*   CSRF settings
*   Impersonation settings
*   AuthKnight settings
*   Observability hooks
*   Logger (`*slog.Logger`)

### 3. Router

`router.go` — `AuthHandler` routes by URI suffix matching using `uriHasPathSuffix()` for path-boundary-safe matching. Routes are built dynamically in `buildAPIRoutes()` based on enabled features.

### 4. Public API Shortcuts

`public_api.go` provides backward-compatible functions:
*   `AuthCookieSet(w, r, token, opts...)`
*   `AuthCookieGet(r) string`
*   `AuthCookieRemove(w, r, opts...)`
*   `AuthTokenRetrieve(r, useCookies) string`

### 5. Impersonation Helpers

`impersonation_helpers.go`:
*   `IsImpersonating(r *http.Request) bool`
*   `GetImpersonatorUserID(r *http.Request) string`

## Usage

```go
authInstance, err := auth.NewPasswordlessAuth(config)
mux := http.NewServeMux()
mux.HandleFunc("/auth/", authInstance.Router().ServeHTTP)
```

## See Also

- [API Internals](api.md)
- [UI Internals](ui.md)
- [Types Module](types.md)
- [Configuration](../configuration.md)

## Changelog
- **v2.0.0** (2026-07-16): Added AuthKnight constructor, public API shortcuts, impersonation helpers, updated return types to interfaces
- **v1.0.0** (2025-12-04): Initial creation
