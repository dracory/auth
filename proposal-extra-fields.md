# Proposal: Optional Extra Registration Fields

## Summary

Add support for optional extra fields on the built-in registration pages:
- `business_name`
- `phone`
- `country`
- `timezone`
- `birthday`

These fields are opt-in — developers enable them via config. When enabled, the fields appear on the registration form (both passwordless and username/password), are submitted to the API, and are passed through to the developer's `FuncUserRegister` callback via a new `RegistrationExtraFields` struct.

---

## Current State

### Registration Form Fields Today

| Mode | Fields |
|------|--------|
| Passwordless | `first_name`, `last_name`, `email` |
| Username/Password | `first_name`, `last_name`, `email`, `password` |

### Current Data Flow

1. **UI** (`internal/ui/page_register/content.go`) — builds HTML form with `first_name`, `last_name`, `email` (and `password` for username/password mode)
2. **JS** — validates required fields, POSTs JSON to `api/register`
3. **API handler** (`internal/api/api_register/api_register.go:63-66`) — extracts fields via `req.GetStringTrimmed(r, "field_name")`
4. **Core** (`internal/core/register_with_username_and_password.go`) — validates, calls `FuncUserRegister(ctx, email, password, firstName, lastName, options)`
5. **Developer callback** — receives typed parameters, stores in their own DB

### Current `FuncUserRegister` Signatures

**Username/Password:**
```go
func(ctx context.Context, username string, password string, first_name string, last_name string, options UserAuthOptions) error
```

**Passwordless:**
```go
func(ctx context.Context, email string, firstName string, lastName string, options UserAuthOptions) error
```

Both pass `first_name` and `last_name` as explicit positional parameters. There is no mechanism for additional fields.

---

## Proposed Design

### 1. New Struct: `RegistrationExtraFields`

In `types/registration_extra_fields.go`:

```go
package types

// RegistrationExtraFields holds optional fields that may be collected during
// registration. All fields are optional — the developer enables them via config
// and the library renders them on the form if enabled.
type RegistrationExtraFields struct {
    BusinessName string
    Phone        string
    Country      string
    Timezone     string
    Birthday     string // format: "2006-01-02" (Go date format)
}
```

### 2. New Config Struct: `RegistrationFieldsConfig`

In `types/registration_fields_config.go`:

```go
package types

// RegistrationFieldsConfig controls which optional fields appear on the
// registration form. All fields default to false (hidden). Set to true to
// render the field on the form and include it in the API payload.
type RegistrationFieldsConfig struct {
    EnableBusinessName bool
    EnablePhone        bool
    EnableCountry      bool
    EnableTimezone     bool
    EnableBirthday     bool
}
```

### 3. Add `RegistrationFields` to Existing Configs

Add a new field to both `ConfigUsernameAndPassword` and `ConfigPasswordless`:

```go
// RegistrationFields controls which optional fields appear on the
// registration form. nil means no extra fields (current behavior).
RegistrationFields *RegistrationFieldsConfig
```

### 4. Extend `UserAuthOptions` with Extra Fields

In `types/user_auth_options.go`:

```go
type UserAuthOptions struct {
    UserIp    string
    UserAgent string

    // ExtraFields holds optional registration fields (business_name, phone,
    // country, timezone, birthday). Populated only during registration.
    // nil when not applicable (e.g., login, logout, password reset).
    ExtraFields *RegistrationExtraFields
}
```

**Why `UserAuthOptions`?** This struct is already passed to every user callback (`FuncUserRegister`, `FuncUserLogin`, `FuncUserLogout`, etc.). Adding `ExtraFields` here means the developer's existing callback signature doesn't change — they just read `options.ExtraFields` if they want the extra data. This is backward compatible: existing callbacks that don't read `ExtraFields` continue to work.

### 5. Extend `AuthSharedInterface`

Add accessors for the registration fields config:

```go
// In types/auth_interfaces.go, add to AuthSharedInterface:

GetRegistrationFields() *RegistrationFieldsConfig
SetRegistrationFields(cfg *RegistrationFieldsConfig)
```

### 6. UI Changes — `internal/ui/page_register/content.go`

Both `RegisterPasswordlessContent` and `RegisterUsernameAndPasswordContent` gain a new parameter:

```go
// Current signature:
func RegisterPasswordlessContent(urlLogin string) string

// New signature:
func RegisterPasswordlessContent(urlLogin string, fieldsConfig *types.RegistrationFieldsConfig) string
```

When `fieldsConfig` is non-nil and a field is enabled, the function appends the corresponding form group to the card body:

| Field | HTML Element | Input Type | Name Attribute | Placeholder |
|-------|-------------|------------|----------------|-------------|
| `business_name` | `<input class="form-control" name="business_name" placeholder="Enter business name">` | text | `business_name` | "Enter business name" |
| `phone` | `<input class="form-control" name="phone" placeholder="Enter phone number">` | tel | `phone` | "Enter phone number" |
| `country` | `<select class="form-control" name="country">...</select>` | select | `country` | — (dropdown) |
| `timezone` | `<select class="form-control" name="timezone">...</select>` | select | `timezone` | — (dropdown) |
| `birthday` | `<input class="form-control" name="birthday" type="date">` | date | `birthday` | — |

**Country dropdown** — populated with a static list of countries (ISO 3166-1 country names). The list will be defined as a Go slice constant in `internal/ui/page_register/countries.go`.

**Timezone dropdown** — populated with a static list of common timezones (e.g., "UTC", "America/New_York", "Europe/London", etc.). The list will be defined as a Go slice constant in `internal/ui/page_register/timezones.go`.

**Birthday** — uses the native HTML5 `<input type="date">` picker. No custom date picker library needed.

### 7. JS Changes — Form Validation and Submission

The `registerFormValidate()` function in both `RegisterPasswordlessScripts` and `RegisterUsernameAndPasswordScripts` is updated to:

1. **Read** the extra field values from the DOM (only if the fields exist in the form)
2. **Validate** — all extra fields are optional, so no required-field checks. However, basic format validation is added:
   - `phone`: must be non-empty if filled (no strict format — too many international formats)
   - `birthday`: must be a valid date if filled
3. **Include** the extra fields in the POST data JSON payload

Example updated POST data for username/password mode:

```javascript
var data = {
    "first_name": first_name,
    "last_name": last_name,
    "email": email,
    "password": password,
    "business_name": business_name,  // only if field exists
    "phone": phone,                  // only if field exists
    "country": country,              // only if field exists
    "timezone": timezone,            // only if field exists
    "birthday": birthday             // only if field exists
};
```

The JS will detect which fields exist in the form via `$('input[name=business_name]').length > 0` checks and only include them if present.

### 8. API Handler Changes — `internal/api/api_register/api_register.go`

The handler extracts the extra fields from the request and populates `UserAuthOptions.ExtraFields`:

```go
// After existing field extraction (email, password, firstName, lastName):

extraFields := &types.RegistrationExtraFields{}
hasExtra := false

if v := req.GetStringTrimmed(r, "business_name"); v != "" {
    extraFields.BusinessName = html.EscapeString(v)
    hasExtra = true
}
if v := req.GetStringTrimmed(r, "phone"); v != "" {
    extraFields.Phone = v
    hasExtra = true
}
if v := req.GetStringTrimmed(r, "country"); v != "" {
    extraFields.Country = v
    hasExtra = true
}
if v := req.GetStringTrimmed(r, "timezone"); v != "" {
    extraFields.Timezone = v
    hasExtra = true
}
if v := req.GetStringTrimmed(r, "birthday"); v != "" {
    extraFields.Birthday = v
    hasExtra = true
}

options := types.UserAuthOptions{
    UserIp:    ip,
    UserAgent: userAgent,
}
if hasExtra {
    options.ExtraFields = extraFields
}
```

The `options` struct is then passed through to `deps.RegisterWithUsernameAndPassword` → `core.RegisterWithUsernameAndPassword` → `registerFn(ctx, email, password, firstName, lastName, options)`.

**No changes to `FuncUserRegister` signature.** The extra fields travel inside `options.ExtraFields`.

### 9. Core Registration Changes — `internal/core/register_with_username_and_password.go`

No changes needed. The core function already receives `options UserAuthOptions` and passes it to `registerFn`. The extra fields are already inside `options`.

### 10. Page Handler Changes — `internal/ui/page_register/page_register.go`

Pass the fields config from the auth instance to the content builder:

```go
func PageRegister(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
    fieldsConfig := a.GetRegistrationFields()

    if a.IsPasswordless() {
        content = RegisterPasswordlessContent(links.Login(a.GetEndpoint()), fieldsConfig)
        // ...
    } else {
        content = RegisterUsernameAndPasswordContent(
            links.Login(a.GetEndpoint()),
            links.PasswordRestore(a.GetEndpoint()),
            fieldsConfig,
        )
        // ...
    }
    // ...
}
```

### 11. `authImplementation` Struct

Add the field to `auth_implementation.go`:

```go
type authImplementation struct {
    // ... existing fields ...
    registrationFields *types.RegistrationFieldsConfig
}
```

Add getter/setter methods following the existing pattern.

### 12. Constructor Changes

In `new_username_and_password_auth.go` and `new_passwordless_auth.go`, wire the config:

```go
impl.registrationFields = config.RegistrationFields
```

---

## Architecture Diagram

```
Developer Config                    UI Form                         API Handler
     |                                |                                |
     |  RegistrationFields: {         |                                |
     |    EnablePhone: true,          |                                |
     |    EnableCountry: true,        |                                |
     |  }                             |                                |
     |                                |                                |
     |-----> authImplementation ----->|                                |
     |      .registrationFields       |                                |
     |                                |                                |
     |                    PageRegister(w, r, a)                        |
     |                    a.GetRegistrationFields()                    |
     |                                |                                |
     |                    RegisterUsernameAndPasswordContent(          |
     |                      urlLogin, urlPasswordRestore,              |
     |                      fieldsConfig                               |
     |                    )                                           |
     |                                |                                |
     |                    Renders form with:                          |
     |                      first_name (required)                     |
     |                      last_name (required)                      |
     |                      email (required)                          |
     |                      password (required, if applicable)         |
     |                      phone (optional, if enabled)              |
     |                      country (optional, if enabled)            |
     |                                |                                |
     |                    JS registerFormValidate()                   |
     |                    POSTs JSON with all fields                  |
     |                                |------------------------------->|
     |                                |                                |
     |                                |    req.GetStringTrimmed(r,    |
     |                                |      "phone")                  |
     |                                |    req.GetStringTrimmed(r,    |
     |                                |      "country")                |
     |                                |                                |
     |                                |    options.ExtraFields =       |
     |                                |      &RegistrationExtraFields{ |
     |                                |        Phone: "...",           |
     |                                |        Country: "...",         |
     |                                |      }                         |
     |                                |                                |
     |                                |    registerFn(ctx, email,      |
     |                                |      password, firstName,      |
     |                                |      lastName, options)        |
     |                                |                                |
     |                                |                    Developer's |
     |                                |                    FuncUserRegister
     |                                |                    reads       |
     |                                |                    options.ExtraFields.Phone
     |                                |                    options.ExtraFields.Country
     |                                |                    stores in DB
```

---

## Files to Create

| File | Purpose |
|------|---------|
| `types/registration_extra_fields.go` | `RegistrationExtraFields` struct |
| `types/registration_extra_fields_test.go` | Struct tests |
| `types/registration_fields_config.go` | `RegistrationFieldsConfig` struct |
| `types/registration_fields_config_test.go` | Config tests |
| `internal/ui/page_register/countries.go` | Static country list for dropdown |
| `internal/ui/page_register/countries_test.go` | Country list tests |
| `internal/ui/page_register/timezones.go` | Static timezone list for dropdown |
| `internal/ui/page_register/timezones_test.go` | Timezone list tests |

## Files to Modify

| File | Change |
|------|--------|
| `types/user_auth_options.go` | Add `ExtraFields *RegistrationExtraFields` |
| `types/user_auth_options_test.go` | Test new field |
| `types/auth_interfaces.go` | Add `GetRegistrationFields()` / `SetRegistrationFields()` |
| `types/config_username_and_password.go` | Add `RegistrationFields *RegistrationFieldsConfig` |
| `types/config_passwordless.go` | Add `RegistrationFields *RegistrationFieldsConfig` |
| `auth_implementation.go` | Add `registrationFields` field + getter/setter |
| `new_username_and_password_auth.go` | Wire `config.RegistrationFields` |
| `new_passwordless_auth.go` | Wire `config.RegistrationFields` |
| `internal/ui/page_register/page_register.go` | Pass `fieldsConfig` to content builders |
| `internal/ui/page_register/content.go` | Accept `fieldsConfig`, render extra fields, update JS |
| `internal/api/api_register/api_register.go` | Extract extra fields, populate `options.ExtraFields` |
| `internal/api/api_register/dependencies.go` | No signature change needed (options already flows through) |
| `auth_implementation_pages.go` | No change needed (delegates to `page_register.PageRegister`) |
| `auth_implementation_pages_test.go` | Add tests for extra fields rendering |
| `internal/ui/page_register/page_register_test.go` | Add tests if exists, or create |

---

## Usage Example

### Username/Password with Extra Fields

```go
authInstance, err := auth.NewUsernameAndPasswordAuth(types.ConfigUsernameAndPassword{
    Endpoint:             "/auth",
    UrlRedirectOnSuccess: "/dashboard",
    UseCookies:           true,
    EnableRegistration:   true,

    // Enable optional registration fields
    RegistrationFields: &types.RegistrationFieldsConfig{
        EnableBusinessName: true,
        EnablePhone:        true,
        EnableCountry:      true,
        EnableTimezone:     true,
        EnableBirthday:     true,
    },

    FuncUserRegister: func(ctx context.Context, username, password, firstName, lastName string, options types.UserAuthOptions) error {
        // Standard fields
        fmt.Println("email:", username)
        fmt.Println("password:", password)
        fmt.Println("first_name:", firstName)
        fmt.Println("last_name:", lastName)

        // Extra fields (nil if no extra fields were enabled or submitted)
        if options.ExtraFields != nil {
            fmt.Println("business_name:", options.ExtraFields.BusinessName)
            fmt.Println("phone:", options.ExtraFields.Phone)
            fmt.Println("country:", options.ExtraFields.Country)
            fmt.Println("timezone:", options.ExtraFields.Timezone)
            fmt.Println("birthday:", options.ExtraFields.Birthday)
        }

        // Your DB insert logic here
        return createUserInDB(username, password, firstName, lastName, options.ExtraFields)
    },

    // ... other required callbacks ...
})
```

### Passwordless with Extra Fields

```go
authInstance, err := auth.NewPasswordlessAuth(types.ConfigPasswordless{
    Endpoint:             "/auth",
    UrlRedirectOnSuccess: "/dashboard",
    UseCookies:           true,
    EnableRegistration:   true,

    // Enable only phone and country
    RegistrationFields: &types.RegistrationFieldsConfig{
        EnablePhone:   true,
        EnableCountry: true,
    },

    FuncUserRegister: func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) error {
        // Extra fields
        if options.ExtraFields != nil {
            fmt.Println("phone:", options.ExtraFields.Phone)
            fmt.Println("country:", options.ExtraFields.Country)
        }

        return createUserInDB(email, firstName, lastName, options.ExtraFields)
    },

    // ... other required callbacks ...
})
```

### Backward Compatible (No Extra Fields)

```go
// RegistrationFields is nil — current behavior, no extra fields on form
authInstance, err := auth.NewUsernameAndPasswordAuth(types.ConfigUsernameAndPassword{
    Endpoint:             "/auth",
    UrlRedirectOnSuccess: "/dashboard",
    UseCookies:           true,
    EnableRegistration:   true,
    // RegistrationFields not set — defaults to nil
    FuncUserRegister: func(ctx context.Context, username, password, firstName, lastName string, options types.UserAuthOptions) error {
        // options.ExtraFields is nil — existing behavior unchanged
        return createUserInDB(username, password, firstName, lastName)
    },
})
```

---

## Design Decisions

### Why `ExtraFields` on `UserAuthOptions` Instead of New Callback Parameters?

**Option A (chosen):** Add `ExtraFields *RegistrationExtraFields` to `UserAuthOptions`

**Option B (rejected):** Change `FuncUserRegister` signature to add more positional parameters

Reasons for choosing A:
- **Backward compatible** — existing callbacks that don't read `ExtraFields` continue to work without changes
- **Extensible** — adding more optional fields in the future doesn't change the callback signature again
- **Consistent** — `UserAuthOptions` already carries contextual data (`UserIp`, `UserAgent`); extra registration fields are contextual data
- **Clean** — avoids 10-parameter callback signatures like `FuncUserRegister(ctx, email, password, firstName, lastName, businessName, phone, country, timezone, birthday, options)`

### Why a Pointer (`*RegistrationExtraFields`) Instead of a Value?

A nil pointer means "no extra fields were enabled or submitted." This is distinct from a zero-value struct (which would mean "extra fields were enabled but all were empty"). The distinction matters:
- `nil` → developer didn't enable any extra fields → don't bother checking
- `&RegistrationExtraFields{}` → developer enabled fields but user left them all blank → store empty strings if you want

### Why Static Country/Timezone Lists Instead of a Library?

- No external dependency needed
- Countries and timezones change rarely
- The lists are defined as Go slice constants — easy to update, easy to test
- Developers can always override via their own layout function if they need a custom UI

### Why HTML5 `<input type="date">` for Birthday?

- Native browser support, no JavaScript date picker library needed
- Consistent format (`YYYY-MM-DD`) across browsers
- Mobile-friendly (shows native date picker on iOS/Android)
- The library stores and passes the raw string — the developer's callback can parse it as needed

### Why Not Make Extra Fields Required?

The proposal keeps all extra fields optional on the form. If a developer wants to make a field required, they can:
1. Use the library's form as a starting point and override via their own layout
2. Add server-side validation in their `FuncUserRegister` callback (return an error if the field is empty)

Making them required in the library would reduce flexibility — different projects have different requirements.

---

## Testing Strategy

### Unit Tests

- **`types/registration_extra_fields_test.go`** — struct instantiation, zero values
- **`types/registration_fields_config_test.go`** — config defaults, all-true, all-false
- **`types/user_auth_options_test.go`** — ExtraFields nil vs populated
- **`internal/ui/page_register/countries_test.go`** — country list not empty, no duplicates, contains common countries
- **`internal/ui/page_register/timezones_test.go`** — timezone list not empty, no duplicates, contains UTC
- **`internal/ui/page_register/content_test.go`** (new or updated) —
  - Passwordless content with no extra fields (current behavior)
  - Passwordless content with all extra fields enabled
  - Username/password content with no extra fields (current behavior)
  - Username/password content with all extra fields enabled
  - Username/password content with partial fields (only phone + country)
  - Verify each enabled field renders the correct `<input>` / `<select>` with correct `name` attribute
  - Verify disabled fields do NOT appear in HTML
- **`internal/api/api_register/api_authknight_callback_test.go`** —
  - Register with extra fields in POST body → `options.ExtraFields` populated
  - Register without extra fields → `options.ExtraFields` is nil
  - Register with some extra fields (only phone) → only `ExtraFields.Phone` set
- **`auth_implementation_pages_test.go`** —
  - GET register page with `RegistrationFields` config → HTML contains enabled field inputs
  - GET register page without `RegistrationFields` config → HTML does not contain extra field inputs

### Test Coverage Target

Maintain the existing ~90% coverage standard. All new files must have corresponding test files.

---

## Migration Path

This is a purely additive, backward-compatible change:

- `RegistrationFields` defaults to `nil` on both configs → no extra fields rendered (current behavior)
- `UserAuthOptions.ExtraFields` defaults to `nil` → existing callbacks unaffected
- No changes to any callback signatures
- No changes to existing API request/response format (extra fields are just additional optional POST parameters)
- No changes to route constants or paths

Existing users of the library are completely unaffected.

---

## Implementation Order

1. `types/registration_extra_fields.go` + tests
2. `types/registration_fields_config.go` + tests
3. `types/user_auth_options.go` — add `ExtraFields` + tests
4. `types/auth_interfaces.go` — add `GetRegistrationFields()` / `SetRegistrationFields()`
5. `types/config_username_and_password.go` — add `RegistrationFields` field
6. `types/config_passwordless.go` — add `RegistrationFields` field
7. `auth_implementation.go` — add `registrationFields` field + getter/setter
8. `new_username_and_password_auth.go` — wire config
9. `new_passwordless_auth.go` — wire config
10. `internal/ui/page_register/countries.go` + tests
11. `internal/ui/page_register/timezones.go` + tests
12. `internal/ui/page_register/content.go` — accept `fieldsConfig`, render extra fields, update JS
13. `internal/ui/page_register/page_register.go` — pass `fieldsConfig` from auth instance
14. `internal/api/api_register/api_register.go` — extract extra fields, populate `options.ExtraFields`
15. `auth_implementation_pages_test.go` — add tests for extra fields rendering
16. Run full test suite: `go test -v ./...`
17. Check coverage: `go test -cover ./...`

---

## Open Questions

1. **Should the country dropdown use ISO codes or country names?** Proposal: use country names as the display value and ISO 3166-1 alpha-2 codes as the `<option value="">`. This is the most common pattern. The developer receives the ISO code in `ExtraFields.Country`.

2. **Should the timezone list be exhaustive (all IANA timezones) or a curated subset?** Proposal: curated subset of ~50 common timezones covering all major UTC offsets. Developers can override via layout if they need the full list.

3. **Should we add client-side validation for phone format?** Proposal: no — phone formats vary wildly internationally. Basic "non-empty if filled" check only. Server-side validation is the developer's responsibility.

4. **Should extra fields be stored in the verification token payload (when `EnableVerification` is true)?** Currently, the verification flow stores `email`, `first_name`, `last_name`, `password` in a temporary key. Proposal: yes, include extra fields in the JSON payload so they survive the verification round-trip. This requires updating `internal/core/register_with_username_and_password.go` to marshal extra fields into the temporary key payload and retrieve them on verification.

5. **Should we support custom extra fields (developer-defined, not just the 5 built-in)?** Proposal: not in this iteration. The 5 fields cover the common case. Custom fields would require a more generic approach (map[string]string) which adds complexity. Can be revisited if needed.
