# Feature Comparison: dracory/auth vs Authentik

Comparison of `github.com/dracory/auth` with [Authentik](https://goauthentik.io),
the flexible self-hosted identity provider.

## High-Level Differences

| | dracory/auth | Authentik |
|---|---|---|
| Type | Go library embedded in your app | Full identity provider (Python, self-hosted) |
| Scope | Per-app authentication | IAM, federation, directory, flows |
| Deployment | No extra infrastructure | Multi-service deployment (DB, Redis, workers) |
| Admin UX | Minimal (impersonation) | Rich admin console, policies, flows |

## Feature Comparison

| | dracory/auth | Authentik |
|---|---|---|
| Email/password | ✅ | ✅ |
| Passwordless (codes/links) | ✅ | ⚠️ Via flows |
| Magic links | ✅ | ⚠️ |
| MFA / TOTP / WebAuthn | ❌ | ✅ TOTP, WebAuthn, Duo |
| SSO / OIDC / SAML / LDAP | ❌ | ✅ All |
| Federation (social, SAML IdP) | ❌ | ✅ |
| Multi-tenancy | ❌ | ✅ Brands/tenants |
| Impersonation | ✅ Built-in | ✅ Admin impersonation |
| Ops burden | ✅ None — it's a library | ❌ Significant |

Legend: ✅ built-in · ⚠️ partial/conditional · ❌ not available

## When to Choose Authentik

- Centralizing identity across many apps with SSO, federation, MFA, and
  directory management — full IAM territory.
- You need an admin console, policies, and staged enrollment/recovery
  flows.

## When to Choose dracory/auth

- One Go app, one auth layer, no infrastructure — email codes, magic
  links, password, impersonation in-process.
