# Feature Comparison: dracory/auth vs SuperTokens

This document compares `github.com/dracory/auth` with [SuperTokens](https://supertokens.com),
a popular open-source authentication provider. It is intended to help decide when
dracory/auth is a good fit and when a full auth service is the better choice.

## High-Level Differences

| Aspect | dracory/auth | SuperTokens |
|---|---|---|
| Type | Go library embedded in your app | Standalone auth service (core) + SDKs |
| Language | Go only | SDKs for Node.js, Go, Python; core in Java |
| Deployment | No extra infrastructure | Requires running the SuperTokens core service |
| Data ownership | Your database, your schema (callback-based) | SuperTokens-managed schema in MySQL/PostgreSQL |
| UI | Built-in, server-rendered HTML pages | Pre-built React widgets; other frontends DIY |
| License | See LICENSE / LICENSE_COMMERCIAL | Apache 2.0 (core) + paid managed/enterprise tiers |

## Feature Matrix

| Feature | dracory/auth | SuperTokens |
|---|---|---|
| Email/password (username + password) | ✅ Built-in flow | ✅ |
| Passwordless (email code/link) | ✅ Built-in flow (verification code) | ✅ (email or SMS OTP / magic link) |
| Delegated/external login | ✅ AuthKnight hosted login flow | ✅ Social login (Google, GitHub, Apple, etc.) |
| Social / OAuth providers | ❌ (only via AuthKnight) | ✅ Wide provider support |
| Session management | ✅ Cookie/session via your store (Bring-your-own store; session invalidation on password reset) | ✅ Managed sessions, refresh tokens, rotating tokens |
| MFA / TOTP | ❌ | ✅ TOTP, email/SMS second factor |
| User roles & permissions | ❌ (handled by host app) | ✅ User roles recipe + permission claims |
| Account linking | ❌ | ✅ |
| Email verification | ✅ (codes via your email service) | ✅ |
| Password reset | ✅ Built-in flow | ✅ |
| Impersonation (admin "login as") | ✅ Built-in (token swap, audit hooks) | ✅ (via custom implementation / user impersonation recipe) |
| CSRF protection | ✅ `dracory/csrf` integration | ✅ Anti-CSRF on cookie auth |
| Rate limiting / brute-force lockout | ✅ Built-in per-IP/per-endpoint | ✅ Built-in throttling + anomaly detection (paid) |
| Bot detection / CAPTCHA | ❌ | ✅ (paid feature) |
| Audit logging | ✅ `log/slog` structured logging hooks | ✅ (via core logs / paid features) |
| Ready-to-use UI pages | ✅ Server-rendered, Bootstrap-styled, customizable | ✅ React components; limited for non-React stacks |
| JSON API endpoints | ✅ Built-in RESTful endpoints | ✅ Exposed via SDK on your API layer |
| Auth middleware | ✅ `WebAuthOrRedirectMiddleware`, `ApiAuthOrErrorMiddleware`, `WebAppendUserIdIfExistsMiddleware` | ✅ Session verification middleware per SDK |
| Multi-tenancy / orgs | ❌ | ✅ (paid feature) |
| SAML / enterprise SSO | ❌ | ✅ (paid feature) |
| Database flexibility | ✅ Any DB — SQL, NoSQL, in-memory (callback-based) | ⚠️ Requires MySQL/PostgreSQL for the core |
| Email provider | ✅ Bring your own (callback) | ✅ SMTP config / integrations |
| Self-hosted | ✅ It's just a library | ✅ Free self-hosted core; managed cloud paid |
| Frontend framework requirement | None — server-rendered or SPA-ready API | Best experience with React; vanilla JS possible |

Legend: ✅ built-in · ⚠️ partial/conditional · ❌ not available

## When to Choose dracory/auth

- You are building a Go application and want authentication embedded with zero
  external services.
- You want full control over your user/session storage (any database, your own schema).
- You want server-rendered auth pages that work out of the box.
- You do not need social login, MFA, SAML, or multi-tenancy.

## When to Choose SuperTokens

- You need social/OAuth login, MFA, enterprise SSO (SAML), or multi-tenancy.
- You run a polyglot stack and want one auth system serving many languages.
- You are comfortable running (and maintaining) the SuperTokens core service
  or paying for managed cloud.
- Your frontend is React and benefits from their pre-built UI components.

## Notes

- dracory/auth's AuthKnight flow provides a hosted-login alternative similar in
  spirit to delegating auth to an external provider, without running your own
  auth service.
- dracory/auth's storage is entirely callback-driven; SuperTokens owns its schema.
  Migrating to SuperTokens later means exporting users and re-mapping sessions.
