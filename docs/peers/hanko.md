# Feature Comparison: dracory/auth vs Hanko

Comparison of `github.com/dracory/auth` with [Hanko](https://www.hanko.io),
the open-source passkey-first authentication system.

## High-Level Differences

| | dracory/auth | Hanko |
|---|---|---|
| Type | Go library embedded in your app | Auth service/API (self-host or Hanko Cloud) |
| Primary focus | Email codes, magic links, password | Passkeys / WebAuthn |
| Deployment | No extra infrastructure | Separate service + DB |
| Passwordless | ✅ Email codes + magic links | ✅ Passkeys, email codes |

## Feature Comparison

| | dracory/auth | Hanko |
|---|---|---|
| Email/password | ✅ | ⚠️ Optional alongside passkeys |
| Passkeys / WebAuthn | ❌ | ✅ Core focus |
| Magic links | ✅ | ⚠️ |
| MFA / TOTP | ❌ | ⚠️ Limited (passkey-centric) |
| Sessions | ✅ Your store + cookies | ✅ Service-managed |
| Middleware | ✅ Web/API/optional | ✅ API/gateway integration |
| Ops burden | ✅ None — it's a library | ⚠️ Separate service |

| Multiple login methods simultaneously | ✅ Composable via NewMultiMethodAuth + types.Method (otp, magiclink, password, authknight, custom methods) | {mm} |
| Remember-me (extended sessions) | ✅ Built-in opt-in session extension | {rm} |
Legend: ✅ built-in · ⚠️ partial/conditional · ❌ not available

## When to Choose Hanko

- Passkeys are a first-class requirement — it's built for them, we're not.
- You want a hosted or self-hosted auth service rather than embedding auth
  into a Go app.

## When to Choose dracory/auth

- Email-based passwordless (codes + magic links) or password login in a Go
  web app with zero extra infrastructure.
- You need auth inside your process, your store, your cookies.
