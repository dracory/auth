---
path: table_of_contents.md
page-type: overview
summary: Table of contents for the Dracory Auth LiveWiki, listing all documentation pages.
tags: [table-of-contents, navigation, livewiki]
created: 2025-12-06
updated: 2026-07-16
version: 2.0.0
---

# LiveWiki Table of Contents

Welcome to the LiveWiki for the Dracory Auth library. This documentation serves as a comprehensive guide for developers working with and contributing to the codebase.

## Core Documentation

1.  **[Overview](overview.md)**
    *   High-level introduction to the purpose and capabilities of the Auth library.
2.  **[Getting Started](getting_started.md)**
    *   Installation, setup, and running examples for all three auth flows.
3.  **[Architecture](architecture.md)**
    *   Design, three auth flows, layered architecture, interfaces, security.
4.  **[API Reference](api_reference.md)**
    *   JSON API endpoints for login, registration, impersonation, AuthKnight callback.
5.  **[Data Flow](data_flow.md)**
    *   Request flow diagrams: login, passwordless, protected routes, impersonation, AuthKnight.
6.  **[Configuration](configuration.md)**
    *   All config structs: ConfigShared, ConfigPasswordless, ConfigUsernameAndPassword, ConfigAuthKnight, CookieConfig.
7.  **[Development](development.md)**
    *   Workflow, testing, project structure, and contribution details.
8.  **[Troubleshooting](troubleshooting.md)**
    *   Common issues and how to resolve them.
9.  **[LLM Context](llm-context.md)**
    *   Single-file codebase summary optimized for AI/LLM consumption.

## Module Documentation

Detailed breakdown of the internal structure:

*   **[Core Module](modules/core.md)**: The root `auth` package and main entry points.
*   **[API Internals](modules/api.md)**: Implementation of API handlers (`internal/api`).
*   **[UI Internals](modules/ui.md)**: Implementation of Page handlers (`internal/ui`).
*   **[Middlewares](modules/middlewares.md)**: Built-in auth middleware.
*   **[Types](modules/types.md)**: Config structs, interfaces, constants, observability.
*   **[Utils](modules/utils.md)**: Cookies, rate limiter, password strength, token retrieval.
*   **[AuthKnight](modules/authknight.md)**: AuthKnight OAuth flow implementation.
*   **[Impersonation](modules/impersonation.md)**: Impersonation feature implementation.

## Changelog
- **v2.0.0** (2026-07-16): Added Types, Utils, AuthKnight, Impersonation module pages. Updated all descriptions.
- **v1.0.0** (2025-12-04): Initial creation
