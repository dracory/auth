# Proposal: Extract Reusable Config Structs

> **Status: COMPLETED** — All items implemented and tests passing (90.5% coverage).

## Summary

Currently `ConfigPasswordless` and `ConfigUsernameAndPassword` duplicate ~30 shared fields. Before adding any new auth mode (e.g. AuthKnight), we extract these into small reusable structs that can be embedded via Go struct embedding.

This is a prerequisite refactoring that must be completed first.

---

## Problem

Both `ConfigPasswordless` and `ConfigUsernameAndPassword` copy-paste the same ~30 fields:

- Endpoint, layout, temp keys, auth token store/find, logout, cookies, logger, observability
- Rate limiting (disable flag, check function, max attempts, lockout)
- CSRF protection (enable flag, secret)
- Impersonation (enable flag, can/start/stop callbacks)

Adding a third config (`ConfigAuthKnight`) without extracting these would mean a third copy.

---

## Proposed Design

### 1. New Shared Structs

In `types/config_shared.go`:

```go
// ConfigShared contains fields common to all auth modes.
type ConfigShared struct {
    EnableRegistration      bool
    Endpoint                string
    FuncLayout              func(content string) string
    FuncTemporaryKeyGet     func(key string) (value string, err error)
    FuncTemporaryKeySet     func(key string, value string, expiresSeconds int) (err error)
    FuncUserStoreAuthToken  func(ctx context.Context, sessionID string, userID string, options UserAuthOptions) error
    FuncUserFindByAuthToken func(ctx context.Context, sessionID string, options UserAuthOptions) (userID string, err error)
    FuncUserLogout          func(ctx context.Context, userID string, options UserAuthOptions) (err error)
    UrlRedirectOnSuccess    string
    UseCookies              bool
    UseLocalStorage         bool
    CookieConfig            *CookieConfig
    Logger                  *slog.Logger
    ObservabilityHooks      ObservabilityHooks
}

// ConfigRateLimiting contains rate limiting options.
type ConfigRateLimiting struct {
    DisableRateLimit   bool
    FuncCheckRateLimit func(ip string, endpoint string) (allowed bool, retryAfter time.Duration, err error)
    MaxLoginAttempts   int
    LockoutDuration    time.Duration
}

// ConfigCSRF contains CSRF protection options.
type ConfigCSRF struct {
    EnableCSRFProtection bool
    CSRFSecret           string
}

// ConfigImpersonation contains impersonation options.
type ConfigImpersonation struct {
    EnableImpersonation    bool
    FuncCanImpersonate     func(ctx context.Context, adminUserID string, targetUserID string) (bool, error)
    FuncImpersonationStart func(ctx context.Context, adminUserID string, targetUserID string) error
    FuncImpersonationStop  func(ctx context.Context, adminUserID string, targetUserID string) error
}
```

### 2. Refactor Existing Configs

```go
type ConfigPasswordless struct {
    ConfigShared
    ConfigRateLimiting
    ConfigCSRF
    ConfigImpersonation

    // Passwordless-specific
    FuncUserFindByEmail           func(ctx context.Context, email string, options UserAuthOptions) (userID string, err error)
    FuncEmailTemplateLoginCode    func(ctx context.Context, email string, logingLink string, options UserAuthOptions) string
    FuncEmailTemplateRegisterCode func(ctx context.Context, email string, registerLink string, options UserAuthOptions) string
    FuncEmailSend                 func(ctx context.Context, email string, emailSubject string, emailBody string) (err error)
    FuncUserRegister              func(ctx context.Context, email string, firstName string, lastName string, options UserAuthOptions) (err error)
}

type ConfigUsernameAndPassword struct {
    ConfigShared
    ConfigRateLimiting
    ConfigCSRF
    ConfigImpersonation

    // Username/password-specific
    EnableVerification               bool
    FuncEmailTemplatePasswordRestore func(ctx context.Context, userID string, passwordRestoreLink string, options UserAuthOptions) string
    FuncEmailTemplateRegisterCode    func(ctx context.Context, userID string, passwordRestoreLink string, options UserAuthOptions) string
    FuncEmailSend                    func(ctx context.Context, userID string, emailSubject string, emailBody string) (err error)
    FuncUserFindByUsername           func(ctx context.Context, username string, firstName string, lastName string, options UserAuthOptions) (userID string, err error)
    FuncUserLogin                    func(ctx context.Context, username string, password string, options UserAuthOptions) (userID string, err error)
    FuncUserPasswordChange           func(ctx context.Context, username string, newPassword string, options UserAuthOptions) (err error)
    FuncUserRegister                 func(ctx context.Context, username string, password string, first_name string, last_name string, options UserAuthOptions) (err error)
    PasswordStrength                 *PasswordStrengthConfig
    LabelUsername                    string
}
```

---

## Backward Compatibility

Go struct embedding promotes fields, so `config.Endpoint`, `config.CSRFSecret`, `config.FuncUserLogout`, etc. all continue to work without any changes to existing user code.

No changes to constructors (`NewPasswordlessAuth`, `NewUsernameAndPasswordAuth`), interfaces, or any other code. Only the config struct definitions change.

---

## Files to Create

| File | Purpose |
|------|---------|
| `types/config_shared.go` | `ConfigShared`, `ConfigRateLimiting`, `ConfigCSRF`, `ConfigImpersonation` structs |
| `types/config_shared_test.go` | Tests for shared config structs |

## Files to Modify

| File | Change |
|------|--------|
| `types/config_passwordless.go` | Refactor to embed `ConfigShared`, `ConfigRateLimiting`, `ConfigCSRF`, `ConfigImpersonation` |
| `types/config_passwordless_test.go` | Update tests if needed (should pass unchanged due to field promotion) |
| `types/config_username_and_password.go` | Refactor to embed shared structs |
| `types/config_username_and_password_test.go` | Update tests if needed |

---

## Testing Strategy

- **Shared config structs**: `ConfigShared`, `ConfigRateLimiting`, `ConfigCSRF`, `ConfigImpersonation` in `types/config_shared_test.go`
- **Existing configs**: Verify `ConfigPasswordless` and `ConfigUsernameAndPassword` still compile and all existing tests pass unchanged (field promotion means `config.Endpoint` etc. still works)
- **Constructors**: `NewPasswordlessAuth` and `NewUsernameAndPasswordAuth` must still work without changes
- Run full test suite: `go test -v ./...`
- Check coverage: `go test -cover ./...`

---

## Implementation Order

1. `types/config_shared.go` — extract `ConfigShared`, `ConfigRateLimiting`, `ConfigCSRF`, `ConfigImpersonation`
2. `types/config_shared_test.go` — tests for shared structs
3. `types/config_passwordless.go` — refactor to embed shared structs
4. `types/config_username_and_password.go` — refactor to embed shared structs
5. Run full test suite: `go test -v ./...`
6. Check coverage: `go test -cover ./...`
