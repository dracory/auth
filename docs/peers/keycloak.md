# Feature Comparison: dracory/auth vs Keycloak

Comparison of `github.com/dracory/auth` with [Keycloak](https://www.keycloak.org),
the open-source identity and access management server (Red Hat / CNCF).

## High-Level Differences

| Aspect | dracory/auth | Keycloak |
|---|---|---|
| Type | Go library embedded in your app | Standalone IAM server (Java) you deploy and operate |
| Language | Go only | Language-agnostic via OIDC/SAML protocols |
| Deployment | No extra infrastructure | JVM service + PostgreSQL/MySQL; clustering for HA |
| Data ownership | Your database, your schema | Keycloak-managed schema in its DB |
| UI | Built-in server-rendered HTML | Hosted login themes + Account Console |
| Protocols | Cookie/session + own endpoints | Full OIDC, OAuth 2.0, SAML 2.0 |

## Feature Matrix

| Feature | dracory/auth | Keycloak |
|---|---|---|
| Email/password | ✅ | ✅ |
| Passwordless | ✅ Email codes + magic links | ✅ WebAuthn/magic link (config) |
| Social / identity brokering | ❌ | ✅ Brokering to Google, GitHub, AD, etc. |
| Enterprise SSO (SAML, OIDC, LDAP/Kerberos) | ❌ | ✅ User federation, mappers |
| MFA / WebAuthn | ❌ | ✅ OTP, WebAuthn, conditional flows |
| SSO across many apps | ❌ Per-app sessions | ✅ Realms + single sign-on |
| Fine-grained authorization | ❌ | ✅ Roles, groups, policies, UMA |
| Organizations / multi-tenancy | ❌ | ✅ Realms / Organizations (newer) |
| Impersonation | ✅ Built-in | ✅ Admin impersonation |
| Brute-force protection | ✅ Rate limiting + lockout | ✅ |
| Ready-to-use UI | ✅ Server-rendered | ✅ Themed hosted pages |
| Middleware | ✅ Web/API/optional | ✅ Adapters / OIDC middleware |
| Ops burden | ✅ None — it's a library | ❌ Significant (JVM, DB, upgrades, HA) |
| Resource footprint | ✅ Tiny | ❌ Heavy (Java + DB) |
| Audit / events | ✅ `log/slog` hooks | ✅ Event listeners, admin events |

Legend: ✅ built-in · ⚠️ partial/conditional · ❌ not available

## When to Choose dracory/auth

- Single Go app, simple auth needs, minimal ops footprint.
- You don't need SSO across applications or SAML/OIDC federation.

## When to Choose Keycloak

- Multiple apps/services need centralized SSO, or you must speak SAML/OIDC/LDAP.
- You can operate a Java service with HA and upgrade discipline.
