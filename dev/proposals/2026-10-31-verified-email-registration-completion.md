# Proposal: Authenticated Register = Profile Completion (verified email, read-only)

**Date:** 2026-10-31
**Status:** Draft — pending review
**Scope:** `internal/ui/page_register`, `internal/api/api_register`, `auth_implementation.go`, `types/`. Deliberately smaller than `2026-10-31-extensible-registration-and-profile-fields.md` — no custom field system.

## Motivation

OTP and magic-link login can create the user record *during login*
(`UserFindByEmailOrCreate`). That user then hits `GET /auth/register`
**already authenticated with an email that was just verified by the
code** — exactly like the AuthKnight flow, where `ak_key` already proves
the email and the form renders it read-only.

Today the built-in register form always shows an editable email and
always *creates* a user — so a verified user can change their email
(wrong) or submit a duplicate account (wrong). Hosts currently fix this
by shadowing `/auth/register` with custom controllers.

**Rule:** if the requester is already authenticated, registration is
*completion* — email comes from the verified session, is shown
read-only, and the operation updates the existing user instead of
creating one. AuthKnight's `ak_key` branch is the existing precedent;
this proposal generalises it to any authenticated session.

## Proposed Changes

### 1. Page: `page_register.PageRegister` — authenticated branch

Resolution order in `PageRegister` (before the current branches):

1. `ak_key` present → existing AuthKnight verified-email form (unchanged).
2. `a.GetCurrentUserID(r) != ""` → **completion form** (new):
   - email rendered as read-only plaintext, same markup as
     `RegisterAuthKnightContent` (`form-control-plaintext` + label
     "Email (verified)"). Email text comes from the new optional hook
     `FuncUserEmailGet` (§3); when the hook is unset, show
     "Your verified email" without the address — the email is never
     needed from the form anyway, it is resolved server-side.
   - first/last name inputs + primary button "Save & continue".
   - no password fields, no hidden `email` input (never trust it).
3. Otherwise → existing public forms (unchanged).

New content builder:

```go
func RegisterCompletionContent(urlLogin, email string) string
```

Largely a call into the same markup `RegisterAuthKnightContent` uses
(minus `ak_key` hidden input). New `register_completion.js` — posts only
`first_name`/`last_name` to `__URL_API_REGISTER__` then redirects to
`__URL_ON_SUCCESS__`.

### 2. API: `api_register.ApiRegister` — authenticated branch

Early in `ApiRegister` (before the passwordless/password dispatch):

```go
if userID := deps.CurrentUserID(r); userID != "" {
    // Authenticated: this is profile completion, not account creation.
    firstName := html.EscapeString(req.GetStringTrimmed(r, "first_name"))
    lastName  := html.EscapeString(req.GetStringTrimmed(r, "last_name"))
    // required-field rules (MsgFirstNameRequired / MsgLastNameRequired)
    // then:
    err := deps.UserProfileUpdate(ctx, userID, firstName, lastName, options)
    // → api.Success("profile saved") with {redirect: redirectOnSuccess}
}
```

- **No** code generation, temp keys, email sends, password handling.
- **No** `FuncUserRegister` call — the user already exists; a dedicated
  update hook keeps "create" vs "update" unambiguous.
- `Dependencies` gains `CurrentUserID func(r *http.Request) string` and
  `UserProfileUpdate func(ctx, userID, firstName, lastName string,
  options UserAuthOptions) error`; `ApiRegisterWithAuth` wires them from
  `a.GetCurrentUserID` and the new `FuncUserProfileUpdate` setter.
- When `FuncUserProfileUpdate` is nil → `api.Error(MsgFailedToProcess)`
  (host opted into auth-impl sessions but no profile update support —
  same failure style as missing register fn).

`AuthSharedInterface` is **not** extended — `FuncUserProfileUpdate` and
`FuncUserEmailGet` live on the implementation + optional-interface
assertion, same pattern as `PostAuthHooks`/`FuncEmailEnqueue`.

### 3. New optional hooks on `authImplementation`

```go
// FuncUserEmailGet resolves a user's email for display on the
// completion form. Optional — form works without it.
funcUserEmailGet func(ctx context.Context, userID string, options types.UserAuthOptions) (string, error)

// FuncUserProfileUpdate persists the completion form for an existing
// authenticated user.
funcUserProfileUpdate func(ctx context.Context, userID string, firstName string, lastName string, options types.UserAuthOptions) error
```

Get/Set accessors in `auth_implementation.go` / `_setters.go`; both
optional — no `ConfigShared` change required (setters allow host
injection post-construction, matching `SetPostAuthHooks`).

### 4. AuthKnight stays as-is

The `ak_key` path already implements "verified email → read-only →
register-with-names". After this change the semantics are uniform:

| Entry | Email source | Form email | Server action |
|---|---|---|---|
| Anonymous, password method | typed input | editable | create user (+code verify) |
| Anonymous, passwordless | typed input | editable | code → create on verify |
| `ak_key` (AuthKnight) | temp key | read-only | create via `UserRegister` |
| **Authenticated (OTP/magic-link)** | session | read-only | **`FuncUserProfileUpdate`** |

## Explicitly Out of Scope

- Custom/extra fields (`RegistrationField` schema) — the full proposal
  `2026-10-31-extensible-registration-and-profile-fields.md` builds on
  this later; the completion form here handles first/last name only.
- Registration-time allowlist hook (`FuncRegistrationAllowed`).
- Auto-detection of "profile incomplete" inside the package — the
  host's `RedirectURL` hook still decides who lands on `/auth/register`.
- Dedicated `/profile-completion` route — reusing `/auth/register`
  means one URL serves both intents, gated on authentication state.

## Host-side usage (dogpace follow-up)

```go
impl.SetFuncUserProfileUpdate(func(ctx, userID, first, last string, o types.UserAuthOptions) error {
    user, _ := userStore.UserFindByID(ctx, userID)
    // vault-tokenize names, persist
    return userStore.UserUpdate(ctx, user)
})
impl.SetFuncUserEmailGet(...) // optional, vault-aware email lookup
```

Then `internal/controllers/auth/register` and `register_password`
both get deleted; only `links.Auth().Register()` redirects remain.

## Test Matrix

| Area | Cases |
|---|---|
| `PageRegister` | unauthenticated → public form; authenticated → email read-only, no password inputs; `FuncUserEmailGet` nil → generic label; `ak_key` still wins |
| `ApiRegister` | authenticated → calls `UserProfileUpdate` with session userID (form `email` field ignored/forged → still uses session); missing fn → error; required names enforced; unauthenticated → unchanged flows |
| hooks | getter/setter round-trip |

## Suggested Order

1. `authImplementation` hooks + optional getters/setters.
2. `RegisterCompletionContent` + `register_completion.js` + `PageRegister` branch.
3. `api_register` authenticated branch + `ApiRegisterWithAuth` wiring.
4. Tests.
