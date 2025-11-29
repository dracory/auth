# Package Overview: dracory/auth

**Last Updated:** 2025-11-28

---

## What is dracory/auth?

`dracory/auth` is a **batteries-included authentication library** for Go that provides ready-to-use authentication flows with minimal setup. It's designed to be **implementation-agnostic**, meaning you bring your own database, session store, and email service—the library handles all the authentication logic, UI pages, and API endpoints.

## Core Philosophy

1. **Two Authentication Strategies**
   - **Username/Password Flow** - Traditional authentication with password storage
   - **Passwordless Flow** - Email-based verification codes (more secure, no password storage)

2. **Complete Solution**
   - Pre-built HTML pages for login, registration, password reset
   - JSON API endpoints for SPA/mobile apps
   - Authentication middleware for protecting routes
   - Customizable email templates

3. **Production-Grade Security**
   - Structured error handling with error codes
   - CSRF protection and rate limiting
   - Session invalidation on password reset
   - Secure cookie defaults and input validation
   - Structured logging for audit trails

4. **Flexible Storage**
   - You implement the storage layer via callback functions
   - Works with any database (SQL, NoSQL, in-memory)
   - You control session/token management

---

## 🏗️ Architecture Overview

```mermaid
graph TB
    subgraph "Entry Points"
        A[NewPasswordlessAuth]
        B[NewUsernameAndPasswordAuth]
    end
    
    subgraph "Core Auth Struct"
        C[Auth]
        C --> D[Configuration]
        C --> E[Routing]
        C --> F[Middleware]
    end
    
    subgraph "Package Organization"
        T[types/]
        U[utils/]
        IA[internal/api/]
        IU[internal/ui/]
    end
    
    subgraph "Flows"
        G[Passwordless Flow]
        H[Username/Password Flow]
    end
    
    subgraph "User Implementation"
        I[Storage Functions]
        J[Email Functions]
        K[Temporary Key Store]
    end
    
    A --> C
    B --> C
    C --> T
    C --> U
    C --> IA
    C --> IU
    C --> G
    C --> H
    G --> I
    G --> J
    G --> K
    H --> I
    H --> J
    H --> K
    
    style C fill:#4CAF50
    style T fill:#FF9800
    style U fill:#FF9800
    style IA fill:#2196F3
    style IU fill:#2196F3
    style I fill:#9C27B0
    style J fill:#9C27B0
    style K fill:#9C27B0
```

---

## 📦 Package Organization

The library is organized into focused packages for maintainability:

### `types/` - Type Definitions

All configuration structs, interfaces, and type definitions:

- **[auth_interfaces.go](file:///d:/PROJECTs/_modules_dracory/auth/types/auth_interfaces.go)** - Core interfaces (`AuthSharedInterface`, etc.)
- **[config_passwordless.go](file:///d:/PROJECTs/_modules_dracory/auth/types/config_passwordless.go)** - `ConfigPasswordless` struct
- **[config_username_and_password.go](file:///d:/PROJECTs/_modules_dracory/auth/types/config_username_and_password.go)** - `ConfigUsernameAndPassword` struct
- **[user_auth_options.go](file:///d:/PROJECTs/_modules_dracory/auth/types/user_auth_options.go)** - `UserAuthOptions` type
- **[password.go](file:///d:/PROJECTs/_modules_dracory/auth/types/password.go)** - Password-related types
- **[cookie_config.go](file:///d:/PROJECTs/_modules_dracory/auth/types/cookie_config.go)** - Cookie configuration
- **[constants.go](file:///d:/PROJECTs/_modules_dracory/auth/types/constants.go)** - Type-level constants

**Usage:** Import as `"github.com/dracory/auth/types"`

### `utils/` - Utility Functions

Reusable utility functions used throughout the library:

- **[auth_cookies.go](file:///d:/PROJECTs/_modules_dracory/auth/utils/auth_cookies.go)** - Cookie management helpers
- **[auth_token_retrieve.go](file:///d:/PROJECTs/_modules_dracory/auth/utils/auth_token_retrieve.go)** - Token extraction from requests
- **[bearer_token_from_header.go](file:///d:/PROJECTs/_modules_dracory/auth/utils/bearer_token_from_header.go)** - Bearer token parsing
- **[email_validation.go](file:///d:/PROJECTs/_modules_dracory/auth/utils/email_validation.go)** - Email format validation
- **[password_strength.go](file:///d:/PROJECTs/_modules_dracory/auth/utils/password_strength.go)** - Password strength checking
- **[login_code.go](file:///d:/PROJECTs/_modules_dracory/auth/utils/login_code.go)** - Verification code generation and validation
- **[rate_limiter.go](file:///d:/PROJECTs/_modules_dracory/auth/utils/rate_limiter.go)** - In-memory rate limiting implementation
- **[cookies.go](file:///d:/PROJECTs/_modules_dracory/auth/utils/cookies.go)** - General cookie utilities
- **[scribble.go](file:///d:/PROJECTs/_modules_dracory/auth/utils/scribble.go)** - JSON file storage helper

**Usage:** Import as `"github.com/dracory/auth/utils"` (internal use only)

### `internal/api/` - API Endpoint Handlers

Each API endpoint has its own subdirectory with handler, dependencies, and tests:

- **[api_login/](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_login)** - Login endpoint (passwordless and username/password)
- **[api_login_code_verify/](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_login_code_verify)** - Passwordless code verification
- **[api_register/](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_register)** - Registration endpoint
- **[api_register_code_verify/](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_register_code_verify)** - Registration code verification
- **[api_logout/](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_logout)** - Logout endpoint
- **[api_password_restore/](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_password_restore)** - Password reset request
- **[api_password_reset/](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_password_reset)** - Password reset completion
- **[api_authenticate_via_username/](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_authenticate_via_username)** - Username authentication helper

Each subdirectory typically contains:
- Main handler file (e.g., `api_login.go`)
- Dependencies interface (e.g., `dependencies.go`)
- Tests (e.g., `api_login_test.go`)
- Constants if needed (e.g., `constants.go`)

### `internal/ui/` - UI Page Handlers

Each UI page has its own subdirectory with handler, content generation, and tests:

- **[page_login/](file:///d:/PROJECTs/_modules_dracory/auth/internal/ui/page_login)** - Login page
- **[page_login_code_verify/](file:///d:/PROJECTs/_modules_dracory/auth/internal/ui/page_login_code_verify)** - Code verification page
- **[page_register/](file:///d:/PROJECTs/_modules_dracory/auth/internal/ui/page_register)** - Registration page
- **[page_register_code_verify/](file:///d:/PROJECTs/_modules_dracory/auth/internal/ui/page_register_code_verify)** - Registration verification page
- **[page_logout/](file:///d:/PROJECTs/_modules_dracory/auth/internal/ui/page_logout)** - Logout page
- **[page_password_restore/](file:///d:/PROJECTs/_modules_dracory/auth/internal/ui/page_password_restore)** - Password restore request page
- **[page_password_reset/](file:///d:/PROJECTs/_modules_dracory/auth/internal/ui/page_password_reset)** - Password reset page
- **[shared/](file:///d:/PROJECTs/_modules_dracory/auth/internal/ui/shared)** - Shared UI components and utilities

Each page subdirectory typically contains:
- Main page handler (e.g., `page_login.go`)
- Content generation (e.g., `content.go`)
- Dependencies interface (e.g., `dependencies.go`)
- Tests (e.g., `page_login_test.go`)

### `examples/` - Working Example Applications

Complete, runnable example applications:

- **[passwordless/](file:///d:/PROJECTs/_modules_dracory/auth/examples/passwordless)** - Passwordless authentication example
  - In-memory storage
  - Email sending via localhost:1025
  - Complete callback implementations
  
- **[usernamepassword/](file:///d:/PROJECTs/_modules_dracory/auth/examples/usernamepassword)** - Username/password authentication example
  - In-memory storage
  - Password reset flow
  - Email sending via localhost:1025
  - Complete callback implementations

### Root Package - Public API

The root package provides the public API:

- **Constructors:** `NewPasswordlessAuth()`, `NewUsernameAndPasswordAuth()`
- **Middleware:** `WebAuthOrRedirectMiddleware()`, `ApiAuthOrErrorMiddleware()`, etc.
- **Main type:** `authImplementation` (implements `AuthSharedInterface`)
- **Delegation:** Thin wrappers that delegate to `internal/api/` and `internal/ui/` packages

---

## 🔑 Key Components

### 1. **Auth Struct** ([auth_implementation.go](file:///d:/PROJECTs/_modules_dracory/auth/auth_implementation.go))

The central struct that holds all configuration and provides methods for authentication operations.

**Key Fields:**
- `endpoint` - Base URL path for auth routes (e.g., `/auth`)
- `enableRegistration` - Toggle registration feature
- `urlRedirectOnSuccess` - Where to redirect after successful auth
- `useCookies` / `useLocalStorage` - Token storage strategy
- `passwordless` - Flag to determine which flow is active
- `logger` - Optional `*slog.Logger` for structured logging
- Function callbacks for user operations (login, register, logout, etc.)

**Key Methods:**
- `Router()` - Returns HTTP router with all auth endpoints
- `WebAuthOrRedirectMiddleware()` - Protects web routes
- `ApiAuthOrErrorMiddleware()` - Protects API routes
- `GetCurrentUserID()` - Retrieves authenticated user from context
- `LinkLogin()`, `LinkRegister()`, etc. - URL helpers

### 2. **Configuration Structs** (in `types/` package)

**[types.ConfigPasswordless](file:///d:/PROJECTs/_modules_dracory/auth/types/config_passwordless.go):**
```go
type ConfigPasswordless struct {
    // Required
    Endpoint                string
    UrlRedirectOnSuccess    string
    FuncUserFindByAuthToken func(ctx context.Context, sessionID string, options UserAuthOptions) (userID string, err error)
    FuncUserFindByEmail     func(ctx context.Context, email string, options UserAuthOptions) (userID string, err error)
    FuncUserLogout          func(ctx context.Context, userID string, options UserAuthOptions) error
    FuncUserStoreAuthToken  func(ctx context.Context, sessionID string, userID string, options UserAuthOptions) error
    FuncEmailSend           func(ctx context.Context, email string, subject string, body string) error
    FuncTemporaryKeyGet     func(key string) (value string, err error)
    FuncTemporaryKeySet     func(key string, value string, expiresSeconds int) error
    UseCookies              bool // OR UseLocalStorage (one must be true)
    
    // Optional
    EnableRegistration           bool
    FuncUserRegister             func(ctx context.Context, email, firstName, lastName string, options UserAuthOptions) error
    FuncEmailTemplateLoginCode   func(ctx context.Context, email, loginLink string, options UserAuthOptions) string
    FuncEmailTemplateRegisterCode func(ctx context.Context, email, registerLink string, options UserAuthOptions) string
    FuncLayout                   func(content string) string
    Logger                       *slog.Logger // Optional structured logger (defaults to slog.Default when nil)
}
```

**[types.ConfigUsernameAndPassword](file:///d:/PROJECTs/_modules_dracory/auth/types/config_username_and_password.go):**
```go
type ConfigUsernameAndPassword struct {
    // Similar to passwordless, plus:
    FuncUserLogin          func(ctx context.Context, username, password string, options UserAuthOptions) (userID string, err error)
    FuncUserPasswordChange func(ctx context.Context, username, newPassword string, options UserAuthOptions) error
    FuncUserRegister       func(ctx context.Context, username, password, firstName, lastName string, options UserAuthOptions) error
    FuncUserFindByUsername func(ctx context.Context, username, firstName, lastName string, options UserAuthOptions) (userID string, err error)
    EnableVerification     bool // Email verification for registration
    Logger                 *slog.Logger // Optional structured logger
}
```

### 3. **Authentication Flows**

#### Passwordless Flow

```mermaid
sequenceDiagram
    participant User
    participant Browser
    participant Auth
    participant YourApp
    participant Email
    
    User->>Browser: Enter email
    Browser->>Auth: POST /auth/api/login
    Auth->>Auth: Generate 8-char code
    Auth->>YourApp: FuncTemporaryKeySet(code, email)
    Auth->>Email: Send code via FuncEmailSend
    Auth->>Browser: Success response
    
    User->>Browser: Enter code from email
    Browser->>Auth: POST /auth/api/login-code-verify
    Auth->>YourApp: FuncTemporaryKeyGet(code)
    Auth->>YourApp: FuncUserFindByEmail(email)
    Auth->>Auth: Generate auth token
    Auth->>YourApp: FuncUserStoreAuthToken(token, userID)
    Auth->>Browser: Set cookie/return token
    Browser->>Browser: Redirect to dashboard
```

**Key Files:**
- [internal/api/api_login/api_login.go](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_login/api_login.go) - Sends verification code
- [internal/api/api_login_code_verify/api_login_code_verify.go](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_login_code_verify/api_login_code_verify.go) - Verifies code and authenticates
- [internal/ui/page_login/page_login.go](file:///d:/PROJECTs/_modules_dracory/auth/internal/ui/page_login/page_login.go) - HTML login page
- [internal/ui/page_login_code_verify/page_login_code_verify.go](file:///d:/PROJECTs/_modules_dracory/auth/internal/ui/page_login_code_verify/page_login_code_verify.go) - Code entry page

#### Username/Password Flow

```mermaid
sequenceDiagram
    participant User
    participant Browser
    participant Auth
    participant YourApp
    
    User->>Browser: Enter email & password
    Browser->>Auth: POST /auth/api/login
    Auth->>YourApp: FuncUserLogin(email, password)
    YourApp->>Auth: Return userID
    Auth->>Auth: Generate auth token
    Auth->>YourApp: FuncUserStoreAuthToken(token, userID)
    Auth->>Browser: Set cookie/return token
    Browser->>Browser: Redirect to dashboard
```

**Key Files:**
- [internal/api/api_login/api_login.go](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_login/api_login.go) - Authenticates user
- [login_with_username_and_password.go](file:///d:/PROJECTs/_modules_dracory/auth/login_with_username_and_password.go) - Core login logic
- [internal/api/api_password_restore/api_password_restore.go](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_password_restore/api_password_restore.go) - Password reset request
- [internal/api/api_password_reset/api_password_reset.go](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_password_reset/api_password_reset.go) - Password reset completion

### 4. **Routing System** ([router.go](file:///d:/PROJECTs/_modules_dracory/auth/router.go))

The router handles all authentication endpoints:

**API Endpoints (JSON responses):**
- `POST /auth/api/login` - Initiate login
- `POST /auth/api/login-code-verify` - Verify passwordless code
- `POST /auth/api/logout` - Logout user
- `POST /auth/api/register` - Initiate registration
- `POST /auth/api/register-code-verify` - Verify registration code
- `POST /auth/api/restore-password` - Request password reset
- `POST /auth/api/reset-password` - Complete password reset

**Page Endpoints (HTML responses):**
- `GET /auth/login` - Login page
- `GET /auth/login-code-verify` - Code verification page
- `GET /auth/logout` - Logout page
- `GET /auth/register` - Registration page
- `GET /auth/register-code-verify` - Registration verification page
- `GET /auth/password-restore` - Password restore request page
- `GET /auth/password-reset?t=TOKEN` - Password reset page

All paths defined in [constants.go](file:///d:/PROJECTs/_modules_dracory/auth/constants.go).

### 5. **Middleware** 

**[WebAuthOrRedirectMiddleware](file:///d:/PROJECTs/_modules_dracory/auth/web_auth_or_redirect_middleware.go)** - For web pages:
- Checks for auth token (cookie or header)
- Validates token via `FuncUserFindByAuthToken`
- On success: adds `userID` to request context
- On failure: redirects to login page

**[ApiAuthOrErrorMiddleware](file:///d:/PROJECTs/_modules_dracory/auth/api_auth_or_error_middleware.go)** - For API endpoints:
- Same validation logic
- On failure: returns JSON error response

**[WebAppendUserIdIfExistsMiddleware](file:///d:/PROJECTs/_modules_dracory/auth/web_append_user_id_if_exists_middleware.go)** - Optional middleware:
- Adds userID to context if authenticated
- Does NOT redirect/error if not authenticated
- Useful for pages that work for both authenticated and guest users

**Usage:**
```go
// Protect web routes
mux.Handle("/dashboard", auth.WebAuthOrRedirectMiddleware(dashboardHandler))

// Protect API routes
mux.Handle("/api/profile", auth.ApiAuthOrErrorMiddleware(profileHandler))

// Optional auth
mux.Handle("/", auth.WebAppendUserIdIfExistsMiddleware(homeHandler))
```

### 6. **Testing Approach**
    
    expectedMessage := `"message":"Email is required field"`
    HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", 
        authInstance.LinkApiLogin(), url.Values{}, expectedMessage, "%")
}
```

**Test Coverage:**
- ✅ Missing field validation
- ✅ Invalid input validation
- ✅ Error handling (DB errors, email errors)
- ✅ Success scenarios
- ✅ Token generation and storage
- ✅ Middleware behavior

### 6. **Development Examples**

The [examples](file:///d:/PROJECTs/_modules_dracory/auth/examples) directory contains working example applications:

**[Passwordless Example](file:///d:/PROJECTs/_modules_dracory/auth/examples/passwordless):**
- Complete passwordless authentication flow
- In-memory storage implementation
- Email sending via localhost:1025
- Registration with verification codes
- Protected dashboard route

**[Username/Password Example](file:///d:/PROJECTs/_modules_dracory/auth/examples/usernamepassword):**
- Traditional username/password authentication
- Password reset flow
- In-memory storage implementation
- Email sending via localhost:1025
- Registration with email verification
- Protected dashboard route

Both examples demonstrate:
- Complete implementation of all required callback functions
- Session/token management
- Email template customization
- Middleware usage for route protection

---

## 💡 How to Use This Package

**Step 1: Choose Your Flow**

```go
import (
    "github.com/dracory/auth"
    "github.com/dracory/auth/types"
)

// Passwordless
auth, err := auth.NewPasswordlessAuth(types.ConfigPasswordless{
    Endpoint: "/auth",
    UrlRedirectOnSuccess: "/dashboard",
    // ... implement required functions
})

// OR Username/Password
auth, err := auth.NewUsernameAndPasswordAuth(types.ConfigUsernameAndPassword{
    Endpoint: "/auth",
    UrlRedirectOnSuccess: "/dashboard",
    // ... implement required functions
})
```

**Step 2: Implement Required Functions**

You must implement:
- User storage (find, create, login)
- Token/session storage
- Email sending
- Temporary key storage (for verification codes)

**Step 3: Attach to Router**

```go
mux := http.NewServeMux()
mux.HandleFunc("/auth/", auth.Router().ServeHTTP)
```

**Step 4: Protect Routes**

```go
mux.Handle("/dashboard", auth.WebAuthOrRedirectMiddleware(dashboardHandler))
```

**Step 5: Get Current User**

```go
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
    userID := auth.GetCurrentUserID(r)
    // Use userID to fetch user data
}
```

---

## 🎯 Key Design Decisions

1. **Callback-Based Architecture** - Maximum flexibility, works with any storage
2. **Dual Flow Support** - Single package for both auth strategies
3. **Complete UI Included** - HTML pages with Bootstrap styling
4. **Token Storage Options** - Cookies OR localStorage (configurable)
5. **Verification Codes** - 8-character codes from limited alphabet (BCDFGHJKLMNPQRSTVXYZ) to avoid confusion
6. **UserAuthOptions + Context** - Callbacks receive `ctx context.Context` plus IP and UserAgent metadata for audit trails and cancellation
7. **Structured Logging with slog** - Core flows emit structured logs (using `log/slog`) including `email`, `user_id`, `ip`, and `user_agent` where available; callers can inject a custom `*slog.Logger` via configuration
8. **Structured Error Handling** - `AuthError` type with error codes ensures user-facing messages don't leak internal details while detailed errors are logged
9. **Security by Default** - CSRF protection, rate limiting, secure cookies, session invalidation on password reset, constant-time password comparison

---

## 📦 Dependencies

**External:**
- `github.com/dracory/api` - JSON API response helpers
- `github.com/dracory/hb` - HTML builder for pages
- `github.com/dracory/req` - Request parsing utilities
- `github.com/dracory/str` - String utilities (random generation)
- `github.com/jordan-wright/email` - Email sending
- `github.com/spf13/cast` - Type conversion

**Internal Dracory Ecosystem:**
- Part of a larger framework with consistent patterns
- Uses shared utilities across packages

---

## 📚 Related Documentation

- [README.md](../README.md) - Getting started guide with examples
- [critical_review.md](./critical_review.md) - Security analysis and production readiness assessment
