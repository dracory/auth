# Feature Comparison: dracory/auth vs Clerk

Comparison of `github.com/dracory/auth` with [Clerk](https://clerk.com), a
managed auth service known for polished drop-in UI components.

## High-Level Differences

| Aspect | dracory/auth | Clerk |
|---|---|---|
| Type | Go library embedded in your app | Hosted SaaS; frontend components + backend APIs |
| Language | Go only | Best for React/Next.js; JS SDKs, limited others |
| Deployment | No extra infrastructure | Vendor cloud only |
| Data ownership | Your database, your schema | User data on Clerk's servers |
| UI | Built-in server-rendered HTML | Pre-built React components (SignIn, UserButton) |
| Pricing | Free (license) | Free tier ~10k MAU; paid beyond |

## Feature Matrix

| Feature | dracory/auth | Clerk |
|---|---|---|
| Email/password | ✅ | ✅ |
| Passwordless (OTP, magic link) | ✅ Email codes + magic links | ✅ Email/SMS OTP, magic links |
| Social / OAuth providers | ❌ (only via AuthKnight) | ✅ Many providers |
| Enterprise SSO (SAML, OIDC) | ❌ | ✅ (paid) |
| MFA / TOTP | ❌ | ✅ |
| Passkeys / WebAuthn | ❌ | ✅ |
| Session management | ✅ Your store + cookies | ✅ Managed sessions, device tracking |
| Organizations / B2B | ❌ | ✅ |
| User roles & permissions | ❌ | ✅ Org roles + custom claims |
| Impersonation | ✅ Built-in | ✅ User impersonation |
| Bot detection / CAPTCHA | ❌ | ✅ |
| Rate limiting / lockout | ✅ Built-in | ✅ Managed |
| Ready-to-use UI | ✅ Server-rendered, Bootstrap | ✅ Best-in-class React UI |
| User profile management UI | ❌ | ✅ UserProfile component |
| JSON API | ✅ Built-in | ✅ REST + SDKs |
| Middleware | ✅ Web/API/optional | ✅ Per-framework SDK middleware |
| Vendor lock-in | ✅ None | ❌ High |
| Offline / self-hosted | ✅ | ❌ |

Legend: ✅ built-in · ⚠️ partial/conditional · ❌ not available

## When to Choose dracory/auth

- Go server-rendered app; you don't want a JS-centric auth dependency.
- You want user data in your own database with zero per-MAU cost.

## When to Choose Clerk

- React/Next.js frontend and you want polished auth UI with near-zero work.
- You need social login, MFA, passkeys, or B2B organizations quickly.
- Per-MAU pricing and external data storage are acceptable.
