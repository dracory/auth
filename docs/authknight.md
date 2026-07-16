# AuthKnight Integration

AuthKnight provides passwordless authentication via a hosted login page. The user enters their email on AuthKnight's site, receives a verification code, and upon success AuthKnight redirects back to your application with a one-time token. Your server verifies the token with AuthKnight's API and receives the verified email.

## How It Works

1. User clicks "Login with AuthKnight" on your site
2. Browser redirects to `https://authknight.com/app/login` with `back_url` (cancel) and `next_url` (callback)
3. User authenticates on AuthKnight (email + verification code)
4. AuthKnight redirects to your `next_url` with a `once` token
5. Your server verifies the `once` token with AuthKnight's API
6. If the user exists, they're logged in. If not, they're redirected to the register page with a temporary key containing the verified email.

## Step 1: Implement Required Functions

```go
// User lookup by email (required)
func userFindByEmail(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
    user, err := db.Query("SELECT id FROM users WHERE email = ?", email)
    if err != nil {
        return "", err
    }
    if user == nil {
        return "", fmt.Errorf("user with email %s not found", email)
    }
    return user.ID, nil
}

// User registration (required) — email is already verified by AuthKnight
func userRegister(ctx context.Context, email string, firstName string, lastName string, options types.UserAuthOptions) (string, error) {
    result, err := db.Exec("INSERT INTO users (email, first_name, last_name) VALUES (?, ?, ?)",
        email, firstName, lastName)
    if err != nil {
        return "", err
    }
    id, _ := result.LastInsertId()
    return fmt.Sprintf("%d", id), nil
}

// Token storage (same as other flows)
func userStoreAuthToken(ctx context.Context, token string, userID string, options types.UserAuthOptions) error {
    return sessionStore.Set("auth_token_"+token, userID, 2*time.Hour)
}

func userFindByAuthToken(ctx context.Context, token string, options types.UserAuthOptions) (string, error) {
    return sessionStore.Get("auth_token_" + token)
}

func userLogout(ctx context.Context, userID string, options types.UserAuthOptions) error {
    return sessionStore.Delete("auth_token_" + userID)
}

// Temporary key storage (for storing verified email during registration)
func tempKeySet(key string, value string, expiresSeconds int) error {
    return cacheStore.Set(key, value, time.Duration(expiresSeconds)*time.Second)
}

func tempKeyGet(key string) (string, error) {
    return cacheStore.Get(key)
}
```

## Step 2: Configure Authentication

```go
authInstance, err := auth.NewAuthKnightAuth(types.ConfigAuthKnight{
    ConfigShared: types.ConfigShared{
        Endpoint:                "/auth",
        UrlRedirectOnSuccess:    "/dashboard",
        UseCookies:              true,
        EnableRegistration:      true,
        FuncUserFindByAuthToken: userFindByAuthToken,
        FuncUserLogout:          userLogout,
        FuncUserStoreAuthToken:  userStoreAuthToken,
        FuncTemporaryKeyGet:     tempKeyGet,
        FuncTemporaryKeySet:     tempKeySet,
        CookieConfig: &types.CookieConfig{
            Secure:   types.CookieInsecure, // use CookieSecure in production
            HttpOnly: types.CookieHttpOnly,
            SameSite: http.SameSiteLaxMode,
            Path:     "/",
            MaxAge:   2 * 60 * 60,
        },
    },

    // AuthKnight-specific
    FuncUserFindByEmail: userFindByEmail,
    FuncUserRegister:    userRegister,

    // Optional: custom redirect URL after auth (defaults to UrlRedirectOnSuccess)
    // FuncRedirectURL: func(ctx context.Context, userID string) string {
    //     return "/dashboard?welcome=1"
    // },

    // Optional: timeout for AuthKnight API calls (defaults to 10 seconds)
    // HTTPTimeout: 10 * time.Second,
})
```

### ConfigAuthKnight Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `ConfigShared` | `types.ConfigShared` | Yes | Shared config (endpoint, cookies, callbacks) |
| `FuncUserFindByEmail` | `func(ctx, email, options) (string, error)` | Yes | Find user by email |
| `FuncUserRegister` | `func(ctx, email, firstName, lastName, options) (string, error)` | Yes | Register new user (email already verified) |
| `FuncRedirectURL` | `func(ctx, userID) string` | No | Custom redirect URL after auth. Falls back to `UrlRedirectOnSuccess` |
| `HTTPTimeout` | `time.Duration` | No | Timeout for AuthKnight API calls. Defaults to 10 seconds |

## Step 3: Setup Routes

```go
mux := http.NewServeMux()

// Auth routes (login, register, callback, etc.)
mux.HandleFunc("/auth/", authInstance.Router().ServeHTTP)

// Public home page with AuthKnight login link
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    loginURL := authInstance.LinkAuthKnightRedirect(r)
    fmt.Fprintf(w, "<h1>Home</h1><p><a href='%s'>Login with AuthKnight</a></p>", loginURL)
})

// Protected routes
mux.Handle("/dashboard", authInstance.WebAuthOrRedirectMiddleware(dashboardHandler))

http.ListenAndServe(":8084", mux)
```

## URL Helpers

```go
// Build the full AuthKnight login URL from the request (scheme + host)
loginURL := authInstance.LinkAuthKnightRedirect(r)

// Or build manually with custom back_url and next_url
loginURL := authInstance.LinkAuthKnightLogin(cancelURL, callbackURL)

// Get the callback path (relative to endpoint)
callbackPath := authInstance.LinkAuthKnightCallback()
```

### `back_url` vs `next_url`

- **`back_url`** — Cancel/fallback URL. Where AuthKnight sends the user if they cancel login.
- **`next_url`** — Callback URL. Where AuthKnight redirects after successful authentication, sending the `once` token.

`LinkAuthKnightRedirect(r)` derives both automatically from the request:
- `back_url` = `scheme://host` + endpoint (e.g. `http://localhost:8084/auth`)
- `next_url` = `scheme://host` + callback path (e.g. `http://localhost:8084/auth/api/authknight/callback`)

## Page Behavior in AuthKnight Mode

| Route | Behavior |
|-------|----------|
| `/auth/login` | Redirects to AuthKnight's hosted login page |
| `/auth/register` (without `ak_key`) | Redirects to AuthKnight's hosted login page |
| `/auth/register?ak_key=xxx` | Shows registration form with verified email as static text |
| `/auth/password-restore` | Redirects to AuthKnight login |
| `/auth/password-reset` | Redirects to AuthKnight login |
| `/auth/login-code-verify` | Redirects to AuthKnight login |
| `/auth/register-code-verify` | Redirects to AuthKnight login |
| `/auth/api/authknight/callback` | Internal endpoint that AuthKnight redirects to after auth |

## Security: Verified Email Protection

The verified email on the registration page is displayed as **static text** (not an input field) to prevent browser extensions like LastPass from modifying it. The email used for registration is always retrieved from the secure temporary key store, not from the form submission.

### How it works:

1. AuthKnight verifies the email and sends it to your callback
2. Your server stores the email in a temporary key (`ak_key` → email)
3. User is redirected to `/auth/register?ak_key=xxx`
4. The register page retrieves the email from the temporary key store (server-side)
5. The email is displayed as static text with a hidden input for form submission
6. The register API handler also retrieves the email from the temporary key store, not from the form

This means even if a browser extension or manual manipulation changes the form's `email` field, the actual email used for registration comes from the server-side temporary key store.

## Endpoints

### API Endpoint

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/auth/api/authknight/callback` | AuthKnight redirect endpoint (receives `once` token) |

### Page Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/auth/login` | Redirects to AuthKnight login |
| GET | `/auth/register` | Registration page (redirects without `ak_key`) |
| GET | `/auth/logout` | Logout page |

## Working Example

A complete working example is available in [`examples/authknight`](../examples/authknight):

```sh
cd examples/authknight
go run main.go
```

Available at `http://localhost:8084`.
