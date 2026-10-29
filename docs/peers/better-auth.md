# Feature Comparison: dracory/auth vs Better Auth

Comparison of `github.com/dracory/auth` with [Better Auth](https://www.better-auth.com),
the framework-agnostic TypeScript auth library.

## High-Level Differences

| Aspect | dracory/auth | Better Auth |
|---|---|---|
| Type | Go library embedded in your app | TypeScript library embedded in your app |
| Language | Go | TypeScript (Node/Bun/Deno) |
| Storage | Your DB via callbacks (any DB) | Your DB via adapters (Drizzle, Prisma, Kysely, MongoDB) |
| UI | Built-in server-rendered HTML | DIY frontend; plugins expose API + client SDK |
| Extensibility | Callback-based config | Plugin ecosystem |

## Feature Matrix

| Feature | dracory/auth | Better Auth |
|---|---|---|
| Email/password | ✅ | ✅ |
| Passwordless | ✅ Email codes + magic links | ✅ Magic link / OTP plugins |
| Social / OAuth providers | ❌ (only via AuthKnight) | ✅ Built-in for many providers |
| Passkeys | ❌ | ✅ Plugin |
| MFA / TOTP | ❌ | ✅ Two-factor plugin |
| Organizations / teams | ❌ | ✅ Organization plugin |
| Admin / impersonation | ✅ Built-in | ✅ Admin plugin (impersonation included) |
| Session management | ✅ Your store + cookies | ✅ DB sessions, JWT/plugin options |
| Rate limiting | ✅ Built-in | ✅ Built-in |
| Ready-to-use UI | ✅ Server-rendered | ❌ API + client only |
| Middleware | ✅ Web/API/optional | ✅ Framework integrations |
| Fits Go backends | ✅ Native | ❌ (TS only; would need separate service) |

| Multiple login methods simultaneously | ✅ Composable via NewMultiMethodAuth + types.Method (otp, magiclink, password, authknight, custom methods) | {mm} |
| Remember-me (extended sessions) | ✅ Built-in opt-in session extension | {rm} |
Legend: ✅ built-in · ⚠️ partial/conditional · ❌ not available

## When to Choose dracory/auth

- Go codebase — Better Auth does not target Go at all.
- You want ready-made server-rendered pages, not API-only primitives.

## When to Choose Better Auth

- TypeScript/Node stack wanting an embedded library with a rich plugin ecosystem
  (social login, 2FA, orgs) without paying a hosted service.
