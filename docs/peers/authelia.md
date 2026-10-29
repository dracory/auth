# Feature Comparison: dracory/auth vs Authelia

Comparison of `github.com/dracory/auth` with [Authelia](https://www.authelia.com),
the self-hosted SSO portal and reverse-proxy auth layer.

## High-Level Differences

| | dracory/auth | Authelia |
|---|---|---|
| Type | Go library embedded in your app | Identity provider / reverse-proxy auth layer |
| Primary use | Per-app login flows | SSO across many apps behind a proxy |
| Deployment | No extra infrastructure | Dedicated service + config + proxy rules |
| Architecture | Your handlers call its APIs | Requests are intercepted before reaching your app |

## Feature Comparison

| | dracory/auth | Authelia |
|---|---|---|
| Email/password | ✅ | ✅ |
| Passwordless (codes/links) | ✅ | ❌ |
| Magic links | ✅ | ❌ |
| MFA / TOTP / WebAuthn | ❌ | ✅ TOTP, WebAuthn, Duo |
| SSO across apps | ❌ | ✅ |
| OIDC provider | ❌ | ✅ |
| Reverse proxy auth | ❌ | ✅ Traefik/Nginx/Caddy/Envoy |
| Middleware | ✅ Web/API/optional | ✅ Proxy-level (all apps) |

Legend: ✅ built-in · ⚠️ partial/conditional · ❌ not available

## When to Choose Authelia

- Internal tools / homelab: protect many apps behind a reverse proxy with
  one login portal and strong MFA.
- You need SSO without touching each app's code.

## When to Choose dracory/auth

- A single Go app owning its own login UX (email codes, magic links,
  password, impersonation) with no proxy/service layer.
