---
path: modules/types.md
page-type: module
summary: Types module documentation for config structs, interfaces, constants, observability hooks, and cookie configuration.
tags: [module, types, config, interfaces, constants, observability, cookie]
created: 2026-07-16
updated: 2026-07-16
version: 1.0.0
---

# Types Module

**Package**: `types`

The types package defines all configuration structs, interfaces, constants, and shared types used across the auth library.

## Configuration Structs

### Shared Sub-Structs

| Struct | Purpose |
| :--- | :--- |
| `ConfigShared` | Common fields: endpoint, callbacks, cookies, logger, observability |
| `ConfigRateLimiting` | Rate limit settings: disable, custom func, max attempts, lockout |
| `ConfigCSRF` | CSRF settings: enable, secret |
| `ConfigImpersonation` | Impersonation settings: enable, callbacks |

### Auth-Specific Configs

| Struct | Embeds | Purpose |
| :--- | :--- | :--- |
| `ConfigPasswordless` | Shared + RateLimit + CSRF + Impersonation | Passwordless auth configuration |
| `ConfigUsernameAndPassword` | Shared + RateLimit + CSRF + Impersonation | Username/password auth configuration |
| `ConfigAuthKnight` | Shared + RateLimit + CSRF + Impersonation | AuthKnight OAuth configuration |

## Interfaces

| Interface | Extends | Key Methods |
| :--- | :--- | :--- |
| `AuthSharedInterface` | — | Router, middleware, URL helpers, accessors |
| `AuthPasswordInterface` | `AuthSharedInterface` | Password reset/restore URL helpers |
| `AuthPasswordlessInterface` | `AuthSharedInterface` | Login code verify URL helpers |
| `AuthAuthKnightInterface` | `AuthSharedInterface` | AuthKnight login/redirect/callback helpers |

## Constants (`types/constants.go`)

- **API paths**: `api/login`, `api/login-code-verify`, `api/logout`, `api/register`, `api/register-code-verify`, `api/restore-password`, `api/reset-password`, `api/impersonate/start`, `api/impersonate/stop`, `api/authknight/callback`
- **UI paths**: `login`, `login-code-verify`, `logout`, `register`, `register-code-verify`, `password-restore`, `password-reset`
- **Endpoint constants**: Used for rate limiting keys
- **Context keys**: `AuthenticatedUserID{}`, `ImpersonatorUserID{}`, `IsImpersonating{}`
- **Other**: `CookieName` = `"authtoken"`, `ImpersonationKeyPrefix` = `"imp:"`, `AuthKnightBaseURL`

## CookieConfig (`types/cookie_config.go`)

Configurable cookie settings with functional options:

| Option | Description |
| :--- | :--- |
| `WithSecure(bool)` | Set Secure flag |
| `WithHttpOnly(bool)` | Set HttpOnly flag |
| `WithSameSite(http.SameSite)` | Set SameSite policy |
| `WithMaxAge(int)` | Set max age in seconds |
| `WithDomain(string)` | Set cookie domain |
| `WithPath(string)` | Set cookie path |
| `WithCookieName(string)` | Set cookie name |
| `WithCookieConfig(CookieConfig)` | Replace entire config |

## ObservabilityHooks (`types/observability.go`)

Interface for metrics/tracing integration:

- `RecordLoginAttempt(method, success, err)`
- `RecordRegistrationAttempt(success, err)`
- `RecordRateLimitHit(endpoint, ip)`
- `RecordSessionCreated(userID)`
- `RecordPasswordReset(success, err)`
- `RecordImpersonationStart(adminUserID, targetUserID)`
- `RecordImpersonationStop(adminUserID, targetUserID)`

Default: `NoopObservabilityHooks` (no-op).

## Other Types

| Type | File | Purpose |
| :--- | :--- | :--- |
| `PasswordStrengthConfig` | `types/password.go` | Password validation rules |
| `UserAuthOptions` | `types/user_auth_options.go` | Options passed to callbacks |
| Message constants | `types/messages.go` | Centralized user-facing strings |

## See Also

- [Configuration](../configuration.md)
- [Core Module](core.md)
- [Architecture](../architecture.md)

## Changelog
- **v1.0.0** (2026-07-16): Initial creation
