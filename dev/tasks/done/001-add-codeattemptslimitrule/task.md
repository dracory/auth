---
created: '2026-10-30'
priority: high
status: done
tags:
- rules
- refactor
- security
title: Add CodeAttemptsLimitRule
---

Extract the duplicated per-nonce attempt-limit block into a new `CodeAttemptsLimitRule` in `internal/rules`.

Duplicated in:
- `internal/api/api_login_code_verify/api_login_code_verify.go` (~L182-195)
- `internal/api/api_register_code_verify/api_register_code_verify.go` (~L230-240)

Both do: read counter key -> `attempts >= utils.MaxCodeAttempts` -> invalidate payload + counter keys -> fail `MsgTooManyAttempts`. Fail-open on read is intentional.

Suggested context:
```go
type CodeAttemptsLimitContext struct {
    Get         func(key string) (string, error)
    Delete      func(key string) error
    Set         func(key, value string, ttlSeconds int) error
    PayloadKey  string
    AttemptsKey string
}
```
Ship with table-driven `*_test.go`. See dev/proposals/2026-10-30-extract-business-logic-to-rules.md