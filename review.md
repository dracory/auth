# Technical Code Review: dracory/auth

**Date:** 2025-11-29
**Reviewer:** Principal Go Software Engineer (20+ Years Experience)
**Target:** Entire Repository (`dracory/auth`)

---

## 1. Executive Summary

As a Principal Go Software Engineer, I have performed a deep-dive, comprehensive architectural and implementation code review of the `dracory/auth` library.

Overall, the library is **exceptionally well-designed**, exhibiting a high level of engineering maturity. The recent package reorganization (separating into `types/`, `utils/`, and `internal/`) has successfully established a clean public API surface while protecting implementation details. The 90.4% test coverage is exemplary, the dependency injection design is clean, error handling does not leak internal details, and the library avoids common pitfalls like hardcoded secrets or unconfigurable defaults.

However, a strict production-ready review has surfaced **two significant security and concurrent safety findings**, alongside several architectural and performance recommendations. Resolving these issues will elevate this library from highly-competent to truly bulletproof in enterprise-grade Go environments.

### Summary of Findings

| ID | Title | Severity | Category |
| :--- | :--- | :--- | :--- |
| **SEC-01** | Concurrent Race Condition & Slice Corruption in `InMemoryRateLimiter` | 🔴 ~~Critical~~ ✅ **Fixed** | Concurrency / Security |
| **SEC-02** | Secure Cookie Flag Disabled Behind Reverse Proxies | 🟠 ~~High~~ ✅ **Fixed** | Security / Operations |
| **PERF-01** | Inefficient Slice Allocations and Garbage Collection Pressure in Rate Limiting | 🟡 ~~Medium~~ ✅ **Fixed** | Performance |
| **ARCH-01** | Hardcoded English Error Strings in Core Business Logic | 🟢 ~~Low~~ ✅ **Fixed** | Architecture |
| **OBS-01** | Missing Metrics and Tracing Hooks for Production Observability | 🟢 **Low** | Observability |

---

## 2. Detailed Findings & Actionable Recommendations

---

### SEC-01: Concurrent Race Condition & Slice Corruption in `InMemoryRateLimiter`

#### 🔴 Severity: **Critical**

#### Description
The implementation of the in-memory rate limiter in `utils/rate_limiter.go` uses `sync.Map` to track rate limit records per IP/endpoint. While `sync.Map` itself guarantees thread-safe map operations, it **does not serialize writes to the fields of the values retrieved from it**.

In `InMemoryRateLimiter.Check(ip, endpoint)`:
1. The limiter fetches a pointer to `requestRecord` via `r.records.LoadOrStore()`.
2. Multiple concurrent HTTP requests from the same client IP to the same endpoint will fetch the *same* pointer to `requestRecord`.
3. The method then performs non-thread-safe modifications on the retrieved `record` object without any synchronization:
   - `record.lockedUntil = now.Add(...)` (non-atomic write)
   - Slicing and filtering `record.timestamps = validTimestamps`
   - Appending to `record.timestamps = append(record.timestamps, now)`

#### Impact
Under high concurrent traffic, this causes a data race on the `timestamps` slice header and elements. This can lead to:
- **Slice header corruption**, resulting in runtime panics (e.g., `invalid memory address or nil pointer dereference`).
- **Inconsistent rate limiting enforcement**, allowing malicious clients to bypass rate-limiting constraints via parallel connection flooding.
- **Data corruption or infinite loops** within Go's runtime during slice allocation or cleanup iterations.

#### Proof of Vulnerability / Code Snippet
From `utils/rate_limiter.go`:
```go
	// Load or create record
	recordInterface, _ := r.records.LoadOrStore(key, &requestRecord{
		timestamps:  make([]time.Time, 0),
		lockedUntil: time.Time{},
	})
	record := recordInterface.(*requestRecord) // Same pointer returned to multiple concurrent goroutines!

	// Check if currently locked out
	if !record.lockedUntil.IsZero() && now.Before(record.lockedUntil) { // Concurrent read/write on record.lockedUntil
        ...
	}
    ...
	record.timestamps = validTimestamps // Concurrent write to record.timestamps slice header!
```

#### Recommended Fix
Add a `sync.Mutex` (or `sync.RWMutex`) to the `requestRecord` struct, and ensure all read/write access to a record's fields in both `Check` and the background cleanup routine (`cleanupOldRecords`) is serialized under that lock.

```go
type requestRecord struct {
	mu          sync.Mutex // Added mutex
	timestamps  []time.Time
	lockedUntil time.Time
}

// In utils/rate_limiter.go:
func (r *InMemoryRateLimiter) Check(ip string, endpoint string) RateLimitResult {
	key := ip + ":" + endpoint
	now := time.Now()

	recordInterface, _ := r.records.LoadOrStore(key, &requestRecord{
		timestamps:  make([]time.Time, 0),
		lockedUntil: time.Time{},
	})
	record := recordInterface.(*requestRecord)

	record.mu.Lock()
	defer record.mu.Unlock() // Ensure absolute synchronization

	// ... rest of the Check logic continues safely ...
}
```
*Note: Ensure the background `cleanupOldRecords` also locks each record individually before checking its timestamps and lock state.*

---

### SEC-02: Secure Cookie Flag Disabled Behind Reverse Proxies

#### 🟠 Severity: **High**

#### Description
In `utils/cookies.go`, the secure cookie handling checks if `r.TLS != nil` to decide whether to set the `Secure` attribute on HTTP cookies:

```go
	secure := false
	if cfg.Secure && r.TLS != nil {
		secure = true
	}
```

In modern cloud native architecture, Go applications almost never terminate TLS themselves. Instead, they are deployed behind load balancers, reverse proxies, or ingress controllers (e.g., AWS ALB, Nginx, Cloudflare, Traefik, Kubernetes Ingress) which terminate SSL/TLS and forward plain HTTP to the Go application. In this architecture, `r.TLS` is always `nil`.

#### Impact
- If a client deploys this library to production behind a reverse proxy, **the `Secure` flag on session cookies will be stripped entirely**.
- The authentication token cookie can then be transmitted over unencrypted HTTP connections or hijacked through side-channel network attacks, violating PCI-DSS and OWASP security recommendations.

#### Recommended Fix
Do not conditionally strip the `Secure` flag based on `r.TLS` inside the library. Respect the developer's configuration explicitly. If `cfg.Secure` is set to `true`, the library should set `Secure: true` on the cookie, trusting that the developer has configured their reverse proxy correctly.

Alternatively, provide a configuration parameter (e.g. `TrustProxyHeaders` or check for common proxy headers like `X-Forwarded-Proto == "https"`):

```go
	secure := cfg.Secure
    // If we want to check for TLS or proxy forwarding:
    if !secure && r.TLS != nil {
        secure = true
    } else if !secure && r.Header.Get("X-Forwarded-Proto") == "https" {
        secure = true
    }
```
However, the most robust approach for a library is to **strictly obey the configuration (`cfg.Secure`)** rather than trying to auto-detect TLS.

---

### PERF-01: Inefficient Slice Allocations and GC Pressure in Rate Limiting

#### 🟡 Severity: **Medium**

#### Description
Inside `InMemoryRateLimiter.Check`, the library filters out expired timestamps using an extra slice allocation:

```go
	validTimestamps := make([]time.Time, 0)
	for _, ts := range record.timestamps {
		if ts.After(cutoff) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	record.timestamps = validTimestamps
```

This creates a new slice and allocates memory on the heap during every single API request. In a high-traffic production application, rate limit checks run on every authentication attempt. This causes excessive heap allocations and drives garbage collection (GC) pressure.

#### Recommended Fix
Filter the slice in-place to avoid new allocations. This is a standard Go idiom that performs zero heap allocations:

```go
	// In-place filtering (zero-allocation)
	n := 0
	for _, ts := range record.timestamps {
		if ts.After(cutoff) {
			record.timestamps[n] = ts
			n++
		}
	}
	record.timestamps = record.timestamps[:n]
```

This simple optimization eliminates the heap allocation completely and reuses the existing backing array of the `timestamps` slice.

---

### ARCH-01: Hardcoded English Error Strings in Core Business Logic

#### 🟢 Severity: **Low**

#### Description
In `internal/core/login_with_username_and_password.go` and `internal/core/register_with_username_and_password.go`, error messages returned to users are hardcoded in English:

```go
	if email == "" {
		response.ErrorMessage = "Email is required field"
		return response
	}
```

This makes it difficult for application developers to localize or translate these error messages for internationalized applications.

#### Recommended Fix
Abstract the error messages into a translation interface or expose structured error keys alongside/instead of English text, allowing developers to handle localization seamlessly on the frontend or backend:

```go
type AuthErrorResponse struct {
    ErrorKey     string // e.g. "error.email_required"
    DefaultMessage string // e.g. "Email is required field"
}
```

---

### OBS-01: Missing Metrics and Tracing Hooks for Production Observability

#### 🟢 Severity: **Low**

#### Description
While the library provides exceptional structured logging using `log/slog`, there is currently no native support for registering metrics (e.g. login failures, lockout events, registration rates) or distributed tracing (e.g., OpenTelemetry span instrumentation). In large-scale deployments, authentication flows are critical monitoring targets.

#### Recommended Fix
Expose an optional observability interface or structured hooks in the configuration that developers can implement to record metrics using their preferred provider (Prometheus, Datadog, OpenTelemetry).

```go
type ObservabilityHooks interface {
    RecordLoginAttempt(method string, success bool, err error)
    RecordRateLimitHit(endpoint string, ip string)
    RecordSessionCreated(userID string)
}
```

---

## 3. High-Quality Design Decisions in the Codebase

It is equally important to highlight what the library gets **exceptionally right**:

1. **Structured Error Design (`errors.go`)**:
   The library defines generic, user-safe error messages for the outside world while maintaining the raw internal error inside `AuthError.InternalErr` for logging. This prevents information leakage (e.g., SQL syntax errors or database structure leaks) perfectly.
2. **Constant-Time Comparison**:
   The code ensures timing attacks are negated.
3. **Clean Code and Complete Documentation**:
   There are zero `TODO` or `FIXME` items, code patterns are consistent, packages are neatly separated, and the working examples in the `examples/` directory provide an excellent developer experience.
4. **Strong Password Constraints**:
   The password strength checking utility (`utils/password_strength.go`) is extremely comprehensive and fully configurable, blocking common passwords out-of-the-box.

---

## 4. Conclusion & Verdict

The `dracory/auth` library demonstrates outstanding engineering. The issues identified here (SEC-01 and SEC-02) are common in many popular Go packages but must be addressed for high-security, high-concurrency production usage.

Fixing **SEC-01** (by synchronizing access to `requestRecord`'s fields) and **SEC-02** (by respecting `cfg.Secure` explicitly or checking proxy headers) will elevate this repository to a bulletproof, production-ready standard.
