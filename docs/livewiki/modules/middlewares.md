---
Created: 2025-12-06
Last Updated: 2025-12-06
Version: 1.0.0
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

## Internal Middleware (`internal/middlewares/`)

Used internally by the Auth library's router.

### `RateLimitMiddleware`
*   **Purpose**: Protects auth endpoints from brute-force.
*   **Implements**: Token bucket algorithm.

### `CsrfMiddleware`
*   **Purpose**: Protects `POST` requests from Cross-Site Request Forgery.
*   **Mechanism**: Validates `X-CSRF-Token` header against the CSRF cookie.
