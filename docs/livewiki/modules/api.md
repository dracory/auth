---
path: modules/api.md
page-type: module
summary: API module documentation for internal/api endpoint handlers, design patterns, and error handling.
tags: [module, api, internal, handlers, json, rest]
created: 2025-12-06
updated: 2026-07-16
version: 2.0.0
---

# API Module

**Package**: `internal/api`

This module contains the logic for all JSON endpoints. Each endpoint has its own sub-package.

## Structure

*   `api_login`: Password checking and magic link generation.
*   `api_login_code_verify`: Verifying login codes (Passwordless).
*   `api_register`: User creation logic.
*   `api_register_code_verify`: Registration code verification.
*   `api_logout`: Session clearing.
*   `api_password_reset`: Setting new password with token.
*   `api_password_restore`: Requesting password reset email.
*   `api_authenticate_via_username`: Username-based authentication.
*   `api_impersonate_start`: Starting impersonation session.
*   `api_impersonate_stop`: Stopping impersonation and restoring original session.
*   `api_authknight_callback`: AuthKnight OAuth callback handler.

## Design Pattern

Each handler follows a similar pattern:
1.  **Parse Payload**: Decode JSON request body.
2.  **Validate**: Check for required fields.
3.  **Execute Logic**: Call relevant callback functions.
4.  **Respond**: Return JSON success or error.

## Middleware Chain

API routes are wrapped with:
1.  **Rate Limit Middleware** — all API endpoints
2.  **CSRF Middleware** — selected endpoints (login, register, password reset, impersonation)

## Error Handling

Errors are standardized. Internal errors are logged via `slog` but returned as generic messages to prevent information leakage. Error messages are centralized in `types/messages.go`.

## See Also

- [Core Module](core.md)
- [API Reference](../api_reference.md)
- [Data Flow](../data_flow.md)

## Changelog
- **v2.0.0** (2026-07-16): Added impersonation, AuthKnight callback, authenticate_via_username handlers. Added middleware chain docs.
- **v1.0.0** (2025-12-04): Initial creation
