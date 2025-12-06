---
Created: 2025-12-06
Last Updated: 2025-12-06
Version: 1.0.0
---

# Troubleshooting

## Common Issues

### 1. "no such table: users"
**Cause**: Your `FuncUserFindByEmail` or similar callback is trying to query a table that doesn't exist in your database.
**Fix**: Ensure your database schema is set up before running the auth flows.

### 2. Emails not arriving
**Cause**: The `FuncEmailSend` callback is failing or your SMTP provider is blocking requests.
**Fix**: Add logging inside your `FuncEmailSend` implementation to verify it's being called. Check your spam folder.

### 3. "CSRF Token Invalid"
**Cause**: You might be testing APIs with Postman without handling cookies properly, or mixing HTTP/HTTPS on localhost.
**Fix**: Ensure cookies are enabled in your client. If running locally, you might need to relax `Secure` cookie settings if not using HTTPS (though the library tries to handle this).

## Debugging

Enable structured logging by passing a `Logger` to the config:

```go
logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
config.Logger = logger
```

This will print detailed auth flow logs to the console.
