---
Created: 2025-12-06
Last Updated: 2025-12-06
Version: 1.0.0
---

# Architecture

Dracory Auth is designed as a **Library**, not a Service. It is embedded directly into your Go application, running in the same process.

## High-Level Design

The library is split into three main layers:

1.  **Transport Layer**: HTTP Handlers and Router (Gin/Echo/Stdlib compatible via `net/http`).
2.  **Logic Layer**: Authentication flows (Passwordless, Username/Password), Session management, Security (CSRF, Rate Limiting).
3.  **Data Layer (Callbacks)**: Interfaces you implement to connect to your database and external services.

```mermaid
graph TD
    User[User / Client] -->|HTTP Request| App[Your Application]
    App -->|Mounts| AuthLib[Dracory Auth Library]
    
    subgraph "Auth Library"
        Router[Router]
        Logic[Auth Logic]
        Mid[Middleware]
    end
    
    AuthLib --> Router
    Router --> Logic
    
    subgraph "Your Implementation"
        DB_Adapter[DB Callbacks]
        Email_Adapter[Email Callbacks]
    end
    
    Logic -->|Calls| DB_Adapter
    Logic -->|Calls| Email_Adapter
    
    DB_Adapter -->|SQL/NoSQL| DB[(Your Database)]
    Email_Adapter -->|SMTP/API| Email[(Email Provider)]
```

## Dependency Injection

The library uses a form of dependency injection via the Configuration structs (`ConfigPasswordless`, `ConfigUsernameAndPassword`). You inject the behavior (functions) rather than the data.

## Security Architecture

*   **Tokens**: Randomly generated, stored via callbacks.
*   **Cookies**: `HttpOnly`, `Secure`, `SameSite`.
*   **Rate Limiting**: In-memory token bucket algorithm, keyed by IP and Endpoint.
*   **CSRF**: Double Submit Cookie pattern (via `dracory/csrf`).
