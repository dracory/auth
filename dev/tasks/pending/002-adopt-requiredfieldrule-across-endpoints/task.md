---
created: '2026-10-30'
priority: medium
status: pending
tags:
- rules
- refactor
title: Adopt RequiredFieldRule across endpoints
---

Replace hand-coded `if x == ""` required checks with `rules.NewRequiredFieldRule(value, types.Msg...)` inside the existing `[]rule.RuleInterface` fail-fast loop.

Files:
- `internal/core/login_with_username_and_password.go` L27, L32 (email, password)
- `internal/core/register_with_username_and_password.go` L34-49 (firstName, lastName, email, password)
- `internal/api/api_register/api_register.go` L353-367 (firstName, lastName, email)
- `internal/api/api_password_restore/api_password_restore.go` L134-148 (email, firstName, lastName)
- `internal/api/api_password_reset/api_password_reset.go` L147-154 (token, password)
- `internal/api/api_magic_link_send/api_magic_link_send.go` L114 (email)
- `internal/api/api_login/login_passwordless.go` L55 (email)

Preserve existing message constants and validation order.