---
created: '2026-10-30'
priority: low
status: done
tags:
- rules
title: 'Add EmailAllowedRule (optional: SafeReturnPathRule)'
---

`EmailAllowedRule` — `api_magic_link_send.go` L126: `deps.EmailAllowed != nil && !deps.EmailAllowed(email)` -> `MsgInvalidCredentials`. Injected-func rule, same shape as `EmailSendThrottleRule`.

Optional `SafeReturnPathRule` — `api_magic_link_send.go` L121 silently drops unsafe `return` paths via `utils.IsSafeReturnPath`; as a rule hosts could surface a message instead. Low priority, semantics differ per endpoint.

See dev/proposals/2026-10-30-extract-business-logic-to-rules.md