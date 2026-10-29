# Feature Comparison: dracory/auth vs WorkOS

Comparison of `github.com/dracory/auth` with [WorkOS](https://workos.com),
the managed enterprise-identity platform.

## High-Level Differences

| | dracory/auth | WorkOS |
|---|---|---|
| Type | Go library embedded in your app | Managed SaaS (enterprise identity) |
| Data ownership | Your database, your schema | Vendor's platform |
| Deployment | No extra infrastructure | API dependency |
| Pricing | Free (license) | Per-MAU pricing |

## Feature Comparison

| | dracory/auth | WorkOS |
|---|---|---|
| Email/password | ✅ | ✅ |
| Passwordless (codes/links) | ✅ | ✅ |
| Magic links | ✅ | ✅ |
| MFA / TOTP | ❌ | ✅ |
| SSO (SAML/OIDC for enterprises) | ❌ | ✅ Flagship feature |
| SCIM / directory sync | ❌ | ✅ |
| Audit logs | ✅ `log/slog` hooks | ✅ Hosted |
| Impersonation | ✅ Built-in | ⚠️ Admin portal |
| Data stays in-house | ✅ Any DB | ❌ Vendor-hosted |
| Works offline / air-gapped | ✅ | ❌ |

Legend: ✅ built-in · ⚠️ partial/conditional · ❌ not available

## When to Choose WorkOS

- Selling to enterprises that demand SAML SSO, SCIM directory sync, and
  hosted audit trails — that's WorkOS's entire reason to exist.
- You'd rather outsource auth operations than maintain them.

## When to Choose dracory/auth

- Self-hosted auth inside a Go app: email codes, magic links, password,
  impersonation — no vendor, no per-MAU pricing, data never leaves.
