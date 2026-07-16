---
path: modules/middlewares.md
page-type: module
summary: Middleware module documentation for public and internal auth middlewares, CSRF, rate limiting, and impersonation.
tags: [module, middleware, csrf, rate-limit, auth, impersonation]
created: 2025-12-06
updated: 2026-07-16
version: 2.0.0
---

# Middlewares Module

This documentation covers both the public authentication middlewares and the internal utility middlewares.

## Public Middleware (`middlewares/`)

Exposed in the root `auth` package for user consumption.

### `WebAuthOrRedirectMiddleware`
*   **Purpose**: Protects web pages.
*   **Behavior**: Checks for valid session. If invalid, redirects to login URL.

### `ApiAuthOrErrorMiddleware`
*   **Purpose**: Protects API endpoints.
*   **Behavior**: Checks for valid session. If invalid, returns 401 JSON error.

### `WebAppendUserIdIfExistsMiddleware`
*   **Purpose**: Optional auth.
*   **Behavior**: If session exists, adds User ID to context. Does **not** block if unauthenticated.

### `ImpersonationAwareMiddleware`
*   **Purpose**: Detects impersonation state.
*   **Behavior**: Checks if the current request is from an impersonating admin. Sets `ImpersonatorUserID` and `IsImpersonating` in context for downstream handlers.

## Internal Middleware (`internal/middlewares/`)

Used internally by the Auth library's router.

### `RateLimitMiddleware` (`WithRateLimit`)
*   **Purpose**: Protects auth endpoints from brute-force.
*   **Implements**: Token bucket algorithm via `InMemoryRateLimiter`.
*   **Config**: `RateLimitConfig{Check, Endpoint}`.

### `CsrfMiddleware` (`WithCSRF`)
*   **Purpose**: Protects `POST` requests from Cross-Site Request Forgery.
*   **Mechanism**: Validates `X-CSRF-Token` header or `csrf_token` form field against CSRF cookie.
*   **Config**: `CSRFConfig{Enabled, Validate}`.
*   **Binding**: IP, User-Agent, and Path via `dracory/csrf`.

## See Also

- [Core Module](core.md)
- [Architecture](../architecture.md)
- [Security Architecture](../architecture.md#security-architecture)

## Changelog
- **v2.0.0** (2026-07-16): Added ImpersonationAwareMiddleware, config structs, binding details
- **v1.0.0** (2025-12-04): Initial creation
