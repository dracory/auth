package types

import (
	"context"
	"net/http"
)

// PostAuthHooks is the optional rule/pipeline hook set every login method's
// verify path funnels through: allowlist → find-or-create user → active
// check → session create → auth cookie → safe redirect. All fields are
// optional; nil hooks apply sane defaults (allow all emails, no active
// check, no custom redirect, no extra rules).
type PostAuthHooks struct {
	// EmailAllowed gates the email; deny message returned when false.
	EmailAllowed func(ctx context.Context, email string) (allowed bool, denyMessage string)

	// UserFindByEmailOrCreate resolves email→userID, creating the user if
	// needed. Vault/blind-index apps implement their lookup here.
	UserFindByEmailOrCreate func(ctx context.Context, email string, options UserAuthOptions) (userID string, err error)

	// UserIsActive rejects inactive/banned accounts.
	UserIsActive func(ctx context.Context, userID string, options UserAuthOptions) bool

	// SessionCreate issues the post-auth session; supersedes
	// FuncSessionCreate for pipeline consumers.
	SessionCreate func(ctx context.Context, userID string, options UserAuthOptions) (SessionInterface, error)

	// SetAuthCookie writes the auth cookie; falls back to the module's
	// default cookie writer when nil.
	SetAuthCookie func(w http.ResponseWriter, r *http.Request, token string)

	// RedirectURL computes the post-login destination (role-aware).
	RedirectURL func(ctx context.Context, userID string) string

	// RegisterRules returns arbitrary rules that must all pass for the
	// flow to continue (covers project-specific rules without a closed
	// enum). First failure short-circuits.
	RegisterRules func(ctx context.Context, email string) []PostAuthRule

	// IssueRememberSession creates a marked long-lived session
	// (utils.RememberSessionMarker) for remember-me. Called only when
	// remember-me is enabled AND the user opted in; the pipeline sets the
	// remember cookie with the returned session key.
	IssueRememberSession func(ctx context.Context, userID string, options UserAuthOptions) (SessionInterface, error)
}

// PostAuthRule is the minimal rule contract for pipeline rules — kept
// independent so dracory/rule implementations adapt trivially.
type PostAuthRule interface {
	Passes() bool
	// FailMessage returns the message shown when the rule fails.
	FailMessage() string
}
