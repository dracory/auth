---
path: getting_started.md
page-type: tutorial
summary: Setup guide for integrating Dracory Auth with Passwordless, Username/Password, or AuthKnight flows.
tags: [getting-started, tutorial, setup, auth, passwordless, password, authknight]
created: 2025-12-06
updated: 2026-07-16
version: 2.0.0
---

# Getting Started

This guide will help you integrate Dracory Auth into your Go application.

## Installation

```bash
go get github.com/dracory/auth
```

## Quick Setup

### 1. Choose Your Flow

Decide between **Passwordless**, **Username/Password**, or **AuthKnight**.

### 2. Implement Callbacks

You must implement core functions to handle data storage. For example, finding a user by email:

```go
func userFindByEmail(ctx context.Context, email string, options types.UserAuthOptions) (userID string, err error) {
    // Your DB logic here
    return "user-123", nil
}
```

### 3. Initialize Auth

#### Passwordless

```go
import "github.com/dracory/auth"

authInstance, err := auth.NewPasswordlessAuth(types.ConfigPasswordless{
    Endpoint: "/auth",
    UrlRedirectOnSuccess: "/dashboard",
    UseCookies: true,
    FuncUserFindByEmail: userFindByEmail,
    FuncEmailSend: emailSend,
    FuncUserRegister: userRegister,
    FuncTemporaryKeyGet: tempKeyGet,
    FuncTemporaryKeySet: tempKeySet,
    FuncUserStoreAuthToken: storeAuthToken,
    FuncUserFindByAuthToken: findByAuthToken,
    FuncUserLogout: userLogout,
})
```

#### Username/Password

```go
authInstance, err := auth.NewUsernameAndPasswordAuth(types.ConfigUsernameAndPassword{
    Endpoint: "/auth",
    UrlRedirectOnSuccess: "/dashboard",
    UseCookies: true,
    FuncUserLogin: userLogin,
    FuncUserRegister: userRegister,
    FuncUserFindByUsername: findByUsername,
    FuncEmailSend: emailSend,
    FuncTemporaryKeyGet: tempKeyGet,
    FuncTemporaryKeySet: tempKeySet,
    FuncUserStoreAuthToken: storeAuthToken,
    FuncUserFindByAuthToken: findByAuthToken,
    FuncUserLogout: userLogout,
})
```

#### AuthKnight

```go
authInstance, err := auth.NewAuthKnightAuth(types.ConfigAuthKnight{
    Endpoint: "/auth",
    UrlRedirectOnSuccess: "/dashboard",
    UseCookies: true,
    FuncUserFindByEmail: userFindByEmail,
    FuncUserRegister: userRegister,
    FuncTemporaryKeyGet: tempKeyGet,
    FuncTemporaryKeySet: tempKeySet,
    FuncUserStoreAuthToken: storeAuthToken,
    FuncUserFindByAuthToken: findByAuthToken,
    FuncUserLogout: userLogout,
})
```

### 4. Mount Routes

```go
mux := http.NewServeMux()
mux.HandleFunc("/auth/", authInstance.Router().ServeHTTP)
http.ListenAndServe(":8080", mux)
```

### 5. Optional: Enable CSRF, Impersonation, Observability

```go
config := types.ConfigUsernameAndPassword{
    // ... core fields ...
    EnableCSRFProtection: true,
    CSRFSecret: "your-secret-key",
    EnableImpersonation: true,
    FuncCanImpersonate: canImpersonate,
    FuncImpersonationStart: impersonateStart,
    FuncImpersonationStop: impersonateStop,
    ObservabilityHooks: myMetricsHooks{},
}
```

## Running Examples

The best way to learn is to run the provided examples:

```bash
# Passwordless
cd examples/passwordless
go run main.go

# Username/Password
cd examples/usernamepassword
go run main.go

# AuthKnight
cd examples/authknight
go run main.go
```

Then visit `http://localhost:8080`.

## Changelog
- **v2.0.0** (2026-07-16): Added AuthKnight setup, CSRF/impersonation/observability examples, updated callback signatures
- **v1.0.0** (2025-12-04): Initial creation
