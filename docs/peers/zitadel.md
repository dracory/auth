# Feature Comparison: dracory/auth vs ZITADEL

Comparison of `github.com/dracory/auth` with [ZITADEL](https://zitadel.com),
the Go-based identity infrastructure platform.

## High-Level Differences

| | dracory/auth | ZITADEL |
|---|---|---|
| Type | Go library embedded in your app | Identity platform (self-host or cloud) |
| Language fit | Go, in-process | Go-based but consumed via APIs |
| Scope | Per-app auth flows | Multi-tenant IAM with organizations |
| Deployment | No extra infrastructure | Postgres + services (or cloud) |

## Feature Comparison

| | dracory/auth | ZITADEL |
|---|---|---|
| Email/password | ✅ | ✅ |
| Passwordless (codes/links) | ✅ | ✅ Passkeys + codes |
| Magic links | ✅ | ✅ |
| MFA / TOTP / WebAuthn | ❌ | ✅ TOTP, WebAuthn/passkeys |
| OIDC / OAuth2 / SAML | ❌ | ✅ All |
| Multi-tenancy / orgs | ❌ | ✅ Organizations |
| Event-sourced audit | ❌ | ✅ Event log + projections |
| Middleware | ✅ Web/API/optional | ✅ OIDC/SDK integration |
| Ops burden | ✅ None — it's a library | ❌ Significant (Postgres, services) |

Legend: ✅ built-in · ⚠️ partial/conditional · ❌ not available

## When to Choose ZITADEL

- Multi-tenant SaaS needing organizations, OIDC/OAuth2 APIs, MFA/passkeys,
  and an event-sourced audit trail in the Go ecosystem.
- You want a full identity platform with a hosted option.

## When to Choose dracory/auth

- Single-app auth inside a Go process — email codes, magic links, password,
  impersonation, your own stores — no platform to run.
