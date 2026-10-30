# Proposal: Extract Remaining Business Logic into `internal/rules`

**Date:** 2026-10-30
**Status:** Completed
**Scope:** `internal/rules`, `internal/api/*`, `internal/core/*`

## Motivation

`internal/rules` now hosts six reusable validation rules wrapping
`github.com/dracory/rule` (`RequiredFieldRule`, `EmailFormatRule`, `NonceRule`,
`VerificationCodeRule`, `PasswordStrengthRule`, `EmailSendThrottleRule`).
However, several endpoints still hand-code the same categories of checks
inline. Moving them into rules:

- keeps security invariants (throttling, attempt limits, constant-time
  compares) named, tested, and uniform;
- lets endpoints collapse to the `[]rule.RuleInterface{...}` fail-fast loop
  already used in `api_login_code_verify.go`;
- keeps rules transport-agnostic — `internal/core` already uses them with no
  HTTP dependency.

## Proposed Changes

### 1. `CodeAttemptsLimitRule` (new) — highest value

The per-nonce attempt-limit block is duplicated almost verbatim in:

- `internal/api/api_login_code_verify/api_login_code_verify.go` (~L182–195)
- `internal/api/api_register_code_verify/api_register_code_verify.go` (~L230–240)

Both do: read counter key → `attempts >= utils.MaxCodeAttempts` → invalidate
payload + counter keys → fail with `MsgTooManyAttempts`. Fail-open on read is
intentional in both.

Proposed context (injected deps, same shape as `EmailSendThrottleContext`):

```go
type CodeAttemptsLimitContext struct {
    Get         func(key string) (string, error)
    Delete      func(key string) error
    Set         func(key, value string, ttlSeconds int) error
    PayloadKey  string
    AttemptsKey string
}
```

Kills the largest duplicated security block in the package.

### 2. Adopt `RequiredFieldRule` everywhere (cleanup, no new code)

Endpoints still hand-code `if x == ""` required checks:

| File | Field(s) |
|---|---|
| `internal/core/login_with_username_and_password.go` L27, L32 | email, password |
| `internal/core/register_with_username_and_password.go` L34–49 | firstName, lastName, email, password |
| `internal/api/api_register/api_register.go` L353–367 | firstName, lastName, email |
| `internal/api/api_password_restore/api_password_restore.go` L134–148 | email, firstName, lastName |
| `internal/api/api_password_reset/api_password_reset.go` L147–154 | token, password |
| `internal/api/api_magic_link_send/api_magic_link_send.go` L114 | email |
| `internal/api/api_login/login_passwordless.go` L55 | email |

Each becomes `rules.NewRequiredFieldRule(value, types.Msg…)` inside the
existing fail-fast loop.

### 3. `PasswordsMatchRule` (new)

`api_password_reset.go` L160 does
`subtle.ConstantTimeCompare(password, passwordConfirm) != 1` inline, failing
with `MsgPasswordsDoNotMatch`. A named rule makes the constant-time
requirement explicit and testable.

### 4. `CodeMatchRule` (new)

Both verify endpoints compare the stored code with
`stored == "" || subtle.ConstantTimeCompare(...) != 1`
(`api_register_code_verify.go` L276 and the equivalent in
`api_login_code_verify.go`). Security invariant; worth naming.

### 5. `EmailAllowedRule` (new)

`api_magic_link_send.go` L126: `deps.EmailAllowed != nil && !deps.EmailAllowed(email)`
→ `MsgInvalidCredentials`. Injected-func rule, same shape as the throttle
rule. Only one call site today, but magic-link and registration flows will
likely share the policy.

### 6. `SafeReturnPathRule` (new) — optional

`api_magic_link_send.go` L121 silently drops unsafe `return` paths via
`utils.IsSafeReturnPath`. As a rule, hosts could choose to surface a message
instead of silent drop. Low priority — semantics differ per endpoint.

## Explicitly Out of Scope

- Store lookups and `err != nil` handling — I/O, not invariants.
- Session-create / token-generation branches in `internal/core` —
  orchestration, not validation.
- `registerMap` unmarshal and type-assertion extraction — deserialization.
- `api_authknight_callback` once-token check — handler redirects instead of
  returning messages, so rule messages would be ignored.
- Impersonation guards (`api_impersonate_start.go`) — viable rules
  ("already impersonating", "can-impersonate") but single call site; revisit
  if a second consumer appears.

## Suggested Order

1. `CodeAttemptsLimitRule` — deduplicates the largest security block.
2. `RequiredFieldRule` adoption — mechanical cleanup across 7 files.
3. `PasswordsMatchRule` + `CodeMatchRule`.
4. `EmailAllowedRule` (and `SafeReturnPathRule` if desired).

Each new rule ships with a table-driven `*_test.go` matching the existing
convention.
