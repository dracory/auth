---
path: modules/utils.md
page-type: module
summary: Utils module documentation for cookies, token retrieval, rate limiter, password strength, and other utilities.
tags: [module, utils, cookies, rate-limiter, password, token, scribble]
created: 2026-07-16
updated: 2026-07-16
version: 1.0.0
---

# Utils Module

**Package**: `utils`

The utils package provides shared utility functions used across the auth library.

## Files

### `auth_cookies.go`
Cookie management functions:
- `AuthCookieSet(w, r, token, opts...)` — Set auth cookie with options
- `AuthCookieGet(r) string` — Get auth cookie value
- `AuthCookieRemove(w, r, opts...)` — Remove auth cookie

### `auth_token_retrieve.go`
Token retrieval from request:
- `AuthTokenRetrieve(r, useCookies) string` — Get token from cookie or Bearer header

### `bearer_token_from_header.go`
- `BearerTokenFromHeader(r) string` — Extract Bearer token from Authorization header

### `cookies.go`
Lower-level cookie helpers used by `auth_cookies.go`.

### `rate_limiter.go`
- `InMemoryRateLimiter` — Thread-safe in-memory rate limiter
- Keyed by IP and endpoint
- Configurable max attempts and lockout duration
- Uses mutex synchronization

### `password_strength.go`
- `PasswordStrengthCheck(password, config) (bool, []string)` — Validate password against rules
- Supports: min length, uppercase, lowercase, digit, special char, common words

### `login_code.go`
- Login code generation and verification utilities
- Configurable code length and character set

### `email_validation.go`
- `EmailValidate(email) bool` — Email format validation using `net/mail`

### `scribble.go`
- JSON-based key-value storage helper
- Used for temporary key storage in examples

## See Also

- [Core Module](core.md)
- [Types Module](types.md)
- [Configuration](../configuration.md)

## Changelog
- **v1.0.0** (2026-07-16): Initial creation
