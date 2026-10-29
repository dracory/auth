# Feature Comparison: dracory/auth vs Auth0 (Okta)

Comparison of `github.com/dracory/auth` with [Auth0](https://auth0.com), the
managed identity platform now owned by Okta.

## High-Level Differences

| Aspect | dracory/auth | Auth0 |
|---|---|---|
| Type | Go library embedded in your app | Hosted SaaS identity provider (self-hosting not offered) |
| Language | Go only | SDKs for most languages/frameworks |
| Deployment | No extra infrastructure | Vendor cloud only; auth traffic leaves your infra |
| Data ownership | Your database, your schema | User data lives on Auth0's servers |
| UI | Built-in, server-rendered HTML pages | Universal Login (hosted) or embedded Lock/SDK |
| Pricing | Free (license) | Free tier ~7.5k MAU; steep pricing beyond |

## Feature Matrix

| Feature | dracory/auth | Auth0 |
|---|---|---|
| Email/password | ✅ | ✅ |
| Passwordless (email code/link, SMS) | ✅ Email codes + magic links | ✅ Email + SMS |
| Social / OAuth providers | ❌ (only via AuthKnight) | ✅ 50+ connections |
| Enterprise SSO (SAML, OIDC, LDAP) | ❌ | ✅ (enterprise tier) |
| MFA | ❌ | ✅ TOTP, SMS, push, WebAuthn |
| Session management | ✅ Your store + cookies | ✅ Tokens + refresh rotation |
| User roles & permissions (RBAC) | ❌ | ✅ |
| Organizations / multi-tenancy | ❌ | ✅ |
| Impersonation (admin "login as") | ✅ Built-in | ⚠️ Deprecated/limited |
| Bot detection / breached password | ❌ | ✅ (paid) |
| Anomaly detection / brute force | ✅ Rate limiting + lockout | ✅ (paid tiers) |
| Audit logs | ✅ `log/slog` hooks | ✅ (retention varies by plan) |
| Ready-to-use UI | ✅ Server-rendered pages | ✅ Hosted Universal Login |
| JSON API | ✅ Built-in endpoints | ✅ Management + Authentication APIs |
| Middleware | ✅ Web/API/optional | ✅ SDK middleware per framework |
| Data stays fully in-house | ✅ Any DB via callbacks | ⚠️ Custom DB scripts; metadata still in Auth0 |
| Vendor lock-in | ✅ None — it's your code | ❌ High |
| Works fully offline / air-gapped | ✅ | ❌ |

| Multiple login methods simultaneously | ✅ Composable via NewMultiMethodAuth + types.Method (otp, magiclink, password, authknight, custom methods) | {mm} |
| Remember-me (extended sessions) | ✅ Built-in opt-in session extension | {rm} |
Legend: ✅ built-in · ⚠️ partial/conditional · ❌ not available

## When to Choose dracory/auth

- Go application, no external dependencies, no per-MAU billing.
- Regulatory/data-residency needs — all user data stays in your DB.
- Simple email/password or passwordless flows without enterprise SSO.

## When to Choose Auth0

- You need social login, enterprise SSO/SAML, MFA, or organizations now.
- Polyglot stack where one IdP serves many apps.
- You accept vendor lock-in and per-user pricing for reduced build effort.
