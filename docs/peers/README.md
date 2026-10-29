# Peer Comparisons

Feature comparisons of `github.com/dracory/auth` against peer authentication
solutions. Each document covers a high-level positioning table, a feature
matrix, and guidance on when each option is the better fit.

| Peer | Type | Document |
|---|---|---|
| SuperTokens | Open-source auth service + SDKs | [supertokens.md](supertokens.md) |
| Auth0 (Okta) | Hosted SaaS IdP | [auth0.md](auth0.md) |
| Clerk | Hosted SaaS, component-first | [clerk.md](clerk.md) |
| Firebase Auth | Google managed service | [firebase-auth.md](firebase-auth.md) |
| Keycloak | Self-hosted IAM server | [keycloak.md](keycloak.md) |
| Ory Kratos | Open-source API-first identity service | [ory-kratos.md](ory-kratos.md) |
| Better Auth | TypeScript embedded library | [better-auth.md](better-auth.md) |
| Hanko | Open-source passkey-first auth | [hanko.md](hanko.md) |
| Authelia | Self-hosted SSO portal / proxy auth | [authelia.md](authelia.md) |
| Authentik | Self-hosted identity provider | [authentik.md](authentik.md) |
| ZITADEL | Go-based identity platform | [zitadel.md](zitadel.md) |
| WorkOS | Managed enterprise identity | [workos.md](workos.md) |

## Summary Positioning

dracory/auth's niche: **an embedded Go library** with built-in server-rendered
UI, JSON API, middleware, impersonation, and rate limiting — storing everything
in *your* database via callbacks, with zero external services and zero per-MAU
cost.

Its main gaps vs the field: social/OAuth login (except via AuthKnight), MFA/passkeys,
enterprise SSO (SAML/OIDC), multi-tenancy/organizations, and hosted user-profile UI.
