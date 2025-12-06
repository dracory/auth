---
Created: 2025-12-06
Last Updated: 2025-12-06
Version: 1.0.0
---

# LiveWiki Table of Contents

Welcome to the LiveWiki for the Dracory Auth Service. This documentation serves as a comprehensive guide for developers working with and contributing to the codebase.

## Core Documentation

1.  **[Overview](overview.md)**
    *   High-level introduction to the purpose and capabilities of the Auth service.
2.  **[Getting Started](getting_started.md)**
    *   Installation, setup, and running your first example.
3.  **[Architecture](architecture.md)**
    *   Design verification, callback patterns, and system structure.
4.  **[API Reference](api_reference.md)**
    *   Details on the JSON API endpoints for login, registration, etc.
5.  **[Data Flow](data_flow.md)**
    *   How requests move through the system, from router to middleware to logic.
6.  **[Configuration](configuration.md)**
    *   Understanding `ConfigPasswordless` and `ConfigUsernameAndPassword`.
7.  **[Development](development.md)**
    *   Workflow, testing, and contribution details.
8.  **[Troubleshooting](troubleshooting.md)**
    *   Common issues and how to resolve them.

## Module Documentation

Detailed breakdown of the internal structure:

*   **[Core Module](modules/core.md)**: The root `auth` package and main entry points.
*   **[API Internals](modules/api.md)**: Implementation of API handlers (`internal/api`).
*   **[UI Internals](modules/ui.md)**: Implementation of Page handlers (`internal/ui`).
*   **[Middlewares](modules/middlewares.md)**: Built-in auth middleware.
