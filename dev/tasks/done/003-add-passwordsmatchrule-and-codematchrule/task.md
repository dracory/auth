---
created: '2026-10-30'
priority: medium
status: done
tags:
- rules
- security
title: Add PasswordsMatchRule and CodeMatchRule
---

Two new rules wrapping inline constant-time compares:

1. `PasswordsMatchRule` — `api_password_reset.go` L160 does `subtle.ConstantTimeCompare(password, passwordConfirm) != 1` -> `MsgPasswordsDoNotMatch`.
2. `CodeMatchRule` — both verify endpoints do `stored == "" || subtle.ConstantTimeCompare(...) != 1` -> `MsgVerificationCodeExpired` (`api_register_code_verify.go` L276 and equivalent in `api_login_code_verify.go`).

Table-driven tests for each. See dev/proposals/2026-10-30-extract-business-logic-to-rules.md