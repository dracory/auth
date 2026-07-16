---
path: modules/authknight.md
page-type: module
summary: AuthKnight module documentation for the OAuth-style hosted login flow via AuthKnight.com.
tags: [module, authknight, oauth, hosted-login, callback]
created: 2026-07-16
updated: 2026-07-16
version: 1.0.0
---

# AuthKnight Module

AuthKnight provides an OAuth-style hosted login flow. Users are redirected to [AuthKnight.com](https://authknight.com) for authentication, then called back to your application.

## Constructor

```go
authInstance, err := auth.NewAuthKnightAuth(types.ConfigAuthKnight{
    Endpoint: "/auth",
    UrlRedirectOnSuccess: "/dashboard",
    UseCookies: true,
    FuncUserFindByEmail: userFindByEmail,
    FuncUserRegister: userRegister,
    FuncTemporaryKeyGet: tempKeyGet,
    FuncTemporaryKeySet: tempKeySet,
    FuncUserStoreAuthToken: storeAuthToken,
    FuncUserFindByAuthToken: findByAuthToken,
    FuncUserLogout: userLogout,
})
```

## Configuration (`types.ConfigAuthKnight`)

Embeds: `ConfigShared`, `ConfigRateLimiting`, `ConfigCSRF`, `ConfigImpersonation`

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `FuncUserFindByEmail` | `func` | Yes | Find user by email |
| `FuncUserRegister` | `func` | Yes | Register new user (email pre-verified by AuthKnight) |
| `FuncRedirectURL` | `func` | No | Custom redirect URL. Falls back to `UrlRedirectOnSuccess`. |
| `HTTPTimeout` | `time.Duration` | No | AuthKnight API timeout. Default: 10s. |

## Interface (`types.AuthAuthKnightInterface`)

Extends `AuthSharedInterface` with:
- `LinkAuthKnightLogin(backURL, nextURL) string`
- `LinkAuthKnightRedirect(r *http.Request) string`
- `LinkAuthKnightCallback() string`
- `IsAuthKnight() bool`
- `GetAuthKnightUserFindByEmail()`
- `GetAuthKnightUserRegister()`
- `GetAuthKnightRedirectURL()`
- `GetAuthKnightHTTPTimeout()`

## Flow

1. User visits login page → redirected to AuthKnight.com
2. User authenticates on AuthKnight hosted page
3. AuthKnight redirects back to `POST /api/authknight/callback`
4. Library verifies `once` token with AuthKnight API
5. Library calls `FuncUserFindByEmail` or `FuncUserRegister`
6. Library stores auth token and sets cookie
7. User redirected to success URL

## API Endpoint

| Method | Path | Description |
| :--- | :--- | :--- |
| `POST` | `/api/authknight/callback` | AuthKnight OAuth callback |

## Constants

- `PathApiAuthKnightCallback` = `"api/authknight/callback"`
- `EndpointAuthKnightCallback` = `"authknight_callback"`
- `AuthKnightBaseURL` = `"https://authknight.com"`

## Example

See `examples/authknight/` for a complete working example.

## See Also

- [Core Module](core.md)
- [API Reference](../api_reference.md)
- [Data Flow](../data_flow.md)
- [Configuration](../configuration.md)

## Changelog
- **v1.0.0** (2026-07-16): Initial creation
