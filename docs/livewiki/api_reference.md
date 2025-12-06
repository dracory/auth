---
Created: 2025-12-06
Last Updated: 2025-12-06
Version: 1.0.0
---

# API Reference

These endpoints are automatically provided when you mount the Auth router.

## Authentication Endpoints

All endpoints return JSON responses.

### Login & Registration

| Method | Endpoint | Description | Payload |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/login` | Initiate login flow. Sends magic link or verifies password. | `{ "username": "...", "password": "..." }` |
| `POST` | `/api/login-code-verify` | Verify a login code (Passwordless). | `{ "email": "...", "code": "..." }` |
| `POST` | `/api/register` | Register a new user. | `{ "email": "...", "password": "...", ... }` |
| `POST` | `/api/register-code-verify` | Verify registration email. | `{ "email": "...", "code": "..." }` |

### Session Management

| Method | Endpoint | Description | Payload |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/logout` | Invalidate current session. | (Empty) |

### Password Management

| Method | Endpoint | Description | Payload |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/restore-password` | Request password reset email. | `{ "email": "..." }` |
| `POST` | `/api/reset-password` | Set new password using token. | `{ "token": "...", "password": "..." }` |

## Error Responses

Errors are returned in a standard JSON format:

```json
{
  "error": "Error message description",
  "code": "ERROR_CODE"
}
```

Common error codes:
*   `INVALID_CREDENTIALS`
*   `RATE_LIMIT_EXCEEDED`
*   `INTERNAL_ERROR`
