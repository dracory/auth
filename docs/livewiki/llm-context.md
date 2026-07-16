---
path: llm-context.md
page-type: overview
summary: Complete codebase summary of Dracory Auth optimized for LLM consumption, covering three auth flows, impersonation, observability, and all modules.
tags: [llm, context, summary, auth, go, passwordless, password, authknight, impersonation]
created: 2025-12-06
updated: 2026-07-16
version: 2.0.0
---

# LLM Context: Dracory Auth

## Project Summary

Dracory Auth is a batteries-included authentication library for Go (1.26.3+). It provides ready-to-use UI pages, JSON API endpoints, and middleware while allowing you to bring your own database and email service. It supports three authentication flows: Passwordless (email verification codes), Username/Password (traditional auth), and AuthKnight (OAuth-style hosted login).

## Key Technologies

- **Language**: Go (1.26.3+)
- **HTTP**: Standard `net/http` compatible
- **UI**: Bootstrap-styled HTML pages via `dracory/hb` HTML builder
- **API**: JSON REST endpoints via `dracory/api`
- **CSRF**: `dracory/csrf`
- **Email**: `github.com/jordan-wright/email`
- **Logging**: `log/slog`

## Three Auth Flows

1. **Passwordless** (`NewPasswordlessAuth`, `types.ConfigPasswordless`) — Email verification codes
2. **Username/Password** (`NewUsernameAndPasswordAuth`, `types.ConfigUsernameAndPassword`) — Traditional auth with password reset
3. **AuthKnight** (`NewAuthKnightAuth`, `types.ConfigAuthKnight`) — OAuth-style hosted login via AuthKnight.com

## Directory Structure

```
auth/
├── auth_implementation.go       # Core authImplementation struct
├── auth_implementation_api.go   # API handler methods
├── auth_implementation_pages.go # Page handler methods
├── auth_implementation_cookies.go # Cookie methods
├── router.go                    # AuthHandler routing
├── new_passwordless_auth.go     # Passwordless constructor
├── new_username_and_password_auth.go # Username/Password constructor
├── new_authknight_auth.go       # AuthKnight constructor
├── public_api.go                # Backward-compatible public shortcuts
├── impersonation_helpers.go     # IsImpersonating, GetImpersonatorUserID
├── constants.go                 # Re-exported path/endpoint constants
├── errors.go                    # Error definitions
├── global_helpers.go            # Global helper functions
├── types/                       # Config structs, interfaces, constants
├── middlewares/                  # Public auth middlewares
├── utils/                       # Cookies, rate limiter, password strength, scribble
├── internal/
│   ├── api/                     # API endpoint handlers (per-endpoint packages)
│   ├── core/                    # Login/register business logic
│   ├── emails/                  # Email template generators
│   ├── helpers/                 # Layout, rate limit helpers
│   ├── links/                   # URL link generation
│   ├── middlewares/             # CSRF and rate limit middleware wrappers
│   ├── testutils/               # Shared test helpers
│   └── ui/                      # HTML page handlers (per-page packages)
├── examples/                    # Example apps
└── docs/livewiki/               # This documentation
```

## Core Concepts

1. **Configuration Objects**: `ConfigPasswordless`, `ConfigUsernameAndPassword`, `ConfigAuthKnight` — all embed `ConfigShared`, `ConfigRateLimiting`, `ConfigCSRF`, `ConfigImpersonation`
2. **Callback Interface**: You provide functions for user lookup, creation, token storage, and email sending
3. **Middleware**: `WebAuthOrRedirectMiddleware`, `ApiAuthOrErrorMiddleware`, `WebAppendUserIdIfExistsMiddleware`, `ImpersonationAwareMiddleware`
4. **Three Flows**: Same library supports passwordless, traditional, and AuthKnight auth
5. **Impersonation**: Opt-in admin "login as" with secure token swap (`imp:` prefix keys)
6. **Observability**: Optional `ObservabilityHooks` for metrics/tracing
7. **Cookie Configuration**: `CookieConfig` with functional options pattern

## Interfaces

| Interface | Extends | Purpose |
|------|---------|---------|
| `AuthSharedInterface` | — | Base: routing, middleware, URL helpers, accessors |
| `AuthPasswordInterface` | `AuthSharedInterface` | Password reset/restore URL helpers |
| `AuthPasswordlessInterface` | `AuthSharedInterface` | Login code verify URL helpers |
| `AuthAuthKnightInterface` | `AuthSharedInterface` | AuthKnight login/redirect/callback helpers |

## Important Files

| File | Purpose |
|------|---------|
| `auth_implementation.go` | Core `authImplementation` struct with all config and state |
| `router.go` | `AuthHandler` routing by URI suffix matching |
| `new_passwordless_auth.go` | `NewPasswordlessAuth` constructor |
| `new_username_and_password_auth.go` | `NewUsernameAndPasswordAuth` constructor |
| `new_authknight_auth.go` | `NewAuthKnightAuth` constructor |
| `public_api.go` | `AuthCookieSet`, `AuthCookieGet`, `AuthCookieRemove`, `AuthTokenRetrieve` |
| `impersonation_helpers.go` | `IsImpersonating()`, `GetImpersonatorUserID()` |
| `types/auth_interfaces.go` | All auth interfaces |
| `types/config_shared.go` | `ConfigShared`, `ConfigRateLimiting`, `ConfigCSRF`, `ConfigImpersonation` |
| `types/config_passwordless.go` | `ConfigPasswordless` |
| `types/config_username_and_password.go` | `ConfigUsernameAndPassword` |
| `types/config_authknight.go` | `ConfigAuthKnight` |
| `types/constants.go` | Path constants, endpoint constants, context keys |
| `types/cookie_config.go` | `CookieConfig` and functional options |
| `types/observability.go` | `ObservabilityHooks` interface and `NoopObservabilityHooks` |
| `types/messages.go` | Centralized user-facing message constants |
| `types/password.go` | `PasswordStrengthConfig` |
| `utils/rate_limiter.go` | `InMemoryRateLimiter` (thread-safe) |
| `utils/password_strength.go` | Password validation logic |
| `utils/auth_cookies.go` | Cookie set/get/remove utilities |
| `utils/auth_token_retrieve.go` | Token retrieval from cookie or Bearer header |

## Route Constants (in `types/constants.go`)

**API paths**: `api/login`, `api/login-code-verify`, `api/logout`, `api/register`, `api/register-code-verify`, `api/restore-password`, `api/reset-password`, `api/impersonate/start`, `api/impersonate/stop`, `api/authknight/callback`

**UI paths**: `login`, `login-code-verify`, `logout`, `register`, `register-code-verify`, `password-restore`, `password-reset`

## Context Keys

- `AuthenticatedUserID{}` — authenticated user ID
- `ImpersonatorUserID{}` — original admin user ID during impersonation
- `IsImpersonating{}` — impersonation flag
- `CookieName` = `"authtoken"` (variable, overridable)
- `ImpersonationKeyPrefix` = `"imp:"`

## Dependencies

- `github.com/dracory/csrf` v0.2.0 — CSRF tokens
- `github.com/dracory/hb` v1.88.0 — HTML builder
- `github.com/dracory/req` v0.1.0 — Request helpers
- `github.com/dracory/str` v0.18.0 — String utilities
- `github.com/dracory/api` v1.7.0 — API response helpers
- `github.com/dracory/uncdn` v0.9.0 — CDN URL handling
- `golang.org/x/crypto` v0.54.0 — Crypto utilities
- `github.com/jordan-wright/email` — Email sending

## See Also

- [Overview](overview.md) - High-level architecture
- [Getting Started](getting_started.md) - Setup guide
- [API Reference](api_reference.md) - Endpoint documentation
- [Configuration](configuration.md) - Config struct reference
- [Architecture](architecture.md) - Design and layers
- [Data Flow](data_flow.md) - Request flow diagrams

## Changelog
- **v2.0.0** (2026-07-16): Complete rewrite with three auth flows, impersonation, observability, cookie config, updated directory structure, interfaces, dependencies
- **v1.0.0** (2025-12-06): Initial creation
