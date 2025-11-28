# Rate Limiting in dracory/auth

**Last Updated:** 2025-11-28

---

## Overview

`dracory/auth` includes **built-in rate limiting** for all authentication endpoints. The goal is to protect your application from:

- Brute-force login attempts
- Credential stuffing
- Abuse of registration and password reset flows

By default, rate limiting is **enabled** and uses an **in-memory sliding window** implementation.

---

## What Is Rate Limited?

Rate limiting is applied per **IP address** and **logical endpoint** for all auth-related operations, including:

- Login (passwordless and username/password)
- Registration
- Verification-code endpoints
- Password restore / reset endpoints

When the limit is exceeded, the library returns:

- HTTP **429 Too Many Requests**
- A `Retry-After` header with the lockout duration (in seconds)
- A JSON error body: `"Too many requests. Please try again later."`

---

## Default Behavior

The default in-memory limiter is implemented in the internal `utils` package (see `utils/rate_limiter.go`) and is initialized automatically when you create an auth instance (passwordless or username/password) and **do not** provide your own rate-limiter function.

**Defaults:**

- **Max attempts:** `5`
- **Window duration:** `15 * time.Minute`
- **Lockout duration:** `15 * time.Minute`

This means:

- Each IP+endpoint pair is allowed up to **5 requests** within a rolling **15-minute window**.
- When the limit is exceeded, further requests from that IP to that endpoint are blocked for **15 minutes**.

> Note: The built-in limiter is **in-memory and per-process**. In a multi-instance deployment, each instance maintains its own counters.

---

## Configuration Options

Rate limiting is configured via shared fields on both `ConfigPasswordless` and `ConfigUsernameAndPassword`.

```go
// Shared rate limiting options (for both configs)
DisableRateLimit   bool                                                                                 // Set to true to disable rate limiting (not recommended for production)
FuncCheckRateLimit func(ip string, endpoint string) (allowed bool, retryAfter time.Duration, err error) // Optional: override default rate limiter
MaxLoginAttempts   int                                                                                  // Maximum attempts before lockout (default: 5)
LockoutDuration    time.Duration                                                                        // Duration for sliding window and lockout (default: 15 minutes)
```

### Examples

#### Passwordless Flow

```go
authInstance, err := auth.NewPasswordlessAuth(auth.ConfigPasswordless{
    Endpoint:             "/auth",
    UrlRedirectOnSuccess: "/dashboard",

    // Rate limiting (defaults shown explicitly)
    MaxLoginAttempts: 5,
    LockoutDuration:  15 * time.Minute,
})
```

#### Username/Password Flow

```go
authInstance, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
    Endpoint:             "/auth",
    UrlRedirectOnSuccess: "/dashboard",

    // Rate limiting (defaults shown explicitly)
    MaxLoginAttempts: 5,
    LockoutDuration:  15 * time.Minute,
})
```

> You can temporarily set `DisableRateLimit: true` in development environments, but this is **strongly discouraged** in production.

---

## Using a Custom Rate Limiter

For production systems running multiple instances, you will usually want a **shared rate limiter** (for example, backed by Redis or another central store).

You can plug in your own implementation via `FuncCheckRateLimit`:

```go
import "time"

func myRateLimiter(ip string, endpoint string) (bool, time.Duration, error) {
    // Look up and increment attempt counters in your store (e.g., Redis).
    // Return allowed = false and a positive retryAfter when the limit is exceeded.
    // Return an error only for internal failures (network, store, etc.).

    // Example (pseudo-code):
    // attempts := redis.Incr(key(ip, endpoint))
    // if attempts > 5 {
    //     return false, 15 * time.Minute, nil
    // }
    // return true, 0, nil

    return true, 0, nil
}

authInstance, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
    Endpoint:             "/auth",
    UrlRedirectOnSuccess: "/dashboard",

    FuncCheckRateLimit: myRateLimiter,
})
```

**Important notes:**

- If `FuncCheckRateLimit` returns an error, the library **fails open** (request is allowed) to avoid accidental outages caused by the limiter.
- Always treat rate limiting as a **defense-in-depth** control. It complements, but does not replace, strong passwords, MFA, and other security measures.

---

## When to Disable Rate Limiting

You might temporarily disable rate limiting by setting `DisableRateLimit: true` in:

- Local development
- Certain automated test environments

However:

- Never deploy with `DisableRateLimit: true` to production.
- Prefer using realistic rate limits and adjusting thresholds instead.

---

## Related Documentation

- [README.md](../README.md) – high-level overview and quick start
- [overview.md](./overview.md) – architecture and package overview
- [critical_review.md](./critical_review.md) – security and production-readiness analysis
