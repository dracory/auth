# Project Structure

The `dracory/auth` library is organized into focused packages for clarity and maintainability.

## Directory Layout

```
dracory/auth/
├── types/              # Configuration structs, interfaces, and type definitions
│   ├── auth_interfaces.go
│   ├── config_passwordless.go
│   ├── config_username_and_password.go
│   └── ...
├── utils/              # Utility functions
│   ├── auth_cookies.go
│   ├── email_validation.go
│   ├── password_strength.go
│   ├── rate_limiter.go
│   ├── login_code.go
│   └── ...
├── internal/           # Internal implementation (not part of public API)
│   ├── api/           # API endpoint handlers (one subdirectory per endpoint)
│   │   ├── api_login/
│   │   ├── api_register/
│   │   ├── api_logout/
│   │   └── ...
│   └── ui/            # UI page handlers (one subdirectory per page)
│       ├── page_login/
│       ├── page_register/
│       └── ...
├── examples/           # Working example applications
│   ├── passwordless/
│   └── usernamepassword/
└── *.go               # Public API (constructors, middleware, main Auth type)
```

## Package Organization

### `types/` - Type Definitions

All configuration structs, interfaces, and type definitions. This is part of the **public API**.

**Files:**
- **[auth_interfaces.go](file:///d:/PROJECTs/_modules_dracory/auth/types/auth_interfaces.go)** - Core interfaces (`AuthSharedInterface`, etc.)
- **[config_passwordless.go](file:///d:/PROJECTs/_modules_dracory/auth/types/config_passwordless.go)** - `ConfigPasswordless` struct
- **[config_username_and_password.go](file:///d:/PROJECTs/_modules_dracory/auth/types/config_username_and_password.go)** - `ConfigUsernameAndPassword` struct
- **[user_auth_options.go](file:///d:/PROJECTs/_modules_dracory/auth/types/user_auth_options.go)** - `UserAuthOptions` type
- **[password.go](file:///d:/PROJECTs/_modules_dracory/auth/types/password.go)** - Password-related types
- **[cookie_config.go](file:///d:/PROJECTs/_modules_dracory/auth/types/cookie_config.go)** - Cookie configuration
- **[constants.go](file:///d:/PROJECTs/_modules_dracory/auth/types/constants.go)** - Type-level constants

**Usage:** Import as `"github.com/dracory/auth/types"`

```go
import (
    "github.com/dracory/auth"
    "github.com/dracory/auth/types"
)

auth, err := auth.NewPasswordlessAuth(types.ConfigPasswordless{
    // ...
})
```

### `utils/` - Utility Functions

Reusable utility functions used throughout the library. For **internal use only**.

**Files:**
- **[auth_cookies.go](file:///d:/PROJECTs/_modules_dracory/auth/utils/auth_cookies.go)** - Cookie management helpers
- **[auth_token_retrieve.go](file:///d:/PROJECTs/_modules_dracory/auth/utils/auth_token_retrieve.go)** - Token extraction from requests
- **[bearer_token_from_header.go](file:///d:/PROJECTs/_modules_dracory/auth/utils/bearer_token_from_header.go)** - Bearer token parsing
- **[email_validation.go](file:///d:/PROJECTs/_modules_dracory/auth/utils/email_validation.go)** - Email format validation
- **[password_strength.go](file:///d:/PROJECTs/_modules_dracory/auth/utils/password_strength.go)** - Password strength checking
- **[login_code.go](file:///d:/PROJECTs/_modules_dracory/auth/utils/login_code.go)** - Verification code generation and validation
- **[rate_limiter.go](file:///d:/PROJECTs/_modules_dracory/auth/utils/rate_limiter.go)** - In-memory rate limiting implementation
- **[cookies.go](file:///d:/PROJECTs/_modules_dracory/auth/utils/cookies.go)** - General cookie utilities
- **[scribble.go](file:///d:/PROJECTs/_modules_dracory/auth/utils/scribble.go)** - JSON file storage helper

### `internal/api/` - API Endpoint Handlers

Each API endpoint has its own subdirectory with handler, dependencies, and tests. **Not part of public API**.

**Subdirectories:**
- **[api_login/](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_login)** - Login endpoint (passwordless and username/password)
- **[api_login_code_verify/](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_login_code_verify)** - Passwordless code verification
- **[api_register/](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_register)** - Registration endpoint
- **[api_register_code_verify/](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_register_code_verify)** - Registration code verification
- **[api_logout/](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_logout)** - Logout endpoint
- **[api_password_restore/](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_password_restore)** - Password reset request
- **[api_password_reset/](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_password_reset)** - Password reset completion
- **[api_authenticate_via_username/](file:///d:/PROJECTs/_modules_dracory/auth/internal/api/api_authenticate_via_username)** - Username authentication helper

**Typical structure of each subdirectory:**
- Main handler file (e.g., `api_login.go`)
- Dependencies interface (e.g., `dependencies.go`)
- Tests (e.g., `api_login_test.go`)
- Constants if needed (e.g., `constants.go`)

### `internal/ui/` - UI Page Handlers

Each UI page has its own subdirectory with handler, content generation, and tests. **Not part of public API**.

**Subdirectories:**
- **[page_login/](file:///d:/PROJECTs/_modules_dracory/auth/internal/ui/page_login)** - Login page
- **[page_login_code_verify/](file:///d:/PROJECTs/_modules_dracory/auth/internal/ui/page_login_code_verify)** - Code verification page
- **[page_register/](file:///d:/PROJECTs/_modules_dracory/auth/internal/ui/page_register)** - Registration page
- **[page_register_code_verify/](file:///d:/PROJECTs/_modules_dracory/auth/internal/ui/page_register_code_verify)** - Registration verification page
- **[page_logout/](file:///d:/PROJECTs/_modules_dracory/auth/internal/ui/page_logout)** - Logout page
- **[page_password_restore/](file:///d:/PROJECTs/_modules_dracory/auth/internal/ui/page_password_restore)** - Password restore request page
- **[page_password_reset/](file:///d:/PROJECTs/_modules_dracory/auth/internal/ui/page_password_reset)** - Password reset page
- **[shared/](file:///d:/PROJECTs/_modules_dracory/auth/internal/ui/shared)** - Shared UI components and utilities

**Typical structure of each subdirectory:**
- Main page handler (e.g., `page_login.go`)
- Content generation (e.g., `content.go`)
- Dependencies interface (e.g., `dependencies.go`)
- Tests (e.g., `page_login_test.go`)

### `examples/` - Working Example Applications

Complete, runnable example applications demonstrating both authentication flows.

**Examples:**
- **[passwordless/](file:///d:/PROJECTs/_modules_dracory/auth/examples/passwordless)** - Passwordless authentication example
  - In-memory storage
  - Email sending via localhost:1025
  - Complete callback implementations
  - Registration with verification codes
  - Protected dashboard route
  
- **[usernamepassword/](file:///d:/PROJECTs/_modules_dracory/auth/examples/usernamepassword)** - Username/password authentication example
  - In-memory storage
  - Password reset flow
  - Email sending via localhost:1025
  - Complete callback implementations
  - Registration with email verification
  - Protected dashboard route

### Root Package - Public API

The root package provides the public API that users interact with:

**Key Components:**
- **Constructors:** `NewPasswordlessAuth()`, `NewUsernameAndPasswordAuth()`
- **Middleware:** `WebAuthOrRedirectMiddleware()`, `ApiAuthOrErrorMiddleware()`, `WebAppendUserIdIfExistsMiddleware()`
- **Main type:** `authImplementation` (implements `AuthSharedInterface` from `types/`)
- **Delegation:** Thin wrappers that delegate to `internal/api/` and `internal/ui/` packages

**Key Files:**
- **[auth_implementation.go](file:///d:/PROJECTs/_modules_dracory/auth/auth_implementation.go)** - Main Auth struct with getters/setters
- **[auth_implementation_api.go](file:///d:/PROJECTs/_modules_dracory/auth/auth_implementation_api.go)** - API endpoint delegation
- **[auth_implementation_pages.go](file:///d:/PROJECTs/_modules_dracory/auth/auth_implementation_pages.go)** - UI page delegation
- **[new_passwordless_auth.go](file:///d:/PROJECTs/_modules_dracory/auth/new_passwordless_auth.go)** - Passwordless constructor
- **[new_username_and_password_auth.go](file:///d:/PROJECTs/_modules_dracory/auth/new_username_and_password_auth.go)** - Username/password constructor
- **[router.go](file:///d:/PROJECTs/_modules_dracory/auth/router.go)** - HTTP router setup
- **Middleware files** - Various middleware implementations

## Design Principles

1. **Separation of Concerns** - Types, utilities, API handlers, and UI handlers are clearly separated
2. **Internal Encapsulation** - Implementation details in `internal/` are not part of the public API
3. **Thin Facade** - Root package provides a clean public API that delegates to internal packages
4. **Testability** - Each component has its own tests alongside the implementation
5. **Modularity** - Each API endpoint and UI page is self-contained in its own subdirectory

## Public vs Internal API

**Public API (safe to import):**
- `github.com/dracory/auth` - Main package with constructors and middleware
- `github.com/dracory/auth/types` - Configuration structs and interfaces

**Internal (do not import directly):**
- `github.com/dracory/auth/utils` - Internal utilities
- `github.com/dracory/auth/internal/*` - Internal implementation details

The Go compiler enforces this: packages under `internal/` cannot be imported by external code.
