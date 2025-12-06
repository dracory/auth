---
Created: 2025-12-06
Last Updated: 2025-12-06
Version: 1.0.0
---

# Getting Started

This guide will help you integrate Dracory Auth into your Go application.

## Installation

```bash
go get github.com/dracory/auth
```

## Quick Setup

### 1. Choose Your Flow

Decide between **Passwordless** or **Username/Password**.

### 2. Implement Callbacks

You must implement a few core functions to handle data storage. For example, finding a user by email:

```go
func userFindByEmail(ctx context.Context, email string, options types.UserAuthOptions) (userID string, err error) {
    // Your DB logic here
    return "user-123", nil
}
```

### 3. Initialize Auth

```go
import "github.com/dracory/auth"

authInstance, err := auth.NewPasswordlessAuth(types.ConfigPasswordless{
    Endpoint: "/auth",
    UrlRedirectOnSuccess: "/dashboard",
    UseCookies: true,
    FuncUserFindByEmail: userFindByEmail,
    // ... other callbacks
})
```

### 4. Mount Routes

```go
mux := http.NewServeMux()
mux.HandleFunc("/auth/", authInstance.Router().ServeHTTP)
http.ListenAndServe(":8080", mux)
```

## Running Examples

The best way to learn is to run the provided examples:

```bash
cd examples/passwordless
go run main.go
```

Then visit `http://localhost:8080`.
