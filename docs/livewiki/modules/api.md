---
Created: 2025-12-06
Last Updated: 2025-12-06
Version: 1.0.0
---

# API Module

**Package**: `internal/api`

This module contains the logic for all JSON endpoints. It is split into sub-packages for cleaner organization.

## Structure

Each major action has its own directory:

*   `api_login`: Handles password checking and Magic Link generation.
*   `api_login_code_verify`: Handles verifying the code sent via email.
*   `api_register`: Handles user creation logic.
*   `api_logout`: Handles clearing the session.

## Design Pattern

Each handler follows a similar pattern:
1.  **Parse Payload**: Decode JSON request body.
2.  **Validate**: Check for required fields.
3.  **Execute Logic**: Call the relevant callback functions (e.g., `FuncUserFindByEmail`).
4.  **Respond**: Return JSON success or error using `api_shared.RespondError` or `api_shared.RespondSuccess`.

## Error Handling

Errors are standardized. Internal errors (like DB connection failure) are logged internally but returned as generic `INTERNAL_ERROR` to the user to prevent information leakage.
