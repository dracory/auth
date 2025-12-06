---
Created: 2025-12-06
Last Updated: 2025-12-06
Version: 1.0.0
---

# Data Flow

Understanding how data moves through the system helps in debugging and customization.

## 1. Authentication Request Flow

When a user attempts to log in:

1.  **Client** sends `POST /api/login`.
2.  **Rate Limiter** checks IP/Endpoint. If limit exceeded -> 429 Too Many Requests.
3.  **Router** routes to `internal/api/api_login`.
4.  **Handler** parses JSON body.
5.  **Logic** calls `FuncUserFindByEmail` (your callback).
    *   If user not found -> Return generic error (prevent enumeration).
6.  **Logic** validates credentials (password check or code generation).
7.  **Logic** calls `FuncUserStoreAuthToken`.
8.  **Handler** sets `HttpOnly` cookie.
9.  **Handler** returns JSON success.

```mermaid
sequenceDiagram
    participant User
    participant AuthLib
    participant YourDB
    
    User->>AuthLib: POST /api/login
    AuthLib->>AuthLib: Check Rate Limit
    AuthLib->>YourDB: FuncUserFindByEmail(email)
    YourDB-->>AuthLib: User Object / Error
    
    alt User Found & Valid
        AuthLib->>YourDB: FuncUserStoreAuthToken(token)
        AuthLib-->>User: 200 OK (Set-Cookie)
    else Invalid
        AuthLib-->>User: 401 Unauthorized
    end
```

## 2. Protected Route Flow

When an authenticated user requests a protected resource:

1.  **Client** sends `GET /dashboard` (with Cookie).
2.  **Middleware** (`WebAuthOrRedirectMiddleware`) intercepts request.
3.  **Middleware** extracts token from Cookie.
4.  **Middleware** calls `FuncUserFindByAuthToken`.
    *   If valid: Adds User ID to Context. calls `next.ServeHTTP`.
    *   If invalid: Redirects to `/login`.
