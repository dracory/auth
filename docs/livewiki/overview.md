---
path: overview.md
page-type: overview
summary: High-level introduction to the Dracory Auth library, its three authentication flows, and key features.
tags: [overview, auth, go, authentication, passwordless, password, authknight]
created: 2025-12-06
updated: 2026-07-16
version: 2.0.0
---

# Overview

The **Dracory Auth** library is a batteries-included authentication solution for Go (1.26.3+). It provides ready-to-use UI pages, JSON API endpoints, and middleware, while allowing you to bring your own database and email service.

## Key Features

*   **Three Authentication Flows**: Supports **Passwordless** (email verification codes), **Username/Password** (traditional auth with password reset), and **AuthKnight** (OAuth-style hosted login via AuthKnight.com).
*   **Impersonation**: Opt-in admin "login as" feature with secure token swap and audit hooks.
*   **Observability Hooks**: Optional metrics/tracing integration via `ObservabilityHooks` interface (Prometheus, Datadog, OpenTelemetry, etc.).
*   **Customizable Cookies**: Functional options pattern for cookie configuration (`WithSecure`, `WithHttpOnly`, `WithSameSite`, `WithMaxAge`, `WithDomain`, `WithPath`, `WithCookieName`).
*   **Implementation Agnostic**: You provide the storage and email logic via callbacks. The library handles the rest.
*   **Secure by Default**: CSRF protection (`dracory/csrf`), rate limiting (in-memory, per-IP/per-endpoint), secure cookie defaults (HttpOnly, SameSite, Secure), and rigorous input validation.
*   **Complete UI**: Pre-built, Bootstrap-styled pages for login, registration, password reset, and code verification.
*   **JSON API**: Full REST API support for SPAs and mobile applications.
*   **Structured Logging**: Uses `log/slog` throughout with configurable logger.

## How It Works

Unlike full-stack auth platforms (like Auth0) or framework-specific plugins, this library sits in the middle. It handles the **HTTP layer** (routes, parsing, validation, error handling) and delegates the **Persistence layer** (database, cache, email) to your application via a simple callback interface.

This design allows:
1.  **Flexibility**: Use any database (SQL, NoSQL) or email provider (SendGrid, SES).
2.  **Control**: You own the user data schema.
3.  **Simplicity**: Drop it into any Go Standard Library `net/http` compatible router.

## Authentication Flows

### Passwordless (`NewPasswordlessAuth`)
Email-based verification codes. Users enter their email, receive a code, and verify it. No passwords required.

### Username/Password (`NewUsernameAndPasswordAuth`)
Traditional auth with username (email) and password. Includes password reset/restore flows and configurable password strength validation.

### AuthKnight (`NewAuthKnightAuth`)
OAuth-style hosted login via [AuthKnight.com](https://authknight.com). Users are redirected to AuthKnight for authentication, then called back to your application. Email is pre-verified by AuthKnight.

## Changelog
- **v2.0.0** (2026-07-16): Added AuthKnight flow, impersonation, observability hooks, cookie configuration, structured logging. Updated Go to 1.26.3.
- **v1.0.0** (2025-12-04): Initial creation
