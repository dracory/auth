package rules

import (
	"net/http"

	"github.com/dracory/rule"
)

// ProfileCompletionRule decides whether a registration request is
// actually a profile completion: the requester already holds a verified
// session (e.g. created implicitly by an OTP or magic-link login), so
// the email is proven and the existing user must be updated — never
// created. It passes when an authenticated user ID resolves from the
// request and fails otherwise (public sign-up path).
//
// Shared by the register page (render read-only email form) and the
// register API (route to UserProfileUpdate) so the invariant is named,
// tested, and identical on both sides.
type ProfileCompletionRule struct {
	rule.Rule
	userID string
}

// ProfileCompletionContext carries the inputs to ProfileCompletionRule.
type ProfileCompletionContext struct {
	// ResolveUserID returns the authenticated user ID for the request,
	// or "" when unauthenticated.
	ResolveUserID func(r *http.Request) string
	// Request is the HTTP request being evaluated.
	Request *http.Request
}

// NewProfileCompletionRule creates a ProfileCompletionRule for the given
// session resolver and request.
func NewProfileCompletionRule(resolveUserID func(r *http.Request) string, r *http.Request) *ProfileCompletionRule {
	rule := &ProfileCompletionRule{}
	rule.Rule.SetContext(ProfileCompletionContext{
		ResolveUserID: resolveUserID,
		Request:       r,
	})
	rule.Rule.SetCondition(func(ctx any) bool {
		c, _ := ctx.(ProfileCompletionContext)
		if c.ResolveUserID == nil || c.Request == nil {
			return false
		}
		rule.userID = c.ResolveUserID(c.Request)
		return rule.userID != ""
	})
	return rule
}

// UserID returns the user ID resolved during the last evaluation, or ""
// when the rule has not been evaluated or the request is unauthenticated.
func (r *ProfileCompletionRule) UserID() string {
	return r.userID
}
