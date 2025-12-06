---
Created: 2025-12-06
Last Updated: 2025-12-06
Version: 1.0.0
---

# Configuration

The library is configured via structs passed to `NewPasswordlessAuth` or `NewUsernameAndPasswordAuth`.

## ConfigPasswordless

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `Endpoint` | `string` | Yes | Base URL path (e.g., `/auth`). |
| `UrlRedirectOnSuccess` | `string` | Yes | Where to go after login. |
| `UseCookies` | `bool` | Yes | Enable cookie-based sessions. |
| `FuncUserFindByEmail` | `func` | Yes | Callback to find user. |
| `FuncEmailSend` | `func` | Yes | Callback to send emails. |
| `FuncTemporaryKeySet` | `func` | Yes | Callback to store temp codes. |

## ConfigUsernameAndPassword

Includes all fields from Passwordless (except code-specific ones) plus:

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `FuncUserLogin` | `func` | Yes | Logic to verify password. |
| `FuncUserPasswordChange` | `func` | Optional | Logic to update password. |

## Rate Limiting Options

Both configs share:

*   `DisableRateLimit` (bool): Turn off for testing.
*   `MaxLoginAttempts` (int): Defaults to 5.
*   `LockoutDuration` (duration): Defaults to 15m.
