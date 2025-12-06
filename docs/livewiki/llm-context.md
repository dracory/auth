---
path: llm-context.md
page-type: overview
summary: Complete codebase summary of Dracory Auth optimized for LLM consumption.
tags: [llm, context, summary, auth, go]
created: 2025-12-06
updated: 2025-12-06
version: 1.0.0
---

# LLM Context: Dracory Auth

## Project Summary

Dracory Auth is a batteries-included authentication library for Go. It provides ready-to-use UI pages, API endpoints, and middleware while allowing you to bring your own database and email service. It supports both modern Passwordless (Magic Link/Code) and traditional Username/Password authentication flows.

Unlike full-stack auth platforms (like Auth0) or framework-specific plugins, this library handles the HTTP layer (routes, parsing, validation, error handling) and delegates the persistence layer (database, cache, email) to your application via a simple callback interface.

## Key Technologies

- **Language**: Go (1.21+)
- **HTTP**: Standard `net/http` compatible
- **UI**: Bootstrap-styled HTML pages
- **API**: JSON REST endpoints

## Directory Structure

```
auth/
├── auth.go              # Main entry point, New() function
├── config_*.go          # Configuration structs for auth modes
├── middleware_*.go      # Authentication middleware
├── internal/
│   ├── api/             # JSON API handlers
│   └── ui/              # HTML page handlers
├── docs/
│   └── livewiki/        # This documentation
└── tests/               # Test files
```

## Core Concepts

1. **Configuration Objects**: `ConfigPasswordless` and `ConfigUsernameAndPassword` define which auth mode to use
2. **Callback Interface**: You provide functions for user lookup, creation, and email sending
3. **Middleware**: Protect routes by requiring authentication
4. **Dual Flows**: Same library supports both passwordless and traditional auth

## Common Patterns

- **Callback Pattern**: The library calls your functions for persistence operations
- **Interface Injection**: Pass implementations via config, not hardcoded dependencies
- **Middleware Chain**: Use `AuthMiddleware()` to wrap protected handlers
- **Router Integration**: Works with any `http.Handler` compatible router

## Important Files

| File | Purpose |
|------|---------|
| `auth.go` | Main `New()` function, primary entry point |
| `config_passwordless.go` | Passwordless mode configuration |
| `config_username_password.go` | Username/password mode configuration |
| `middleware_auth.go` | Authentication middleware implementation |
| `internal/api/handlers.go` | JSON API endpoint handlers |
| `internal/ui/pages.go` | HTML page handlers |

## See Also

- [Overview](overview.md) - High-level architecture
- [Getting Started](getting_started.md) - Setup guide
- [API Reference](api_reference.md) - Endpoint documentation
