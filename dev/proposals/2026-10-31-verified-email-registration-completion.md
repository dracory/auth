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

## Follow-up: user-creation policy + email capture

Discussion surfaced a design axis the package should expose rather than
hardcode: **when is the user record created relative to registration?**
Hosts disagree — some want every verified email captured even if the
user abandons registration; others do not want incomplete rows
polluting `users`.

The naive two-policy framing has a downside either way:

| Policy | Downside |
|---|---|
| Session-first (current OTP: verify → create row → collect details) | Row is **verified but incomplete** — ghost users in `users` |
| Register-first (current AuthKnight: collect details → create row) | If the email was captured before verification, the row is **complete but unverified** — a second verification email is needed |

Both downsides disappear with a two-step registration — verify the
email first, then collect details:

```
1. enter email → send code
2. verify code → issue short-lived verified-email token
3. full registration form (email read-only) → submit → single INSERT
```

Step 2's token is the same artifact as AuthKnight's `ak_key`: a
verified-email credential that authorizes the final create. The proof
is not consumed at verify time — it is *exchanged* for the token
carried to the register submit.

- Nothing is written to `users` before the form submits → clean DB,
  no ghost rows.
- No pending data rides inside the proof — names are collected by the
  post-verify form, not pre-verify.
- The single insert is both verified and complete — no
  "verified-but-empty" state, no post-hoc verification email.
- The "registration form" and the authenticated "completion form"
  (§1) become the **same** form — read-only verified email + names —
  differing only in the server action: `UserRegister` (token path,
  create) vs `FuncUserProfileUpdate` (session path, update).

The only thing register-first loses is the email of users who abandon
mid-flow. That is a separate concern — *lead capture*, not account
creation — and belongs in an optional hook:

```go
// FuncEmailCaptured fires when an email is submitted and again when it
// is verified. The verified flag distinguishes the two events;
// capturing unverified emails is a spam/abuse vector, so hosts decide.
funcEmailCaptured func(ctx context.Context, email string, verified bool, options types.UserAuthOptions) error
```

- Hosts wanting leads implement it (leads table, CRM, re-engagement).
- Hosts wanting a clean DB leave it nil — nothing extra is persisted.
- Neither policy forces the other's trade-off.

Resulting matrix:

| Flow | Capture | Insert | Result |
|---|---|---|---|
| Session-first (OTP login) | verified email → user row | details update later | verified-but-incomplete rows |
| Register-first (details → verify → insert) | optional `FuncEmailCaptured` | one verified + complete insert | clean; abandoned emails kept only if host opts in |

Register-first is strictly better for DB hygiene; the capture hook is
strictly better for lead retention.

### Unification goal — one pipeline, three front doors

OTP, magic link, and AuthKnight are the same flow with different proof
delivery, and the package **must unify them** on register-first
ordering:

```
enter email → issue proof → verify → verified-email token → form → single INSERT
```

| Method | Proof | Delivery |
|---|---|---|
| OTP | short code in temp store | email |
| Magic link | signed token in URL | email |
| AuthKnight | `ak_key` temp key | cross-app redirect |

All three are "issue a proof, verify it, carry the verified-email
token to the form, then create". Today's divergence is accidental:
OTP/magic-link shortcut via `UserFindByEmailOrCreate` at verify time
(session → incomplete row), while AuthKnight carries the `ak_key`
token forward to the register submit. Once OTP/magic-link issue a
verified-email token on successful verification — the way `ak_key`
already works — there is **one** verify → form → create pipeline and
the method-specific branches (including the `ak_key` register branch)
collapse into it.

`RegistrationComplete` gating + `FuncUserProfileUpdate` remain as the
session-first variant for hosts that deliberately choose
`CreateUserOnAuth`, but it is the opt-in exception — not three parallel
implementations.

### Password method stays orthogonal

Password registration does not fit the verified-first shape and is a
**host choice**, not a pipeline branch: developers who want passwords
enable the password method (verify-after-create as today, or a future
email+password → verify → names → insert variant); developers who do
not simply leave it disabled. The unified flow is unaffected either
way.

### Caveat: user enumeration

One downside of the unified entry point: differing behaviour for
"email already registered" vs "new email" leaks account existence
(user enumeration). Mitigation — identical external responses
pre-verify:

- Request side: always return "if this address is eligible, we sent
  you a code" — same status, same timing.
- Email side: always send *something* — existing users get a
  login/verify email; the already-registered case may get a "someone
  tried to register" notice instead of failing silently.
- Post-verify divergence is safe: the requester has proved inbox
  ownership, so "existing → log in (+ `RegistrationComplete` gate)" vs
  "new → collect details → insert" reveals nothing to anyone but the
  mailbox owner.

Enumeration resistance is therefore a requirement on the pre-verify
response surface, not a reason to keep the flows separate.

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
