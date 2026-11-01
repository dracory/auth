package rules

import (
	"github.com/dracory/auth/types"
	"github.com/dracory/rule"
)

// VerifiedEmailKeyRule resolves a verified-email token (verified_key)
// to the email it certifies. It passes when the key is non-empty and
// resolves to a non-empty email in the temporary key store, and fails —
// with the expired-link message — otherwise. The resolved email is
// carried on the rule, same convention as ProfileCompletionRule.UserID.
//
// Shared by the register page (render read-only email) and the register
// API (token-verified create) so the resolution invariant is identical
// on both sides.
type VerifiedEmailKeyRule struct {
	rule.Rule
	email string
}

// VerifiedEmailKeyContext carries the inputs to VerifiedEmailKeyRule.
type VerifiedEmailKeyContext struct {
	// TemporaryKeyGet resolves the token to the verified email.
	TemporaryKeyGet func(key string) (string, error)
	// Key is the verified_key value presented by the client.
	Key string
}

// NewVerifiedEmailKeyRule creates a VerifiedEmailKeyRule for the given
// store lookup and token.
func NewVerifiedEmailKeyRule(tempKeyGet func(key string) (string, error), key string) *VerifiedEmailKeyRule {
	r := &VerifiedEmailKeyRule{}
	r.Rule.SetContext(VerifiedEmailKeyContext{
		TemporaryKeyGet: tempKeyGet,
		Key:             key,
	})
	r.Rule.SetCondition(func(ctx any) bool {
		c, _ := ctx.(VerifiedEmailKeyContext)
		if c.Key == "" || c.TemporaryKeyGet == nil {
			return false
		}
		email, err := c.TemporaryKeyGet(c.Key)
		if err != nil || email == "" {
			r.AddFailMessage(types.MsgLinkNotValidOrExpired)
			return false
		}
		r.email = email
		return true
	})
	return r
}

// Email returns the verified email resolved during the last evaluation,
// or "" when the rule has not been evaluated or failed.
func (r *VerifiedEmailKeyRule) Email() string {
	return r.email
}
