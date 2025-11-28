# Critical Review: dracory/auth

**Review Date:** 2025-11-27  
**Reviewer:** Critical Analysis  
**Perspective:** Security, Architecture, Production Readiness

---

## Executive Summary

The `dracory/auth` library demonstrates **solid engineering fundamentals** with good test coverage (90.2%) and a well-thought-out callback architecture. However, several **critical security and production concerns** prevent it from being recommended for production use without significant modifications.

**Overall Rating:** ⚠️ **Not Production-Ready** (Requires Security Hardening)

### Key Findings

| Category | Rating | Summary |
|----------|--------|---------|
| **Security** | 🔴 Critical Issues | Weak error messages expose internals; CSRF & rate limiting implemented but other gaps remain |
| **Architecture** | 🟢 Good | Clean callback pattern, good separation of concerns |
| **Error Handling** | 🔴 Poor | Exposes internal errors, inconsistent patterns, uses `log.Println` |
| **Input Validation** | 🟡 Basic | Email validation present but incomplete, no password strength enforcement |
| **Testing** | 🟢 Excellent | 90.2% coverage, comprehensive test suite |
| **Documentation** | 🟡 Good | Well-documented but missing security guidance |
| **Context Propagation** | � Implemented | `context.Context` propagated to public APIs and callbacks |
| **Observability** | 🔴 Poor | Basic `log.Println`, no structured logging or metrics |

---

## 🔴 Critical Security Issues

### 2. **No Password Strength Enforcement** - MEDIUM

**Severity:** 🟡 **MEDIUM**  
**Impact:** Weak passwords compromise accounts

**Problem:**
The library accepts ANY password, no matter how weak:

```go
// api_password_reset.go
if password == "" {
    api.Respond(w, r, api.Error("Password is required field"))
    return
}
// ❌ No strength check - "1" is valid, "password" is valid
```

**Recommendation:**
```go
type PasswordStrengthConfig struct {
    MinLength          int  // e.g., 8
    RequireUppercase   bool
    RequireLowercase   bool
    RequireDigit       bool
    RequireSpecial     bool
    ForbidCommonWords  bool
}

func validatePasswordStrength(password string, config PasswordStrengthConfig) error {
    if len(password) < config.MinLength {
        return errors.New("password too short")
    }
    // ... additional checks
}
```

---

### 3. **Session Fixation Vulnerability** - MEDIUM

**Severity:** 🟡 **MEDIUM**  
**Impact:** Session hijacking

**Problem:**
Tokens are generated client-side in passwordless flow:

```go
// api_login.go:33 - Passwordless flow
verificationCode := req.GetStringTrimmed(r, "verification_code")
// ❌ Client provides the code, not server-generated
```

**Attack:**
1. Attacker generates code "ABCD1234"
2. Attacker sends to victim: "Your code is ABCD1234"
3. Victim uses code, gets authenticated
4. Attacker uses same code (if not invalidated properly)

**Recommendation:**
```go
// Server generates code, not client
verificationCode, err := str.RandomFromGamma(LoginCodeLength, LoginCodeGamma)
if err != nil {
    api.Respond(w, r, api.Error("Failed to generate code"))
    return
}
```

---

## 🟠 High Priority Issues

### 4. **Inconsistent Error Handling** - HIGH

**Problem:**
Mixed error handling patterns:

```go
// Pattern 1: Log and return generic error
log.Println(errEmailSent)  // ❌ Using log.Println
api.Respond(w, r, api.Error("Login code failed to be send"))

// Pattern 2: Return error details
api.Respond(w, r, api.Error("token store failed. "+errTempTokenSave.Error()))  // ❌ Exposes internals

// Pattern 3: Silent failure
if err != nil {
    return  // ❌ No logging, no user feedback
}
```

**Recommendation:**
Standardize on structured logging with error codes:

```go
type AuthError struct {
    Code       string
    Message    string  // User-facing
    InternalErr error  // For logging only
}

func (e AuthError) Error() string {
    return e.Message
}

// Usage
if err != nil {
    authErr := AuthError{
        Code:       "EMAIL_SEND_FAILED",
        Message:    "Failed to send verification email",
        InternalErr: err,
    }
    logger.Error("Email send failed", "error", err, "email", email)
    api.Respond(w, r, api.Error(authErr.Message))
    return
}
```

---

### 5. **No Structured Logging** - HIGH

**Problem:**
Uses basic `log.Println`:

```go
log.Println(errEmailSent)  // ❌ No context, no levels, no structure
log.Println(urlApiLogout)  // ❌ Debugging code left in production
```

**Impact:**
- Cannot filter logs by level
- Cannot search/query logs effectively
- No correlation IDs for request tracing
- Debugging statements leak to production

**Recommendation:**
```go
import "log/slog"

// In config
type Config struct {
    Logger *slog.Logger  // Configurable logger
}

// Usage
a.logger.Error("Email send failed",
    "error", err,
    "email", email,
    "ip", options.UserIp,
    "user_agent", options.UserAgent,
)
```

---

### 6. **Timing Attack Vulnerability** - MEDIUM

**Problem:**
String comparisons may leak timing information:

```go
// api_password_reset.go:16
if password != passwordConfirm {  // ❌ Not constant-time
    api.Respond(w, r, api.Error("Passwords do not match"))
    return
}
```

**Recommendation:**
```go
import "crypto/subtle"

if subtle.ConstantTimeCompare([]byte(password), []byte(passwordConfirm)) != 1 {
    api.Respond(w, r, api.Error("Passwords do not match"))
    return
}
```

---

## 🟡 Medium Priority Issues

### 7. **Hardcoded Cookie Settings** - MEDIUM

**Problem:**
Cookie security settings are not configurable:

```go
// auth_cookie_remove.go:14
cookie := http.Cookie{
    HttpOnly: false,  // ❌ Should be true for security
    Secure:   secureCookie,
    SameSite: 0,      // ❌ Not set, should be SameSiteLax or SameSiteStrict
}
```

**Recommendation:**
```go
type CookieConfig struct {
    HttpOnly bool          // Default: true
    Secure   bool          // Default: true in production
    SameSite http.SameSite // Default: http.SameSiteLax
    MaxAge   int           // Default: 2 hours
    Domain   string        // Configurable
    Path     string        // Default: "/"
}
```

---

### 8. **No Input Sanitization** - MEDIUM

**Problem:**
User inputs are not sanitized before storage/display:

```go
firstName := req.GetStringTrimmed(r, "first_name")  // ❌ No sanitization
lastName := req.GetStringTrimmed(r, "last_name")    // ❌ Could contain XSS
```

**Recommendation:**
```go
import "html"

firstName := html.EscapeString(req.GetStringTrimmed(r, "first_name"))
lastName := html.EscapeString(req.GetStringTrimmed(r, "last_name"))

// OR provide sanitization callback
type Config struct {
    FuncSanitizeInput func(input string) string
}
```

---

### 9. **Email Validation Inconsistency** - MEDIUM

**Problem:**
Email validation is inconsistent:

```go
// Some endpoints validate
if _, err := mail.ParseAddress(email); err != nil {
    return api.Error("Invalid email")
}

// Others don't validate at all
email := req.GetStringTrimmed(r, "email")
// ❌ No validation, directly used
```

**Recommendation:**
Create centralized validation function used everywhere.

---

### 10. **No Account Enumeration Protection** - MEDIUM

**Problem:**
Different error messages reveal if user exists:

```go
// If user exists but wrong password
"authentication failed. invalid password"

// If user doesn't exist
"User not found"
```

Attacker can enumerate valid email addresses.

**Recommendation:**
Always return same message: "Invalid credentials"

---

### 11. **Deprecated Code Not Removed** - LOW

**Problem:**
Deprecated middleware still in codebase:

```go
// auth_middleware.go
// DEPRECATED use the Web or the API middleware instead
// func (a Auth) AuthMiddleware(next http.Handler) http.Handler {
//     ... 30+ lines of commented code
// }
```

**Recommendation:**
Remove deprecated code entirely. Add migration guide to docs.

---

## 🟢 Strengths

### 1. **Excellent Test Coverage**

- 90.2% code coverage
- 34 comprehensive test files
- Tests cover error cases, edge cases, and happy paths
- Good use of table-driven tests

### 2. **Clean Architecture**

- Callback-based design provides flexibility
- Good separation of concerns
- Implementation-agnostic (works with any database)
- Clear distinction between API and web endpoints

### 3. **Dual Flow Support**

- Both passwordless and username/password in one package
- Well-designed configuration structs
- Easy to switch between flows

### 4. **Good Documentation**

- Comprehensive README
- Working examples in `development/` directory
- Clear function signatures
- Good inline comments

### 5. **Middleware Design**

- Three middleware options for different use cases
- Clear separation: Web vs API vs Optional
- Context-based user ID propagation

---

## 📊 Production Readiness Checklist

| Requirement | Status | Notes |
|-------------|--------|-------|
| Rate Limiting | ✅ Implemented | In-memory per-IP/per-endpoint limiter with lockout; configurable |
| CSRF Protection | ✅ Implemented | CSRF protection via `github.com/dracory/csrf` when enabled |
| Error Sanitization | 🟡 Partial | Core auth flows use generic messages; full error-code system not implemented |
| Structured Logging | ❌ Missing | Uses `log.Println` |
| Context Propagation | ✅ Implemented | `context.Context` propagated into public APIs and callbacks |
| Input Validation | 🟡 Partial | Email only, no password strength |
| Input Sanitization | ❌ Missing | XSS risk |
| Password Strength | ❌ Missing | Accepts any password |
| Account Lockout | ✅ Implemented | Lockout after N failed attempts via rate limiter |
| Session Management | 🟡 Basic | No session invalidation on password change |
| Audit Logging | 🟡 Partial | Has IP/UserAgent but no structured logs |
| Metrics/Monitoring | ❌ Missing | No instrumentation |
| Security Headers | ❌ Missing | No CSP, X-Frame-Options, etc. |
| Test Coverage | ✅ Excellent | 90.2% coverage |
| Documentation | ✅ Good | Comprehensive README |

**Production Ready:** ❌ **NO** - Requires security hardening

---

## 🎯 Recommended Action Plan

### Phase 1: Security Critical (MUST DO BEFORE PRODUCTION)

**Estimated Time:** 2-3 weeks

1. **Sanitize Error Messages**
   - Create error code system
   - Never expose internal errors to users
   - Log detailed errors internally only

2. **Add Structured Logging**
   - Replace `log.Println` with `slog`
   - Add log levels (Debug, Info, Warn, Error)
   - Include context in all logs (request ID, user ID, IP)

3. **Implement Context Propagation**
   - Add `context.Context` to all functions
   - Implement request timeouts
   - Add cancellation support

4. **Enforce Password Strength and Input Sanitization**
   - Add password strength requirements
   - Sanitize all user inputs
   - Validate all fields consistently

### Phase 2: Security Enhancements (SHOULD DO)

**Estimated Time:** 1-2 weeks

5. **Password Strength Enforcement**
   - Add configurable password requirements
   - Integrate with haveibeenpwned API (optional)
   - Add password complexity scoring

6. **Input Sanitization**
   - Sanitize all user inputs
   - Add XSS protection
   - Validate all fields consistently

7. **Improve Cookie Security**
   - Make cookie settings configurable
   - Set HttpOnly=true by default
   - Set SameSite=Lax by default
   - Add Secure flag for HTTPS

8. **Account Enumeration Protection**
   - Standardize all error messages
   - Add timing delays to prevent timing attacks
   - Use constant-time comparisons

### Phase 3: Production Hardening (NICE TO HAVE)

**Estimated Time:** 2-3 weeks

9. **Add Metrics/Monitoring**
    - Instrument all endpoints
    - Add Prometheus metrics
    - Track login success/failure rates
    - Monitor verification code usage

10. **Add Security Headers**
    - CSP (Content Security Policy)
    - X-Frame-Options
    - X-Content-Type-Options
    - Strict-Transport-Security

11. **Session Management**
    - Invalidate sessions on password change
    - Add session expiration
    - Add "logout all devices" functionality

12. **Audit Logging**
    - Log all authentication events
    - Include IP, UserAgent, timestamp
    - Make logs tamper-evident
    - Add log retention policy

---

## 💡 Architectural Recommendations

### 1. **Introduce Middleware Chain**

Instead of single middleware, allow chaining:

```go
auth.Use(
    RateLimitMiddleware(),
    CSRFMiddleware(),
    AuthMiddleware(),
)
```

### 2. **Add Hooks System**

Allow users to hook into authentication flow:

```go
type Hooks struct {
    BeforeLogin  func(ctx context.Context, email string) error
    AfterLogin   func(ctx context.Context, userID string) error
    OnLoginFail  func(ctx context.Context, email string, reason string)
}
```

### 3. **Separate Concerns**

Split into sub-packages:

```
auth/
├── core/          # Core authentication logic
├── middleware/    # HTTP middleware
├── handlers/      # HTTP handlers
├── validation/    # Input validation
├── security/      # Security utilities (rate limit, CSRF)
└── observability/ # Logging, metrics
```

---

## 🔍 Code Quality Issues

### 1. **Magic Numbers**

```go
errTempTokenSave := a.funcTemporaryKeySet(verificationCode, email, 3600)
                                                                    ^^^^
// Should be: const DefaultCodeExpiration = 1 * time.Hour
```

### 2. **Inconsistent Naming**

```go
FuncUserFindByAuthToken func(sessionID string, ...) // ❌ Parameter named sessionID but it's a token
```

### 3. **Code Duplication**

Similar validation logic duplicated between:
- `new_passwordless_auth.go`
- `new_username_and_password_auth.go`

Extract to shared validator.

### 4. **Typos in Error Messages**

```go
api.Error("Link not valid of expired")  // ❌ "of" should be "or"
api.Error("Login code failed to be send")  // ❌ "send" should be "sent"
```

---

## 📝 Conclusion

The `dracory/auth` library has a **solid foundation** with good architecture and excellent test coverage. However, it has **critical security vulnerabilities** that make it **unsuitable for production use** without significant hardening.

### Key Takeaways

✅ **Strengths:**
- Clean, flexible architecture
- Excellent test coverage (90.2%)
- Good documentation
- Dual authentication flow support

❌ **Critical Issues:**
- Exposes internal errors (information leakage)
- Poor logging (no structure, no levels)

### Final Recommendation

**DO NOT use in production without:**
1. Sanitizing all error messages
2. Adding structured logging
3. Enforcing password strength and input sanitization

**Estimated effort to production-ready:** 4-6 weeks of security hardening

**Alternative:** Consider using battle-tested libraries like:
- [go-pkgz/auth](https://github.com/go-pkgz/auth)
- [authorizerdev/authorizer](https://github.com/authorizerdev/authorizer)
- [ory/kratos](https://github.com/ory/kratos)

Or use this library as a **learning resource** and **starting point**, but implement all security recommendations before production deployment.

---

**Reviewed by:** Critical Security Analysis  
**Date:** 2025-11-27  
**Severity Scale:** 🔴 Critical | 🟠 High | 🟡 Medium | 🟢 Low
