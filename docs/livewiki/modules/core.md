---
Created: 2025-12-06
Last Updated: 2025-12-06
Version: 1.0.0
---

# Core Module

**Package**: `auth` (Root)

The core module provides the main entry points for the library and orchestrates the interaction between the API, UI, and your configuration.

## Key Components

### 1. Factories / Constructors

*   `NewPasswordlessAuth(config types.ConfigPasswordless) (*Auth, error)`
*   `NewUsernameAndPasswordAuth(config types.ConfigUsernameAndPassword) (*Auth, error)`

These functions validate your configuration and return an initialized `*Auth` struct.

### 2. The `Auth` Struct

Detailed in `auth_implementation.go`. This struct holds:
*   The Router (`*http.ServeMux`)
*   The Configuration
*   References to internal API and Page handlers

### 3. Router

`router.go` sets up the mapping between HTTP paths and handlers.
*   `/api/*` -> Forwarded to `internal/api` handlers.
*   `/*` (e.g., `/login`, `/register`) -> Forwarded to `internal/ui` handlers.

## Usage

You primarily interact with this module by calling one of the factories and then mounting the `Router()`:

```go
auth.Router().ServeHTTP(w, r)
```
