# Proposal: AuthKnight Integration for `dracory/auth`

## Summary

Add native AuthKnight (https://authknight.com/) support to the `dracory/auth` package so developers can integrate with the AuthKnight hosted passwordless authentication service without writing custom callback handlers, HTTP client logic, or token verification code in every project.

---

## Background

### How AuthKnight Works

AuthKnight is a hosted passwordless auth provider. The flow is:

1. **Your app redirects** the user to `https://authknight.com/app/login?back_url=<your-cancel-url>&next_url=<your-callback-url>`
2. **User authenticates** via OAuth (Google, Facebook, or GitHub) on AuthKnight's servers
3. **AuthKnight redirects back** to your `next_url` with a `once` parameter (one-time token)
4. **Your app calls** `POST https://authknight.com/api/who` with the `once` token to verify
5. **AuthKnight returns** `{"status":"success","message":"success","data":{"email":"user@example.com","back_url":"..."}}` — only a validated email, no name fields
6. **Your app** checks if a user with that email already exists
7. **If existing user**: creates a session, sets auth cookie, redirects to success URL
8. **If new user**: stores the verified email in a temporary key, redirects to the register page where the email is pre-populated and read-only (already verified by AuthKnight). The user fills in first name / last name and submits the form — the user is created with all fields at once, same as the existing register flow

### Current Pain Point

Developers using `dracory/auth` currently must write ~300 lines of custom code per project to handle steps 3-8 (as seen in the example `authenticationController` implementation). This includes:
- Extracting the `once` parameter from the request
- Making the HTTP POST to `https://authknight.com/api/who`
- Parsing and validating the JSON response
- Finding or creating a user by email
- Creating a session and setting the auth cookie
- Handling the register page with pre-populated email
- Calculating redirect URLs

This proposal moves all of that into the library.

---

## Proposed Design

> **Prerequisite**: This proposal depends on `proposal_config.md` being completed first (extraction of `ConfigShared`, `ConfigRateLimiting`, `ConfigCSRF`, `ConfigImpersonation` reusable structs).

### 1. New Config Type: `ConfigAuthKnight`

A new config struct in `types/config_authknight.go`, embedding the shared structs:

```go
type ConfigAuthKnight struct {
    ConfigShared
    ConfigRateLimiting
    ConfigCSRF
    ConfigImpersonation

    // AuthKnight-specific

    // FuncUserFindByEmail finds a user by email and returns their ID.
    // Required.
    FuncUserFindByEmail func(ctx context.Context, email string, options UserAuthOptions) (userID string, err error)

    // FuncUserRegister creates a new user with email, first name, and last name.
    // The email is already verified by AuthKnight — no verification step needed.
    // Called when the user submits the register form (first name + last name).
    // Required.
    FuncUserRegister func(ctx context.Context, email string, firstName string, lastName string, options UserAuthOptions) (userID string, err error)

    // FuncRedirectURL calculates the redirect URL after successful auth.
    // If nil, falls back to UrlRedirectOnSuccess.
    FuncRedirectURL func(ctx context.Context, userID string) string

    // HTTPTimeout for calls to the AuthKnight API. Defaults to 10 seconds.
    HTTPTimeout time.Duration
}
```

### 2. New Interface: `AuthAuthKnightInterface`

In `types/auth_interfaces.go`:

```go
// AuthAuthKnightInterface represents AuthKnight-based authentication.
// It extends the shared interface with AuthKnight-specific helpers.
type AuthAuthKnightInterface interface {
    AuthSharedInterface

    // LinkAuthKnightLogin returns the AuthKnight-hosted login URL
    // with back_url and next_url parameters set.
    LinkAuthKnightLogin(backURL, nextURL string) string

    // LinkAuthKnightCallback returns the callback URL on this server
    // that AuthKnight will redirect to after successful authentication.
    LinkAuthKnightCallback() string
}
```

### 3. New Constructor: `NewAuthKnightAuth`

In `new_authknight_auth.go`:

```go
func NewAuthKnightAuth(config types.ConfigAuthKnight) (types.AuthAuthKnightInterface, error)
```

This follows the same pattern as `NewPasswordlessAuth` and `NewUsernameAndPasswordAuth`:
- Validates the config
- Creates an `authImplementation` with AuthKnight-specific fields
- Returns the typed interface

### 4. New Route Constants

In `constants.go`:

```go
// AuthKnightBaseURL is the base URL of the AuthKnight service.
AuthKnightBaseURL string = "https://authknight.com"

// PathApiAuthKnightCallback is the API endpoint that AuthKnight redirects to
// after successful authentication. It receives the "once" parameter.
PathApiAuthKnightCallback string = "api/authknight/callback"
```

### 5. New API Handler: `api_authknight_callback`

In `internal/api/api_authknight_callback/`:

This handler:
1. Extracts the `once` parameter from the request (query param or form body)
2. Calls `POST https://authknight.com/api/who` with the `once` token
3. Parses the JSON response: `{"status":"success","message":"success","data":{"email":"...","back_url":"..."}}`
4. Validates the `status == "success"` field
5. Calls `FuncUserFindByEmail` to find the user by email
6. **If user exists**: generates a random auth token, calls `FuncUserStoreAuthToken`, sets auth cookie, redirects to `UrlRedirectOnSuccess` (or calls `FuncRedirectURL` if set)
7. **If user not found**: generates a temporary key, stores the verified email via `FuncTemporaryKeySet` (expires in 1 hour), redirects to the register page at `PathRegister?ak_key=<temp_key>`

### 5b. Reuse Existing Register Page (`PathRegister`)

The existing register page (`internal/ui/page_register/page_register.go`) already has two branches:
- Passwordless: email + verification code
- Username/password: username, password, email

We add a **third branch for AuthKnight mode** that:
- Reads the `ak_key` query parameter
- Calls `FuncTemporaryKeyGet(ak_key)` to retrieve the verified email
- Renders a form with:
  - **Email** field — read-only, pre-populated with the verified email
  - **First name** field — text input
  - **Last name** field — text input
  - **Submit** button — POSTs to `PathApiRegister` (existing register API, with AuthKnight branch)

The form uses the same `FuncLayout` wrapper and `hb` HTML builder as the other branches. No new page route is needed — `PathRegister` is reused.

### 5c. Extend Existing `api_register` Handler

The existing `api_register` handler (`internal/api/api_register/api_register.go`) already has two branches (passwordless, username/password). We add a **third branch for AuthKnight mode** that:
1. Reads the `ak_key` from the form body
2. Calls `FuncTemporaryKeyGet(ak_key)` to retrieve the verified email
3. Extracts `first_name` and `last_name` from the form body
4. Validates that all fields are non-empty
5. Calls `FuncUserRegister(ctx, email, firstName, lastName, options)` to create the user with all fields at once
6. Generates a random auth token, calls `FuncUserStoreAuthToken`, sets auth cookie
7. Returns success response or redirects to `UrlRedirectOnSuccess`

The email is already verified by AuthKnight — no verification code or email sending step is needed. This matches the existing pattern where the user is created on the register page, not before it.

### 6. New Links

In `internal/links/links.go`:

```go
func AuthKnightCallback(endpoint string) string {
    return Join(endpoint, "api/authknight/callback")
}
```

### 7. Router Integration

In `router.go`, add routing for the new callback path:

```go
} else if uriHasPathSuffix(uri, PathApiAuthKnightCallback) {
    path = PathApiAuthKnightCallback
}
```

And in `getRoute` / `buildAPIRoutes`, register the handler.

### 8. AuthKnight Login URL Helper

The `LinkAuthKnightLogin(backURL, nextURL string) string` method builds:

```
https://authknight.com/app/login?back_url=<backURL>&next_url=<nextURL>
```

Where `nextURL` defaults to `LinkAuthKnightCallback()` (the callback endpoint on this server).

---

## Architecture Diagram

```
User Browser          Your App (dracory/auth)           AuthKnight.com
     |                        |                              |
     |  Click "Login"         |                              |
     |----------------------->|                              |
     |  302 Redirect to       |                              |
     |  LinkAuthKnightLogin() |                              |
     |<-----------------------|                              |
     |                        |                              |
     |  Redirect to AuthKnight login page                   |
     |----------------------------------------------------->|
     |                        |                              |
     |  User picks Google/FB/GitHub                         |
     |  OAuth flow completes                                |
     |                        |                              |
     |  302 Redirect back to next_url?once=<token>          |
     |<-----------------------------------------------------|
     |                        |                              |
     |  GET /auth/api/authknight/callback?once=<token>      |
     |----------------------->|                              |
     |                        |  POST /api/who?once=<token>  |
     |                        |----------------------------->|
     |                        |  {"status":"success",        |
     |                        |   "data":{"email":"..."}}    |
     |                        |<-----------------------------|
     |                        |                              |
     |                        |  FuncUserFindByEmail(email)  |
     |                        |                              |
     |  If user EXISTS:       |                              |
     |  Store auth token      |                              |
     |  Set cookie            |                              |
     |  302 Redirect to       |                              |
     |  UrlRedirectOnSuccess  |                              |
     |<-----------------------|                              |
     |  Authenticated!        |                              |
     |                        |                              |
     |  If user NOT FOUND:    |                              |
     |  Store verified email  |                              |
     |  in temp key           |                              |
     |  302 Redirect to       |                              |
     |  /auth/register        |                              |
     |  ?ak_key=<temp_key>    |                              |
     |<-----------------------|                              |
     |                        |                              |
     |  GET /auth/register?ak_key=<temp_key>                 |
     |----------------------->|                              |
     |  Register page         |                              |
     |  (AuthKnight branch):  |                              |
     |  email (read-only,     |                              |
     |   from temp key)       |                              |
     |  first_name (input)    |                              |
     |  last_name (input)     |                              |
     |<-----------------------|                              |
     |                        |                              |
     |  POST /auth/api/register                               |
     |  ak_key=<temp_key>     |                              |
     |  first_name=John       |                              |
     |  last_name=Doe         |                              |
     |----------------------->|                              |
     |                        |  FuncTemporaryKeyGet(ak_key) |
     |                        |  → verified email            |
     |                        |  FuncUserRegister(           |
     |                        |    email, firstName,         |
     |                        |    lastName)                 |
     |                        |  Store auth token            |
     |                        |  Set cookie                  |
     |  302 Redirect to       |                              |
     |  UrlRedirectOnSuccess  |                              |
     |<-----------------------|                              |
     |                        |                              |
     |  Authenticated!        |                              |
```

---

## Files to Create

| File | Purpose |
|------|---------|
| `types/config_authknight.go` | `ConfigAuthKnight` struct (embeds shared structs from `proposal_config.md`) |
| `types/config_authknight_test.go` | Config validation tests |
| `new_authknight_auth.go` | `NewAuthKnightAuth()` constructor |
| `new_authknight_auth_test.go` | Constructor tests |
| `internal/api/api_authknight_callback/api_authknight_callback.go` | Callback handler (extracts `once`, calls AuthKnight API, finds user by email, creates session or stores verified email in temp key and redirects to register) |
| `internal/api/api_authknight_callback/api_authknight_callback_test.go` | Handler tests with mocked HTTP |
| `internal/api/api_authknight_callback/constants.go` | Constants for the handler |
| `internal/api/api_authknight_callback/dependencies.go` | `Dependencies` struct (same pattern as other handlers) |
| `internal/ui/page_register/register_authknight.go` | AuthKnight branch content for the existing register page (email read-only from temp key, first name, last name) |
| `internal/ui/page_register/register_authknight_test.go` | Tests for AuthKnight register page branch |
| `examples/authknight/main.go` | Complete working example |
| `examples/authknight/main_test.go` | Example test |
| `examples/authknight/Taskfile.yml` | Task runner for the example |

## Files to Modify

| File | Change |
|------|--------|
| `types/auth_interfaces.go` | Add `AuthAuthKnightInterface` |
| `constants.go` | Add `PathApiAuthKnightCallback` |
| `internal/links/links.go` | Add `AuthKnightCallback()` link helper |
| `internal/links/links_test.go` | Add test for new link |
| `router.go` | Add routing for `PathApiAuthKnightCallback` |
| `auth_implementation.go` | Add AuthKnight-specific fields to `authImplementation` struct |
| `auth_implementation_api.go` | Add `apiAuthKnightCallback` method |
| `internal/api/api_register/api_register.go` | Add AuthKnight branch (third mode alongside passwordless and username/password) |
| `internal/api/api_register/api_register_test.go` | Add test for AuthKnight branch |
| `internal/ui/page_register/page_register.go` | Add AuthKnight branch (reads `ak_key`, shows email read-only) |
| `internal/ui/page_register/page_register_test.go` | Add test for AuthKnight branch |
| `public_api.go` | Optionally export `AuthKnightLoginURL()` helper |

---

## Usage Example (Target API)

```go
package main

import (
    "context"
    "fmt"
    "net/http"

    auth "github.com/dracory/auth"
    "github.com/dracory/auth/types"
)

func main() {
    authInstance, err := auth.NewAuthKnightAuth(types.ConfigAuthKnight{
        Endpoint:             "/auth",
        UrlRedirectOnSuccess: "/dashboard",
        UseCookies:           true,

        // User store callbacks (bring your own database)
        FuncUserFindByEmail: func(ctx context.Context, email string, opts types.UserAuthOptions) (string, error) {
            // Your DB lookup logic here
            return findUserByEmail(email)
        },
        FuncUserRegister: func(ctx context.Context, email, firstName, lastName string, opts types.UserAuthOptions) (string, error) {
            // Create user with email + first name + last name.
            // Email is already verified by AuthKnight — no verification step needed.
            return createUser(email, firstName, lastName)
        },
        FuncUserStoreAuthToken: func(ctx context.Context, token, userID string, opts types.UserAuthOptions) error {
            // Your session store logic here
            return storeSession(token, userID)
        },
        FuncUserFindByAuthToken: func(ctx context.Context, token string, opts types.UserAuthOptions) (string, error) {
            // Your session lookup logic here
            return findUserByToken(token)
        },
        FuncUserLogout: func(ctx context.Context, userID string, opts types.UserAuthOptions) error {
            // Your logout logic here
            return deleteSessions(userID)
        },
    })
    if err != nil {
        panic(err)
    }

    mux := http.NewServeMux()

    // Mount auth routes (includes /auth/api/authknight/callback)
    mux.HandleFunc("/auth/", authInstance.Router().ServeHTTP)

    // Login button redirects to AuthKnight
    mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r,
            authInstance.LinkAuthKnightLogin(
                "http://localhost:8083/",              // back_url
                authInstance.LinkAuthKnightCallback(),  // next_url
            ),
            http.StatusTemporaryRedirect,
        )
    })

    // Protected route
    mux.Handle("/dashboard", authInstance.WebAuthOrRedirectMiddleware(
        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := authInstance.GetCurrentUserID(r)
            fmt.Fprintf(w, "<h1>Dashboard</h1><p>User: %s</p>", userID)
        }),
    ))

    http.ListenAndServe(":8083", mux)
}
```

---

## Design Decisions

### Why a Separate Config/Interface (Not Extending Existing Ones)?

AuthKnight has a fundamentally different flow from the existing passwordless and username/password modes:
- No email sending needed (AuthKnight handles the OAuth flow)
- No verification codes (AuthKnight handles identity verification)
- Requires an outbound HTTP call to verify the `once` token
- The login page is hosted externally (not served by the library)

Mixing this into `ConfigPasswordless` or `ConfigUsernameAndPassword` would pollute those configs with irrelevant fields. A clean separation follows the existing pattern where each auth mode has its own config and interface.

### Why Keep the Same `authImplementation` Struct?

Following the existing architecture where a single `authImplementation` struct serves all modes. AuthKnight-specific fields are added as a new section, same as the passwordless and password sections. This avoids code duplication for shared functionality (cookies, rate limiting, CSRF, impersonation, middleware).

### Why Same-Phase Registration (User Created on Register Page, Not Before)?

AuthKnight returns only a validated email — no first name, no last name. The flow matches the existing library pattern where the user is created on the register page, not before it:

1. **Callback**: Verifies the `once` token, gets the email, checks if user exists by email.
   - **If user exists**: Creates session, sets cookie, redirects to success URL.
   - **If user not found**: Stores the verified email in a temporary key (using existing `FuncTemporaryKeySet`), redirects to the register page with `?ak_key=<temp_key>`.
2. **Register page**: Shows email (read-only, pre-populated from temp key) + first name + last name inputs. User fills in the form and submits.
3. **Register API**: Retrieves email from temp key, gets first_name + last_name from form, calls `FuncUserRegister(ctx, email, firstName, lastName, options)` to create the user with all fields at once. Creates session, sets cookie, redirects to success URL.

**Why not create the user in the callback?** Because the existing library pattern creates users on the register page with all fields at once. Creating an email-only user in the callback would introduce a new pattern (incomplete user records, profile completion step) that doesn't exist in the current library. By storing the verified email in a temp key and redirecting to the register page, we reuse the existing registration flow with minimal changes.

**Why use a temp key?** The verified email needs to be passed from the callback to the register page. Using `FuncTemporaryKeySet`/`FuncTemporaryKeyGet` (which the library already uses for verification codes) is the simplest mechanism. The temp key expires in 1 hour.

`FuncUserRegister` takes `(ctx, email, firstName, lastName, options)` and returns `(userID, error)` — similar to the existing passwordless `FuncUserRegister` which takes `(ctx, email, firstName, lastName, options)` but returns `error`. The AuthKnight version returns the `userID` so the handler can create a session immediately.

### HTTP Client Design

The handler will use `net/http` with:
- Configurable timeout (default 10s, matching the existing implementation)
- Context propagation for cancellation
- Proper response body closing
- JSON decoding with error handling

---

## Testing Strategy

### Unit Tests

- **Config validation**: `ConfigAuthKnight` validation in `types/config_authknight_test.go`
- **Constructor**: `NewAuthKnightAuth` in `new_authknight_auth_test.go`
- **Callback handler**: `api_authknight_callback_test.go` with:
  - Mock HTTP server simulating AuthKnight API responses
  - Success flow (valid `once` token, existing user → session created, redirect to success URL)
  - Success flow (valid `once` token, new user → verified email stored in temp key, redirect to register page with `ak_key`)
  - Failure flow (invalid `once` token)
  - Failure flow (missing `once` parameter)
  - Failure flow (AuthKnight API unreachable)
  - Failure flow (user not found, `FuncUserFindByEmail` returns error)
- **Register handler (AuthKnight branch)**: `api_register_test.go` with:
  - Success flow (valid `ak_key` + first_name + last_name → user created with all fields, session created, redirect to success URL)
  - Failure flow (missing `ak_key`)
  - Failure flow (invalid/expired `ak_key`)
  - Failure flow (missing first_name)
  - Failure flow (missing last_name)
  - Failure flow (FuncUserRegister returns error)
- **Links**: `AuthKnightCallback` link in `links_test.go`
- **Router**: Route matching for `PathApiAuthKnightCallback` in `router_test.go`
- **Login URL builder**: `LinkAuthKnightLogin` with various `back_url`/`next_url` combinations

### Integration Test

- `examples/authknight/main_test.go` — end-to-end test with a mock AuthKnight server

### Test Coverage Target

Maintain the existing ~90% coverage standard. All new files must have corresponding test files.

---

## Migration Path

This is a purely additive change. No existing code is modified in a breaking way:
- No changes to `ConfigPasswordless` or `ConfigUsernameAndPassword`
- No changes to `AuthPasswordInterface` or `AuthPasswordlessInterface`
- New interface, config, constructor, and routes are added alongside existing ones
- Existing users of the library are unaffected

---

## Resolved Questions

1. **Callback response format** — Keep what works now (redirect). No need for JSON response mode.

2. **Multiple AuthKnight instances** — Not needed. One instance per `authImplementation`, matching existing pattern.

3. **Rate limiting on callback** — Not needed for now. Left for future enhancement.

4. **Additional fields on register page** — Out of scope for this proposal. Will be addressed in a separate proposal.

---

## Implementation Order

1. `types/config_authknight.go` + tests
2. `types/auth_interfaces.go` — add `AuthAuthKnightInterface`
3. `constants.go` — add `AuthKnightBaseURL`, `PathApiAuthKnightCallback`
4. `internal/links/links.go` — add `AuthKnightCallback()` + tests
5. `auth_implementation.go` — add AuthKnight fields to struct
6. `internal/api/api_authknight_callback/` — callback handler + dependencies + tests
7. `internal/ui/page_register/page_register.go` — add AuthKnight branch (reads `ak_key`, shows email read-only)
8. `internal/ui/page_register/register_authknight.go` — AuthKnight branch content + tests
9. `internal/api/api_register/api_register.go` — add AuthKnight branch (creates user with all fields)
10. `internal/api/api_register/api_register_test.go` — add AuthKnight branch tests
11. `auth_implementation_api.go` — wire `apiAuthKnightCallback` method
12. `router.go` — add route for `PathApiAuthKnightCallback`
13. `new_authknight_auth.go` + tests
14. `examples/authknight/` — complete example
15. `public_api.go` — export helpers if needed
16. Run full test suite: `go test -v ./...`
17. Check coverage: `go test -cover ./...`
