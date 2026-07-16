---
path: development.md
page-type: reference
summary: Development workflow, testing, project structure, and contribution guidelines for Dracory Auth.
tags: [development, testing, contributing, project-structure, go]
created: 2025-12-06
updated: 2026-07-16
version: 2.0.0
---

# Development

## Prerequisites

*   Go 1.26.3+
*   Git

## Running Tests

Run the full test suite:

```bash
go test -v ./...
```

For coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

With race detection:

```bash
go test -race -v ./...
```

## Project Structure

```
auth/
├── *.go                    # Root package: constructors, router, implementation
├── types/                  # Config structs, interfaces, constants, observability
├── middlewares/            # Public auth middlewares (redirect, error, append, impersonation)
├── utils/                  # Cookies, token retrieval, rate limiter, password strength, scribble
├── internal/
│   ├── api/                # API endpoint handlers (one package per endpoint)
│   ├── core/               # Login/register business logic
│   ├── emails/             # Email template generators
│   ├── helpers/            # Layout, rate limit helpers
│   ├── links/              # URL link generation
│   ├── middlewares/        # Internal CSRF and rate limit middleware wrappers
│   ├── testutils/          # Shared test helpers
│   └── ui/                 # HTML page handlers (one package per page)
├── examples/               # Example apps: passwordless, usernamepassword, authknight
└── docs/livewiki/          # This documentation
```

## Key Files

| File | Purpose |
| :--- | :--- |
| `auth_implementation.go` | Core `authImplementation` struct |
| `router.go` | `AuthHandler` routing logic |
| `new_passwordless_auth.go` | Passwordless constructor |
| `new_username_and_password_auth.go` | Username/Password constructor |
| `new_authknight_auth.go` | AuthKnight constructor |
| `public_api.go` | Backward-compatible public shortcuts |
| `impersonation_helpers.go` | `IsImpersonating`, `GetImpersonatorUserID` |
| `types/auth_interfaces.go` | All auth interfaces |
| `types/config_*.go` | Configuration structs |
| `types/constants.go` | Path, endpoint, and context key constants |

## Contributing

1.  Fork the repository.
2.  Create a feature branch.
3.  Add tests for your changes.
4.  Ensure `go test ./...` passes.
5.  Submit a Pull Request.

## Changelog
- **v2.0.0** (2026-07-16): Updated Go version to 1.26.3, added full project structure, key files table
- **v1.0.0** (2025-12-04): Initial creation
