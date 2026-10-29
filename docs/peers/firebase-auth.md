# Feature Comparison: dracory/auth vs Firebase Authentication

Comparison of `github.com/dracory/auth` with
[Firebase Authentication](https://firebase.google.com/products/auth) (Google Identity Platform).

## High-Level Differences

| Aspect | dracory/auth | Firebase Auth |
|---|---|---|
| Type | Go library embedded in your app | Managed Google service + client SDKs |
| Language | Go only | Client SDKs (JS, iOS, Android, Flutter); Admin SDKs server-side |
| Deployment | No extra infrastructure | Google cloud only |
| Data ownership | Your database, your schema | User data in Firebase/Google |
| UI | Built-in server-rendered HTML | FirebaseUI widget or DIY |
| Pricing | Free (license) | Generous free tier; SMS and Identity Platform features billed |

## Feature Matrix

| Feature | dracory/auth | Firebase Auth |
|---|---|---|
| Email/password | ✅ | ✅ |
| Passwordless (email link) | ✅ Codes + magic links | ✅ Magic links, SMS OTP |
| Phone/SMS auth | ❌ | ✅ |
| Social / OAuth providers | ❌ (only via AuthKnight) | ✅ Google, Apple, GitHub, X, etc. |
| Anonymous auth | ❌ | ✅ |
| Custom token minting | ❌ | ✅ |
| MFA | ❌ | ✅ SMS/TOTP (Identity Platform) |
| SAML / OIDC enterprise SSO | ❌ | ✅ (Identity Platform) |
| Multi-tenancy | ❌ | ✅ (Identity Platform tenants) |
| Session management | ✅ Your store + cookies | ✅ ID tokens + session cookies |
| Impersonation | ✅ Built-in | ❌ (custom claims workarounds) |
| Rate limiting / lockout | ✅ Built-in | ✅ Managed |
| Ready-to-use UI | ✅ Server-rendered | ⚠️ FirebaseUI (dated) or DIY |
| Middleware | ✅ Web/API/optional | ✅ Admin SDK token verification |
| Works without GCP | ✅ | ❌ |
| Vendor lock-in | ✅ None | ❌ High |

Legend: ✅ built-in · ⚠️ partial/conditional · ❌ not available

## When to Choose dracory/auth

- Traditional Go web app with server-rendered pages and sessions.
- You don't want a Google dependency or Firebase lock-in.

## When to Choose Firebase Auth

- Mobile-first app (iOS/Android/Flutter) needing phone auth or anonymous sign-in.
- Already invested in the Firebase/GCP ecosystem.
