---
path: modules/impersonation.md
page-type: module
summary: Impersonation module documentation for the opt-in admin login-as feature with secure token swap.
tags: [module, impersonation, admin, token-swap, security]
created: 2026-07-16
updated: 2026-07-16
version: 1.0.0
---

# Impersonation Module

Impersonation allows an admin user to "log in as" another user via secure token swap. The original admin token is preserved for restoration.

## Configuration

Enable via `ConfigImpersonation` (embedded in all config structs):

```go
config := types.ConfigUsernameAndPassword{
    // ... core fields ...
    EnableImpersonation: true,
    FuncCanImpersonate: func(ctx context.Context, adminUserID, targetUserID string) (bool, error) {
        // Check if adminUserID is allowed to impersonate targetUserID
        return true, nil
    },
    FuncImpersonationStart: func(ctx context.Context, adminUserID, targetUserID string) error {
        // Audit log: admin started impersonating
        return nil
    },
    FuncImpersonationStop: func(ctx context.Context, adminUserID, targetUserID string) error {
        // Audit log: admin stopped impersonating
        return nil
    },
}
```

## API Endpoints

| Method | Path | Description | CSRF |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/impersonate/start` | Start impersonating target user | Yes |
| `POST` | `/api/impersonate/stop` | Stop impersonating, restore original session | Yes |

## Token Swap Mechanism

1. **Start**: Original admin token stored via `FuncTemporaryKeySet` with key `imp:<new_token>`
2. **During**: User sees the target user's session
3. **Stop**: Original admin token retrieved via `FuncTemporaryKeyGet` with key `imp:<current_token>`

The `ImpersonationKeyPrefix` = `"imp:"` constant is defined in `types/constants.go`.

## Context Keys

During impersonation, the following context keys are set:
- `ImpersonatorUserID{}` — Original admin user ID
- `IsImpersonating{}` — Boolean flag

## Helper Functions

From `impersonation_helpers.go`:
- `IsImpersonating(r *http.Request) bool` — Check if current request is impersonated
- `GetImpersonatorUserID(r *http.Request) string` — Get original admin user ID

## Middleware

`ImpersonationAwareMiddleware` (in `middlewares/`) detects impersonation state and sets context values for downstream handlers.

## Observability

- `RecordImpersonationStart(adminUserID, targetUserID)` — Called when impersonation starts
- `RecordImpersonationStop(adminUserID, targetUserID)` — Called when impersonation stops

## Security Considerations

- Admin must be authenticated before calling impersonation endpoints
- `FuncCanImpersonate` must return `true` for the admin/target pair
- CSRF protection is enforced on both endpoints
- Rate limiting applies to both endpoints
- Original token is preserved using temporary key storage

## See Also

- [Core Module](core.md)
- [API Reference](../api_reference.md)
- [Data Flow](../data_flow.md)
- [Middlewares Module](middlewares.md)

## Changelog
- **v1.0.0** (2026-07-16): Initial creation
