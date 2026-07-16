---
path: troubleshooting.md
page-type: reference
summary: Common issues and solutions for Dracory Auth, including CSRF, impersonation, AuthKnight, and rate limiting.
tags: [troubleshooting, debugging, csrf, impersonation, authknight, rate-limit]
created: 2025-12-06
updated: 2026-07-16
version: 2.0.0
---

# Troubleshooting

## Common Issues

### 1. "no such table: users"
**Cause**: Your `FuncUserFindByEmail` or similar callback is querying a non-existent table.
**Fix**: Ensure your database schema is set up before running auth flows.

### 2. Emails not arriving
**Cause**: `FuncEmailSend` callback is failing or SMTP provider is blocking.
**Fix**: Add logging inside `FuncEmailSend`. Check spam folder.

### 3. "CSRF Token Invalid"
**Cause**: Testing APIs without cookies, or mixing HTTP/HTTPS.
**Fix**: Ensure cookies are enabled. CSRF binds to IP, User-Agent, and Path.

### 4. Impersonation not working
**Cause**: `EnableImpersonation` not set, or `FuncCanImpersonate` returns false.
**Fix**: Verify `ConfigImpersonation` fields. Ensure admin is authenticated before calling `/api/impersonate/start`.

### 5. AuthKnight callback fails
**Cause**: `once` token expired/invalid, or AuthKnight API unreachable.
**Fix**: Check `HTTPTimeout` (default 10s). Verify connectivity to `authknight.com`.

### 6. Rate limit exceeded (429)
**Cause**: Too many attempts from same IP/endpoint.
**Fix**: Increase `MaxLoginAttempts` or `LockoutDuration` in config. Use `DisableRateLimit` for testing.

## Debugging

Enable structured logging:

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
config.Logger = logger
```

## Changelog
- **v2.0.0** (2026-07-16): Added impersonation, AuthKnight, rate limit troubleshooting
- **v1.0.0** (2025-12-04): Initial creation
