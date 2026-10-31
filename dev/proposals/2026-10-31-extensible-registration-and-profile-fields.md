# Proposal: Extensible Registration & Profile-Completion Fields

**Date:** 2026-10-31
**Status:** Draft — pending review
**Scope:** `types/`, `auth_implementation.go`, `internal/ui/*`, `internal/api/api_register`, `internal/api/api_register_code_verify`, `internal/core`, `internal/rules`, `multi_method.go`, new `page_profile_completion` + `api_profile_completion`

## Motivation

Two gaps force host projects to shadow `/auth/register` with their own
controllers (dogpace.com-golang does exactly this — `internal/controllers/auth/register_password`
and `internal/controllers/auth/register`):

1. **Fixed field set.** The built-in register forms only collect
   `first_name`, `last_name`, `email`, `password`. Hosts commonly need
   country, timezone, phone, business name, birthday, etc. Peers solve
   this with declared field schemas (better-auth `additionalFields`,
   SuperTokens form fields, Keycloak declarative user profile). Our
   callback-driven storage model fits the same approach — the host
   already persists whatever it wants inside `FuncUserRegister`; it just
   never receives the values.

2. **Two different flows are conflated.** A user who registers via
   password (`/auth/register` → code verify → create) is *creating* an
   account. A user who arrives via OTP/magic-link is *already
   authenticated with a verified email* — their user record was created
   by `UserFindByEmailOrCreate` during the post-auth pipeline. For them
   "registration" is really **profile completion**: they must NOT be
   able to edit the email (it was just proven theirs), and the operation
   must update the existing user, not create a second one. The package
   currently has no completion flow, so hosts build one — and collide
   with the package's own `/auth/register` mount.

3. **No registration-time allowlist hook.** `PostAuthHooks.EmailAllowed`
   runs only at login; nothing stops a disallowed email from hitting
   `POST /auth/api/register`. Hosts enforcing an allowlist must shadow
   the endpoint.

## Design Principles

- **Non-breaking.** All existing `FuncUserRegister` signatures keep
  working untouched; field values travel inside `UserAuthOptions`.
- **Host-declared, package-rendered.** Hosts describe fields in config;
  the built-in pages render Bootstrap markup matching the existing
  `FieldGroup`/`NameRow` style.
- **Server-side truth.** Client-side validation is UX only; every field
  is re-validated (required + custom `Validate`) in the API before the
  values reach host code or the temp-key payload.
- **Minimal v1.** Static select options only — no dynamic option
  providers, no conditional fields, no JSON-schema engine.

## Proposed Changes

### 1. Field schema — new `types/fields.go`

```go
package types

// FieldType enumerates the HTML input kinds the built-in pages can render.
type FieldType string

const (
    FieldTypeText     FieldType = "text"
    FieldTypeEmail    FieldType = "email"
    FieldTypeTel      FieldType = "tel"
    FieldTypeDate     FieldType = "date"
    FieldTypeNumber   FieldType = "number"
    FieldTypeSelect   FieldType = "select"
    FieldTypeCheckbox FieldType = "checkbox"
)

// FieldOption is one <option> in a FieldTypeSelect field.
type FieldOption struct {
    Value string
    Label string
}

// RegistrationField declares one extra input rendered on the built-in
// register / profile-completion pages and validated server-side.
type RegistrationField struct {
    Name         string        // form field name, e.g. "country"
    Label        string        // rendered above the input, e.g. "Country"
    Type         FieldType     // empty => FieldTypeText
    Required     bool
    Placeholder  string
    Options      []FieldOption // FieldTypeSelect only
    Autocomplete string        // e.g. "tel", "bday", "country-name"
    // Validate optionally inspects the submitted value and returns an
    // error message; "" means valid.
    Validate func(value string) string
}
```

`ConfigShared` gains:

```go
// RegistrationFields are extra inputs rendered on the public sign-up
// form (all methods) after the built-in fields. Collected values are
// delivered to FuncUserRegister via UserAuthOptions.Fields.
RegistrationFields []RegistrationField

// ProfileFields are inputs on the authenticated profile-completion
// page — for users created implicitly (OTP / magic link / AuthKnight)
// whose email is already verified and shown read-only. Values are
// delivered to FuncUserProfileUpdate.
ProfileFields []RegistrationField

// FuncRegistrationAllowed optionally gates registration before any
// code is generated or email sent — e.g. an email allowlist.
FuncRegistrationAllowed func(email string) (allowed bool, message string)

// FuncUserProfileUpdate persists completion-page values for an
// existing authenticated user.
FuncUserProfileUpdate func(ctx context.Context, userID string, fields map[string]string, options UserAuthOptions) error

// FuncUserProfile optionally returns current values to prefill the
// completion form; missing keys render empty inputs.
FuncUserProfile func(ctx context.Context, userID string, options UserAuthOptions) (map[string]string, error)
```

### 2. `UserAuthOptions` carries field values

```go
type UserAuthOptions struct {
    UserIp    string
    UserAgent string
    // Fields holds values for the declared RegistrationFields /
    // ProfileFields. nil when none are configured.
    Fields map[string]string
}
```

Every existing `FuncUserRegister` variant already receives `options` —
no signature changes, fully backward compatible.

### 3. Field validation helper — `internal/rules` or `utils`

```go
// utils/collect_fields.go (shared by api_register, core, api_profile_completion)
func CollectFields(r *http.Request, fields []types.RegistrationField) (map[string]string, *FieldError)
```

- Reads `req.GetStringTrimmed(r, f.Name)` for **declared names only** —
  undeclared form keys are ignored (no injection vector).
- HTML-escapes each value (`html.EscapeString`, matching the existing
  first/last-name handling).
- `Required` → non-empty; `Validate` → must return `""`; `select` →
  value must be in `Options` (prevents forged values).
- Returns first failure as a user-facing message.

### 4. Page rendering — `internal/ui/shared/components.go` + `page_register`

New component:

```go
// shared/components.go
func FieldInput(f types.RegistrationField, value string) hb.TagInterface
```

- maps `FieldType` → `hb.NewInput().Type(...)` or `hb.NewSelect()` with
  `Options`; applies `required`, `placeholder`, `autocomplete`,
  `form-control form-control-lg`; pre-selects `value` for prefill.

`page_register` — all three content builders gain a `fields` parameter
and render `shared.FieldInput(f, "")` after `NameRow`/email and before
`PrimaryButton`:

- `RegisterPasswordlessContent(urlLogin, fields)`
- `RegisterUsernameAndPasswordContent(urlLogin, urlPasswordRestore, fields)`
- `RegisterAuthKnightContent(urlLogin, verifiedEmail, akKey, fields)`

`PageRegister` passes `a.GetRegistrationFields()` (optional-interface
resolution, see §8).

### 5. JS serialization — `register_*.js`

Each register page embeds the declared field names via a template
placeholder injected next to the existing `__URL_API_REGISTER__` ones:

```js
var extraFieldNames = __EXTRA_FIELD_NAMES__; // e.g. ["country","phone"]
```

In `registerFormValidate()`: after built-in validation, loop
`extraFieldNames`, `inputVal(name)` (checkbox → checked flag as
`"true"`/`""`), client-side required check, then merge into the
`postForm` body. Names arrive JSON-encoded so markup stays safe.

### 6. API + payload flow

**`api_register.go` (`RegisterPasswordlessInit` + password branch):**
collect & validate fields via `CollectFields`; add `"fields":
map[string]string` to the `registercode:` JSON payload (passwordless)
or pass straight into `options.Fields` (password, no-verification
branch via core). Invoke `FuncRegistrationAllowed(email)` before code
generation — failure returns `api.Error(message)`.

**`core/register_with_username_and_password.go`:** signature unchanged;
fields arrive inside `options.Fields` (the caller
`deps.RegisterWithUsernameAndPassword` in `api_register.go` collects
them). Validation + `FuncRegistrationAllowed` run there too; `"fields"`
added to the verify-enabled JSON payload (line ~112).

**`api_register_code_verify.go`:** extract `registerMap["fields"]`
(`map[string]any` → `map[string]string`) alongside the existing
email/first_name/last_name/password extraction; the deps wrappers
(`PasswordlessUserRegister`, `UserRegister`) gain a `fields` parameter,
and `ApiRegisterCodeVerifyWithAuth` sets `options.Fields = fields` when
invoking the host funcs.

**AuthKnight register (`api_register.go` ak branch):** collect +
validate fields, pass via `options.Fields` into `UserRegister`.

Backward compatibility: old payloads without `"fields"` unmarshal to
`nil` map — hosts see `options.Fields == nil`.

### 7. Profile-completion flow (the OTP case)

New files:

- `internal/ui/page_profile_completion/page_profile_completion.go` +
  `content.go` + `profile_completion.js`
- `internal/api/api_profile_completion/api_profile_completion.go`

**Route mounts in `multi_method.go`** (only when `len(ProfileFields) > 0`
and `FuncUserProfileUpdate != nil`):

```
GET  {endpoint}/profile-completion
POST {endpoint}/api/profile-completion
```

**Auth resolution (both handlers):** read the auth cookie (same name the
package sets), call `FuncUserFindByAuthToken` → `userID`. Unauthenticated:
page redirects to `links.Login(endpoint)`; API returns
`api.Unauthenticated`. This mirrors how `FuncUserFindByAuthToken` already
backs the web/API middlewares.

**Page:** email rendered read-only exactly like
`RegisterAuthKnightContent` (`form-control-plaintext` + hidden input —
actually just display text, no hidden input needed since the email is
resolved server-side from the session, never trusted from the form).
Then `ProfileFields` rendered via `shared.FieldInput(f,
prefill[f.Name])` using `FuncUserProfile` when set. Submit button
"Save & continue".

**API:** `CollectFields(r, profileFields)` → `FuncUserProfileUpdate(ctx,
userID, fields, options)` → success JSON `{redirect: UrlRedirectOnSuccess}`.

**Trigger:** intentionally NOT automatic. The host's `PostAuthHooks.RedirectURL`
decides when a user needs completion (e.g. dogpace's
`IsRegistrationCompleted`). The package ships the page + API; hosts
point redirects at `{endpoint}/profile-completion`. (Future iteration
could add a `FuncProfileComplete(userID) bool` so the package itself
gates, but that needs a way to know "complete" — host-specific.)

### 8. Implementation storage + optional getters/setters

`authImplementation` gains fields:

```go
registrationFields    []types.RegistrationField
profileFields         []types.RegistrationField
funcRegistrationAllowed func(email string) (bool, string)
funcUserProfileUpdate func(ctx context.Context, userID string, fields map[string]string, options types.UserAuthOptions) error
funcUserProfile       func(ctx context.Context, userID string, options types.UserAuthOptions) (map[string]string, error)
```

Each gets `GetX`/`SetX` on `auth_implementation.go` +
`auth_implementation_setters.go` (the setter-file pattern already in
use). `ConfigShared` values are copied in `new_passwordless_auth.go`,
`new_username_and_password_auth.go`, `new_authknight_auth.go`.

`AuthSharedInterface` is not extended (would break external
implementers — the documented reason `FuncEmailEnqueue`/`PostAuthHooks`
use optional interfaces). Consumers resolve via type assertion after
`types.UnwrapAuth(a)`:

```go
type registrationFieldsProvider interface {
    GetRegistrationFields() []types.RegistrationField
}
```

### 9. Dogpace-side follow-up (documented, not part of this package change)

Once this lands, `dogpace.com-golang`:

- deletes `internal/controllers/auth/register_password` entirely
- replaces `internal/controllers/auth/register` with `ProfileFields` +
  `FuncUserProfileUpdate` (vault-tokenizing) config
- wires `FuncRegistrationAllowed` to `authrules.NewEmailAllowedRule`
- fixes `FuncUserRegister` (passwordless) to persist first/last name —
  currently ignored, producing incomplete users who bounce back to
  register
- timezone-by-country AJAX stays a small host-side `Validate`/JS
  concern (options are static in v1; timezone list can be static too —
  ~400 entries — or deferred)

## Explicitly Out of Scope

- Dynamic option providers (`OptionsFunc`), cascading selects, conditional fields.
- JSON-Schema identity traits (Kratos-style).
- Profile *editing* after completion (settings page) — same machinery, later.
- Field ordering beyond declaration order; field grouping/rows.
- Translatable labels (labels are host-supplied strings — already the host's problem).

## Test Matrix

| Area | Cases |
|---|---|
| `CollectFields` | required missing, validate error, select value not in options, undeclared key ignored, HTML escaping, checkbox true/empty |
| `shared.FieldInput` | each FieldType renders; required/autocomplete/placeholder attrs; select options + preselected value |
| `page_register` | fields appear in all 3 forms; AuthKnight keeps read-only email |
| `api_register` | fields validated before code gen; `FuncRegistrationAllowed` blocks; `fields` in temp-key payload |
| `register_code_verify` | fields round-trip: stored at init → delivered in `options.Fields`; old payloads → nil Fields |
| `core` | fields included in verify payload; allowed-check invoked |
| `page_profile_completion` | unauthenticated → login redirect; authenticated → email read-only + prefilled fields |
| `api_profile_completion` | unauthenticated → `api.Unauthenticated`; valid → `FuncUserProfileUpdate` called with userID + fields; invalid → error message |
| `multi_method` | completion routes mounted only when configured; no mounts → 404 |

## Suggested Order

1. `types/fields.go` + `UserAuthOptions.Fields` + `ConfigShared` fields.
2. `authImplementation` storage, getters/setters, config wiring (×3 constructors).
3. `utils.CollectFields` + tests.
4. `shared.FieldInput` + `page_register` rendering + JS (`__EXTRA_FIELD_NAMES__`).
5. `api_register` + `core` collection/validation/payload + allowlist hook.
6. `api_register_code_verify` field extraction → `options.Fields`.
7. `page_profile_completion` + `api_profile_completion` + `multi_method` mounts.
8. Docs (register page section, README example, `docs/authknight.html` table row).

## Risks / Open Questions

- **Escaping direction:** values are escaped at collection (`html.EscapeString`)
  then stored in JSON — prefill must not double-escape (hb escapes on render;
  store the *raw* value? Decide: store raw, escape only at render — check the
  existing first-name path, which escapes at collection time, and match it for
  consistency).
- **Checkbox semantics:** `"true"`/`""` string encoding documented for hosts.
- **Select with 400+ options** (timezones): acceptable statically; flag that a
  `SelectMultiple` or AJAX provider is future work if UX suffers.
- **`FuncUserProfileUpdate` vs reusing `FuncUserRegister`:** a separate hook
  keeps "create" vs "update" unambiguous — the OTP-completion path must never
  create a user.
