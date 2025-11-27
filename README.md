# Auth <a href="https://gitpod.io/#https://github.com/dracory/auth" style="float:right;"><img src="https://gitpod.io/button/open-in-gitpod.svg" alt="Open in Gitpod" loading="lazy"></a>

[![Tests Status](https://github.com/dracory/auth/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/dracory/auth/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dracory/auth)](https://goreportcard.com/report/github.com/dracory/auth)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/dracory/auth)](https://pkg.go.dev/github.com/dracory/auth)

<img src="logo.jpg" width="100%" />

**Batteries-included authentication library for Go** with ready-to-use UI pages, API endpoints, and middleware. You bring your own database and email service—we handle the rest.

## ✨ Features

- 🔐 **Two Authentication Flows**
  - **Passwordless** - Email-based verification codes (recommended for security)
  - **Username/Password** - Traditional authentication with password storage
  
- 🎨 **Complete UI Included**
  - Pre-built HTML pages (login, registration, password reset)
  - Bootstrap-styled and customizable
  - Works out of the box

- 🚀 **JSON API Endpoints**
  - Ready for SPAs and mobile apps
  - RESTful design
  - Comprehensive error handling

- 🛡️ **Authentication Middleware**
  - `WebAuthOrRedirectMiddleware` - For web pages
  - `ApiAuthOrErrorMiddleware` - For API routes
  - `WebAppendUserIdIfExistsMiddleware` - Optional authentication

- 🔧 **Implementation Agnostic**
  - Works with any database (SQL, NoSQL, in-memory)
  - Bring your own email service
  - Callback-based architecture for maximum flexibility

- 🚦 **Built-in Rate Limiting**
  - Per-IP and per-endpoint limits on authentication endpoints
  - Sensible defaults (5 attempts per 15 minutes, 15-minute lockout)
  - Fully configurable or replaceable with a custom rate limiter
 
- ✅ **Production Ready**
  - 90%+ test coverage
  - 34 comprehensive test files
  - Battle-tested in production

## 📦 Installation

```sh
go get github.com/dracory/auth
```

## 🚀 Quick Start

### Choose Your Flow

<table>
<tr>
<th>Passwordless (Recommended)</th>
<th>Username/Password</th>
</tr>
<tr>
<td>

```go
auth, err := auth.NewPasswordlessAuth(
  auth.ConfigPasswordless{
    Endpoint: "/auth",
    UrlRedirectOnSuccess: "/dashboard",
    UseCookies: true,
    // ... implement callbacks
  },
)
```

</td>
<td>

```go
auth, err := auth.NewUsernameAndPasswordAuth(
  auth.ConfigUsernameAndPassword{
    Endpoint: "/auth",
    UrlRedirectOnSuccess: "/dashboard",
    UseCookies: true,
    // ... implement callbacks
  },
)
```

</td>
</tr>
</table>

### Attach to Router

```go
mux := http.NewServeMux()

// Attach auth routes
mux.HandleFunc("/auth/", auth.Router().ServeHTTP)

// Public route
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Welcome! <a href='" + auth.LinkLogin() + "'>Login</a>"))
})

// Protected route
mux.Handle("/dashboard", auth.WebAuthOrRedirectMiddleware(dashboardHandler))
```

### Get Current User

```go
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
    userID := auth.GetCurrentUserID(r)
    // Use userID to fetch user data from your database
    fmt.Fprintf(w, "Welcome, user %s!", userID)
}
```

## 📚 Complete Examples

### Passwordless Flow

#### Step 1: Implement Required Functions

```go
// Email sending
func emailSend(email string, subject string, body string) error {
    // Use your email service (SendGrid, AWS SES, SMTP, etc.)
    return yourEmailService.Send(email, subject, body)
}

// User lookup by email
func userFindByEmail(email string, options auth.UserAuthOptions) (userID string, err error) {
    // Query your database
    user, err := db.Query("SELECT id FROM users WHERE email = ?", email)
    if err != nil {
        return "", err
    }
    return user.ID, nil
}

// User registration (optional, if EnableRegistration is true)
func userRegister(email string, firstName string, lastName string, options auth.UserAuthOptions) error {
    // Insert into your database
    _, err := db.Exec("INSERT INTO users (email, first_name, last_name) VALUES (?, ?, ?)", 
        email, firstName, lastName)
    return err
}

// User logout
func userLogout(userID string, options auth.UserAuthOptions) error {
    // Remove token from your session/cache store
    return sessionStore.Delete("auth_token_" + userID)
}

// Token storage
func userStoreAuthToken(token string, userID string, options auth.UserAuthOptions) error {
    // Store in session/cache with expiration (e.g., 2 hours)
    return sessionStore.Set("auth_token_"+token, userID, 2*time.Hour)
}

// Token lookup
func userFindByAuthToken(token string, options auth.UserAuthOptions) (userID string, err error) {
    // Retrieve from session/cache
    userID, err = sessionStore.Get("auth_token_" + token)
    return userID, err
}

// Temporary key storage (for verification codes)
func tempKeySet(key string, value string, expiresSeconds int) error {
    // Store temporarily (e.g., in Redis, cache, or database)
    return cacheStore.Set(key, value, time.Duration(expiresSeconds)*time.Second)
}

func tempKeyGet(key string) (value string, err error) {
    // Retrieve temporary key
    return cacheStore.Get(key)
}
```

#### Step 2: Configure Authentication

```go
authInstance, err := auth.NewPasswordlessAuth(auth.ConfigPasswordless{
    // Required
    Endpoint:                "/auth",
    UrlRedirectOnSuccess:    "/dashboard",
    UseCookies:              true, // OR UseLocalStorage: true
    FuncUserFindByAuthToken: userFindByAuthToken,
    FuncUserFindByEmail:     userFindByEmail,
    FuncUserLogout:          userLogout,
    FuncUserStoreAuthToken:  userStoreAuthToken,
    FuncEmailSend:           emailSend,
    FuncTemporaryKeyGet:     tempKeyGet,
    FuncTemporaryKeySet:     tempKeySet,
    
    // Optional
    EnableRegistration:           true,
    FuncUserRegister:             userRegister,
    FuncEmailTemplateLoginCode:   customLoginEmailTemplate,    // optional
    FuncEmailTemplateRegisterCode: customRegisterEmailTemplate, // optional
    FuncLayout:                   customPageLayout,             // optional
    DisableRateLimit:             false,                        // optional
    MaxLoginAttempts:             5,                            // optional
    LockoutDuration:              15 * time.Minute,             // optional
    FuncCheckRateLimit:           nil,                          // optional
})
```

#### Step 3: Setup Routes

```go
mux := http.NewServeMux()

// Auth routes
mux.HandleFunc("/auth/", authInstance.Router().ServeHTTP)

// Public routes
mux.HandleFunc("/", homeHandler)

// Protected routes (web)
mux.Handle("/dashboard", authInstance.WebAuthOrRedirectMiddleware(dashboardHandler))
mux.Handle("/profile", authInstance.WebAuthOrRedirectMiddleware(profileHandler))

// Protected routes (API)
mux.Handle("/api/data", authInstance.ApiAuthOrErrorMiddleware(apiDataHandler))

// Optional auth (works for both authenticated and guest users)
mux.Handle("/products", authInstance.WebAppendUserIdIfExistsMiddleware(productsHandler))

http.ListenAndServe(":8080", mux)
```

### Username/Password Flow

#### Step 1: Implement Required Functions

```go
// User login with password verification
func userLogin(username string, password string, options auth.UserAuthOptions) (userID string, err error) {
    // Query database and verify password (use bcrypt or similar)
    user, err := db.Query("SELECT id, password_hash FROM users WHERE email = ?", username)
    if err != nil {
        return "", err
    }
    
    if !bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)) {
        return "", errors.New("invalid credentials")
    }
    
    return user.ID, nil
}

// User registration with password
func userRegister(username string, password string, firstName string, lastName string, options auth.UserAuthOptions) error {
    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    
    // Insert into database
    _, err = db.Exec("INSERT INTO users (email, password_hash, first_name, last_name) VALUES (?, ?, ?, ?)",
        username, hashedPassword, firstName, lastName)
    return err
}

// User lookup by username
func userFindByUsername(username string, firstName string, lastName string, options auth.UserAuthOptions) (userID string, err error) {
    // Query database (firstName and lastName used for password reset verification)
    user, err := db.Query("SELECT id FROM users WHERE email = ? AND first_name = ? AND last_name = ?",
        username, firstName, lastName)
    if err != nil {
        return "", err
    }
    return user.ID, nil
}

// Password change
func userPasswordChange(username string, newPassword string, options auth.UserAuthOptions) error {
    // Hash new password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    
    // Update in database
    _, err = db.Exec("UPDATE users SET password_hash = ? WHERE email = ?", hashedPassword, username)
    return err
}

// Other functions same as passwordless (userLogout, userStoreAuthToken, etc.)
```

#### Step 2: Configure Authentication

```go
authInstance, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
    // Required
    Endpoint:                "/auth",
    UrlRedirectOnSuccess:    "/dashboard",
    UseCookies:              true,
    FuncUserFindByAuthToken: userFindByAuthToken,
    FuncUserFindByUsername:  userFindByUsername,
    FuncUserLogin:           userLogin,
    FuncUserLogout:          userLogout,
    FuncUserStoreAuthToken:  userStoreAuthToken,
    FuncEmailSend:           emailSend,
    FuncTemporaryKeyGet:     tempKeyGet,
    FuncTemporaryKeySet:     tempKeySet,
    
    // Optional
    EnableRegistration:               true,
    EnableVerification:               true, // Require email verification
    FuncUserRegister:                 userRegister,
    FuncUserPasswordChange:           userPasswordChange,
    FuncEmailTemplatePasswordRestore: customPasswordResetEmailTemplate, // optional
    FuncLayout:                       customPageLayout,                  // optional
    DisableRateLimit:                 false,                             // optional
    MaxLoginAttempts:                 5,                                  // optional
    LockoutDuration:                  15 * time.Minute,                   // optional
    FuncCheckRateLimit:               nil,                                // optional
})
```

## 🔌 Available Endpoints

Once configured, the following endpoints are automatically available:

### API Endpoints (JSON responses)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/api/login` | Initiate login (sends code for passwordless) |
| POST | `/auth/api/login-code-verify` | Verify passwordless login code |
| POST | `/auth/api/logout` | Logout user |
| POST | `/auth/api/register` | Initiate registration |
| POST | `/auth/api/register-code-verify` | Verify registration code |
| POST | `/auth/api/restore-password` | Request password reset |
| POST | `/auth/api/reset-password` | Complete password reset |

### Page Endpoints (HTML responses)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/auth/login` | Login page |
| GET | `/auth/login-code-verify` | Code verification page |
| GET | `/auth/logout` | Logout page |
| GET | `/auth/register` | Registration page |
| GET | `/auth/register-code-verify` | Registration verification page |
| GET | `/auth/password-restore` | Password restore request page |
| GET | `/auth/password-reset?t=TOKEN` | Password reset page |

## 🛡️ Middleware Options

### WebAuthOrRedirectMiddleware

For web pages - redirects to login if not authenticated:

```go
mux.Handle("/dashboard", auth.WebAuthOrRedirectMiddleware(dashboardHandler))
```

### ApiAuthOrErrorMiddleware

For API endpoints - returns JSON error if not authenticated:

```go
mux.Handle("/api/profile", auth.ApiAuthOrErrorMiddleware(profileHandler))
```

### WebAppendUserIdIfExistsMiddleware

Optional authentication - adds userID to context if authenticated, but doesn't redirect/error:

```go
mux.Handle("/products", auth.WebAppendUserIdIfExistsMiddleware(productsHandler))

func productsHandler(w http.ResponseWriter, r *http.Request) {
    userID := auth.GetCurrentUserID(r)
    if userID != "" {
        // Show personalized products
    } else {
        // Show public products
    }
}
```

## 🎨 Customization

### Custom Email Templates

```go
func customLoginEmailTemplate(email string, code string, options auth.UserAuthOptions) string {
    return fmt.Sprintf(`
        <h1>Your Login Code</h1>
        <p>Hi %s,</p>
        <p>Your verification code is: <strong>%s</strong></p>
        <p>This code will expire in 1 hour.</p>
    `, email, code)
}

// Use in config
FuncEmailTemplateLoginCode: customLoginEmailTemplate,
```

### Custom Page Layout

```go
func customPageLayout(content string) string {
    return fmt.Sprintf(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>My App - Authentication</title>
            <link rel="stylesheet" href="/css/custom.css">
        </head>
        <body>
            <div class="container">
                %s
            </div>
        </body>
        </html>
    `, content)
}

// Use in config
FuncLayout: customPageLayout,
```

## 🔐 Token Storage Options

### Cookies (Recommended for web apps)

```go
UseCookies: true,
UseLocalStorage: false,
```

- Automatically set and sent with requests
- HttpOnly for security
- Works with server-side rendering

### LocalStorage (For SPAs)

```go
UseCookies: false,
UseLocalStorage: true,
```

- Client manages token storage
- Must send token in Authorization header
- Better for single-page applications

## 🚦 Rate Limiting

All authentication endpoints (login, registration, password restore/reset, verification) are protected by rate limiting.

**Defaults (in-memory limiter):**

- 5 attempts per IP and endpoint within a 15-minute sliding window
- Further attempts are blocked for 15 minutes (HTTP 429 with `Retry-After` header)

These options are shared by both `ConfigPasswordless` and `ConfigUsernameAndPassword`:

```go
// Rate limiting options (shared by both configs)
DisableRateLimit   bool                                                                                 // Set to true to disable rate limiting (not recommended for production)
FuncCheckRateLimit func(ip string, endpoint string) (allowed bool, retryAfter time.Duration, err error) // Optional: override default rate limiter
MaxLoginAttempts   int                                                                                  // Maximum attempts before lockout (default: 5)
LockoutDuration    time.Duration                                                                        // Duration for sliding window and lockout (default: 15 minutes)
```

**Example (username/password):**

```go
authInstance, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
    Endpoint:         "/auth",
    UrlRedirectOnSuccess: "/dashboard",

    MaxLoginAttempts: 5,
    LockoutDuration:  15 * time.Minute,

    // Optional: use your own distributed rate limiter (e.g., Redis-based)
    // FuncCheckRateLimit: func(ip, endpoint string) (bool, time.Duration, error) {
    //     // ... implement custom logic
    // },
})
```

## 📖 UserAuthOptions

All callback functions receive `UserAuthOptions` with request context:

```go
type UserAuthOptions struct {
    UserIp    string  // Client IP address
    UserAgent string  // Client user agent
}
```

Use this for audit logging, security checks, or analytics:

```go
func userLogin(username string, password string, options auth.UserAuthOptions) (userID string, err error) {
    // Log login attempt
    log.Printf("Login attempt from IP: %s, UserAgent: %s", options.UserIp, options.UserAgent)
    
    // Your login logic...
}
```

## 🔍 Helper Methods

```go
// Get current authenticated user ID from request context
userID := auth.GetCurrentUserID(r)

// URL helpers
loginURL := auth.LinkLogin()
registerURL := auth.LinkRegister()
logoutURL := auth.LinkLogout()
passwordRestoreURL := auth.LinkPasswordRestore()
passwordResetURL := auth.LinkPasswordReset(token)

// API URL helpers
apiLoginURL := auth.LinkApiLogin()
apiRegisterURL := auth.LinkApiRegister()
apiLogoutURL := auth.LinkApiLogout()

// Enable/disable registration dynamically
auth.RegistrationEnable()
auth.RegistrationDisable()
```

## ❓ Frequently Asked Questions

**Q: Can I use email instead of username?**  
A: Yes! The "username" parameter accepts email addresses. Most modern apps use email for authentication.

**Q: Can I run multiple auth instances?**  
A: Yes! You can have separate instances for different user types:

```go
// Regular users with passwordless
userAuth, _ := auth.NewPasswordlessAuth(...)
mux.HandleFunc("/auth/", userAuth.Router().ServeHTTP)

// Admins with username/password
adminAuth, _ := auth.NewUsernameAndPasswordAuth(...)
mux.HandleFunc("/admin/auth/", adminAuth.Router().ServeHTTP)
```

**Q: How do I customize the UI?**  
A: Provide a custom `FuncLayout` function to wrap the content with your own HTML/CSS.

**Q: What databases are supported?**  
A: Any! You implement the storage callbacks, so it works with PostgreSQL, MySQL, MongoDB, Redis, or even in-memory stores.

**Q: Is this production-ready?**  
A: Yes! The library has 90%+ test coverage and is used in production applications.

**Q: How do I handle password reset?**  
A: The library includes built-in password reset flow. Users enter their email, receive a reset link, and set a new password.

**Q: Can I use this with an existing user system?**  
A: Absolutely! Just implement the callback functions to integrate with your existing database schema.

## 🧪 Testing

The library includes comprehensive tests. Run them with:

```sh
go test -v ./...
```

For coverage:

```sh
go test -cover ./...
```

## 📝 Working Example

Check the [development](./development) directory for a complete working example with:
- Both authentication flows
- JSON file storage (Scribble)
- Email sending
- All callbacks implemented

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

See [LICENSE](LICENSE) file for details.

## 🔗 Related Projects

- [authorizerdev/authorizer](https://github.com/authorizerdev/authorizer) - Open source authentication and authorization
- [markbates/goth](https://github.com/markbates/goth) - Multi-provider authentication
- [teamhanko/hanko](https://github.com/teamhanko/hanko) - Passwordless authentication
- [go-pkgz/auth](https://github.com/go-pkgz/auth) - Authentication service

---

Made with ❤️ by the Dracory team
