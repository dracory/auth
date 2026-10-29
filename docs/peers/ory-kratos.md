# Feature Comparison: dracory/auth vs Ory Kratos

Comparison of `github.com/dracory/auth` with [Ory Kratos](https://www.ory.sh/kratos),
the open-source, API-first identity server written in Go.

## High-Level Differences

| Aspect | dracory/auth | Ory Kratos |
|---|---|---|
| Type | Go library embedded in your app | Standalone Go microservice + admin/public APIs |
| Language | Go | Go (service); any language consumes the APIs |
| Deployment | No extra infrastructure | Kratos service + PostgreSQL/MySQL/CockroachDB |
| Data ownership | Your database, your schema | Kratos-managed schema (JSON identity traits) |
| UI | Built-in server-rendered HTML | Bring your own; reference UIs (Ory Elements, accountexperience) |
| Model | In-process sessions/cookies | Headless flows (browser/API), self-service everything |

## Feature Matrix

| Feature | dracory/auth | Ory Kratos |
|---|---|---|
| Email/password | ✅ | ✅ |
| Passwordless (code/link) | ✅ Codes + magic links | ✅ Magic link / code strategy |
| Passkeys / WebAuthn | ❌ | ✅ |
| Social / OIDC sign-in | ❌ (only via AuthKnight) | ✅ Generic OIDC + providers |
| MFA / TOTP / lookup codes | ❌ | ✅ |
| Session management | ✅ Your store + cookies | ✅ Session tokens/cookies, refresh |
| Identity schema | ❌ Host-defined | ✅ JSON Schema per identity |
| Self-service settings/profile | ❌ | ✅ Flows API |
| Account recovery / verification | ✅ Password reset | ✅ Flows API |
| Impersonation | ✅ Built-in | ⚠️ Via admin session exchange |
| Rate limiting / lockout | ✅ Built-in | ⚠️ External (deploy-level / Ory Network) |
| Ready-to-use UI | ✅ Server-rendered | ⚠️ DIY or Ory Elements |
| Middleware | ✅ Web/API/optional | ✅ Oathkeeper / session check API |
| Ops burden | ✅ None | ❌ Extra service + DB + migrations |
| Managed cloud option | ❌ | ✅ Ory Network |

| Multiple login methods simultaneously | ✅ Composable via NewMultiMethodAuth + types.Method (otp, magiclink, password, authknight, custom methods) | {mm} |
| Remember-me (extended sessions) | ✅ Built-in opt-in session extension | {rm} |
Legend: ✅ built-in · ⚠️ partial/conditional · ❌ not available

## When to Choose dracory/auth

- Monolithic Go app wanting zero extra services and built-in UI.

## When to Choose Ory Kratos

- Microservice architecture needing centralized, API-first identity.
- You need WebAuthn/passkeys, TOTP MFA, or OIDC social login while staying open-source/self-hosted.
