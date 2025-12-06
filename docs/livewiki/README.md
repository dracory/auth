---
Created: 2025-12-06
Last Updated: 2025-12-06
Version: 1.0.0
---

# Overview

The **Dracory Auth** service is a batteries-included authentication library for Go. It provides ready-to-use UI pages, API endpoints, and middleware, while allowing you to bring your own database and email service.

## Key Features

*   **Dual Authentication Flows**: Supports both modern **Passwordless** (Magic Link/Code) and traditional **Username/Password** flows.
*   **Implementation Agnostic**: You provide the storage and email logic via callbacks. The library handles the rest.
*   **Secure by Default**: Includes CSRF protection, rate limiting, secure headers, and rigorous input validation.
*   **Complete UI**: Pre-built, Bootstrap-styled pages for login, registration, and password reset.
*   **JSON API**: Full REST API support for SPAs and mobile applications.

## How It Works

Unlike full-stack auth platforms (like Auth0) or framework-specific plugins, this library sits in the middle. It handles the **HTTP layer** (routes, parsing, validation, error handling) and delegates the **Persistence layer** (database, cache, email) to your application via a simple interface.

This design allows:
1.  **Flexibility**: Use any database (SQL, NoSQL) or email provider (SendGrid, SES).
2.  **Control**: You own the user data schema.
3.  **Simplicity**: Drop it into any Go Standard Library `net/http` compatible router.
