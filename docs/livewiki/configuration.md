---
path: configuration.md
page-type: reference
summary: Configuration reference for all three auth flows, shared config, cookie options, rate limiting, CSRF, and impersonation.
tags: [configuration, config, auth, cookie, rate-limit, csrf, impersonation, authknight]
created: 2025-12-06
updated: 2026-07-16
version: 2.0.0
---

# Configuration

The library is configured via structs passed to `NewPasswordlessAuth`, `NewUsernameAndPasswordAuth`, or `NewAuthKnightAuth`. All three embed shared sub-structs.

## ConfigShared (embedded by all configs)

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `Endpoint` | `string` | Yes | Base URL path (e.g., `/auth`). |
| `UrlRedirectOnSuccess` | `string` | Yes | Redirect URL after successful auth. |
| `UseCookies` | `bool` | Yes | Enable cookie-based sessions. |
| `UseLocalStorage` | `bool` | No | Enable localStorage token storage. |
| `CookieConfig` | `*CookieConfig` | No | Custom cookie settings. |
| `FuncLayout` | `func(string) string` | No | Custom HTML layout wrapper. |
| `FuncTemporaryKeyGet` | `func(key) (string, error)` | Yes | Retrieve temporary keys. |
| `FuncTemporaryKeySet` | `func(key, value, expiresSec) error` | Yes | Store temporary keys. |
| `FuncUserStoreAuthToken` | `func(ctx, token, userID, opts) error` | Yes | Store auth token. |
| `FuncUserFindByAuthToken` | `func(ctx, token, opts) (userID, error)` | Yes | Find user by auth token. |
| `FuncUserLogout` | `func(ctx, userID, opts) error` | Yes | Handle logout. |
| `Logger` | `*slog.Logger` | No | Structured logger. |
| `ObservabilityHooks` | `ObservabilityHooks` | No | Metrics/tracing hooks. |
| `EnableRegistration` | `bool` | No | Enable registration routes. |

## ConfigRateLimiting (embedded by all configs)

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `DisableRateLimit` | `bool` | No | Turn off rate limiting. |
| `FuncCheckRateLimit` | `func(ip, endpoint) (bool, Duration, error)` | No | Custom rate limit function. |
| `MaxLoginAttempts` | `int` | No | Max attempts per window. Default: 5. |
| `LockoutDuration` | `time.Duration` | No | Lockout window. Default: 15m. |

## ConfigCSRF (embedded by all configs)

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `EnableCSRFProtection` | `bool` | No | Enable CSRF protection. |
| `CSRFSecret` | `string` | Yes (if CSRF enabled) | Secret for CSRF token generation. |

## ConfigImpersonation (embedded by all configs)

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `EnableImpersonation` | `bool` | No | Enable impersonation. |
| `FuncCanImpersonate` | `func(ctx, adminID, targetID) (bool, error)` | Yes (if enabled) | Check impersonation permission. |
| `FuncImpersonationStart` | `func(ctx, adminID, targetID) error` | Yes (if enabled) | Called when impersonation starts. |
| `FuncImpersonationStop` | `func(ctx, adminID, targetID) error` | Yes (if enabled) | Called when impersonation stops. |

## ConfigPasswordless

Embeds: `ConfigShared`, `ConfigRateLimiting`, `ConfigCSRF`, `ConfigImpersonation`

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `FuncUserFindByEmail` | `func(ctx, email, opts) (userID, error)` | Yes | Find user by email. |
| `FuncEmailTemplateLoginCode` | `func(ctx, email, loginLink, opts) string` | No | Custom login code email template. |
| `FuncEmailTemplateRegisterCode` | `func(ctx, email, registerLink, opts) string` | No | Custom register code email template. |
| `FuncEmailSend` | `func(ctx, email, subject, body) error` | Yes | Send email. |
| `FuncUserRegister` | `func(ctx, email, firstName, lastName, opts) error` | Yes | Register new user. |

## ConfigUsernameAndPassword

Embeds: `ConfigShared`, `ConfigRateLimiting`, `ConfigCSRF`, `ConfigImpersonation`

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `EnableVerification` | `bool` | No | Require email verification. |
| `FuncUserLogin` | `func(ctx, username, password, opts) (userID, error)` | Yes | Verify password. |
| `FuncUserRegister` | `func(ctx, username, password, firstName, lastName, opts) error` | Yes | Register new user. |
| `FuncUserFindByUsername` | `func(ctx, username, firstName, lastName, opts) (userID, error)` | Yes | Find user by username. |
| `FuncUserPasswordChange` | `func(ctx, username, newPassword, opts) error` | No | Change password. |
| `FuncEmailTemplatePasswordRestore` | `func(ctx, userID, link, opts) string` | No | Password restore email template. |
| `FuncEmailTemplateRegisterCode` | `func(ctx, userID, link, opts) string` | No | Registration code email template. |
| `FuncEmailSend` | `func(ctx, userID, subject, body) error` | Yes | Send email. |
| `PasswordStrength` | `*PasswordStrengthConfig` | No | Password strength rules. |
| `LabelUsername` | `string` | No | Custom username field label. |

## ConfigAuthKnight

Embeds: `ConfigShared`, `ConfigRateLimiting`, `ConfigCSRF`, `ConfigImpersonation`

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `FuncUserFindByEmail` | `func(ctx, email, opts) (userID, error)` | Yes | Find user by email. |
| `FuncUserRegister` | `func(ctx, email, firstName, lastName, opts) (userID, error)` | Yes | Register new user (email pre-verified). |
| `FuncRedirectURL` | `func(ctx, userID) string` | No | Custom redirect URL after auth. Falls back to `UrlRedirectOnSuccess`. |
| `HTTPTimeout` | `time.Duration` | No | Timeout for AuthKnight API calls. Default: 10s. |

## CookieConfig

| Field | Type | Description |
| :--- | :--- | :--- |
| `Name` | `string` | Cookie name. Default: `authtoken`. |
| `HttpOnly` | `string` | `httponly` or `httpwritable`. |
| `Secure` | `string` | `secure` or `insecure`. |
| `SameSite` | `http.SameSite` | SameSite policy. |
| `MaxAge` | `int` | Max age in seconds. |
| `Domain` | `string` | Cookie domain. |
| `Path` | `string` | Cookie path. |

### Functional Options

- `WithSecure(bool)`
- `WithHttpOnly(bool)`
- `WithSameSite(http.SameSite)`
- `WithMaxAge(int)`
- `WithDomain(string)`
- `WithPath(string)`
- `WithCookieName(string)`
- `WithCookieConfig(CookieConfig)`

## PasswordStrengthConfig

| Field | Type | Description |
| :--- | :--- | :--- |
| `MinLength` | `int` | Minimum password length. |
| `RequireUppercase` | `bool` | Require at least one uppercase letter. |
| `RequireLowercase` | `bool` | Require at least one lowercase letter. |
| `RequireDigit` | `bool` | Require at least one digit. |
| `RequireSpecial` | `bool` | Require at least one special character. |
| `ForbidCommonWords` | `bool` | Reject common passwords. |

## Changelog
- **v2.0.0** (2026-07-16): Added ConfigAuthKnight, ConfigShared, ConfigCSRF, ConfigImpersonation, CookieConfig, PasswordStrengthConfig, functional options
- **v1.0.0** (2025-12-04): Initial creation
