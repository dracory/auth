---
path: api_reference.md
page-type: reference
summary: Complete API endpoint documentation for all authentication, session, password, impersonation, and AuthKnight flows.
tags: [api, endpoints, rest, authentication, impersonation, authknight]
created: 2025-12-06
updated: 2026-07-16
version: 2.0.0
---

# API Reference

These endpoints are automatically provided when you mount the Auth router. Available endpoints depend on which auth flow and features are enabled.

## Authentication Endpoints

All endpoints return JSON responses.

### Login & Registration

| Method | Endpoint | Description | Payload | CSRF |
| :--- | :--- | :--- | :--- | :--- |
| `POST` | `/api/login` | Initiate login flow. Password check or code generation. | `{ "username": "...", "password": "..." }` | Yes |
| `POST` | `/api/login-code-verify` | Verify a login code (Passwordless). | `{ "email": "...", "code": "..." }` | No |
| `POST` | `/api/register` | Register a new user. | `{ "email": "...", "password": "...", "first_name": "...", "last_name": "..." }` | Yes |
| `POST` | `/api/register-code-verify` | Verify registration email. | `{ "email": "...", "code": "..." }` | No |

### Session Management

| Method | Endpoint | Description | Payload | CSRF |
| :--- | :--- | :--- | :--- | :--- |
| `POST` | `/api/logout` | Invalidate current session. | (Empty) | No |

### Password Management

| Method | Endpoint | Description | Payload | CSRF |
| :--- | :--- | :--- | :--- | :--- |
| `POST` | `/api/restore-password` | Request password reset email. | `{ "email": "..." }` | No |
| `POST` | `/api/reset-password` | Set new password using token. | `{ "token": "...", "password": "..." }` | Yes |

### Impersonation (opt-in)

Available only when `EnableImpersonation` is `true` in config.

| Method | Endpoint | Description | Payload | CSRF |
| :--- | :--- | :--- | :--- | :--- |
| `POST` | `/api/impersonate/start` | Start impersonating a target user. | `{ "target_user_id": "..." }` | Yes |
| `POST` | `/api/impersonate/stop` | Stop impersonating and restore original session. | (Empty) | Yes |

### AuthKnight Callback

Available only when using `NewAuthKnightAuth`.

| Method | Endpoint | Description | Payload | CSRF |
| :--- | :--- | :--- | :--- | :--- |
| `POST` | `/api/authknight/callback` | Callback endpoint for AuthKnight OAuth flow. | `{ "once": "...", "email": "...", "first_name": "...", "last_name": "..." }` | No |

## Rate Limiting

All API endpoints are wrapped with rate limiting middleware. When the limit is exceeded, the response is `429 Too Many Requests` with a `Retry-After` header.

Default limits: 5 attempts per 15-minute window per IP/endpoint.

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
*   `IMPERSONATION_NOT_ENABLED`
*   `IMPERSONATION_FORBIDDEN`
*   `AUTHKNIGHT_VERIFICATION_FAILED`

## Changelog
- **v2.0.0** (2026-07-16): Added impersonation endpoints, AuthKnight callback endpoint, CSRF column, rate limiting docs
- **v1.0.0** (2025-12-04): Initial creation
