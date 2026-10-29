# Changelog

## Unreleased

### Breaking
- `types.SessionStoreInterface` now requires `NewSession() SessionInterface`
  (store mints its concrete session type). Host adapters (e.g. a
  sessionstore adapter) must add the method before upgrading.
- Code verification requires `nonce` (and `email` for login verify) —
  pre-upgrade in-flight codes will not verify.

### Added
- Nonce-bound verification codes (logincode:/registercode:), attempt
  limiting, per-email send throttling, constant-time compare.
- Password anti-enumeration via optional `FuncUserExists` + dummy bcrypt.
- `SessionInterface`/`SessionStoreInterface` + `MemSessionStore` +
  `FuncSessionCreate` (store-derived by default).
- Magic-link send/callback (`BindIP` opt-in, single-use, safe-return).
- `FuncEmailEnqueue` queue-safe delivery + `ResolveDeliveryKey`/
  `ConsumeDeliveryKey` worker helpers.
- `types.Method` + `NewMultiMethodAuth` multi-method mux.
- `LoginAlternatives` fragment + per-route `SetRateLimits`.
- `PostAuthHooks` pipeline + `IsSafeBackURL`.
- Remember-me: `IssueRememberSession`/`IssueAuthSession`,
  `RememberMeMiddleware`, logout cleanup.
