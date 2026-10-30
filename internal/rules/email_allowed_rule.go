package rules

import (
	"github.com/dracory/auth/types"
	"github.com/dracory/rule"
)

// EmailAllowedRule enforces an optional host-provided email allowlist —
// the check magic-link send performs before any token is created or email
// sent. A nil allowlist function means no restriction: the rule passes.
type EmailAllowedRule struct {
	rule.Rule
}

// EmailAllowedContext carries the inputs to EmailAllowedRule.
type EmailAllowedContext struct {
	Allowed func(email string) bool
	Email   string
}

// NewEmailAllowedRule creates an EmailAllowedRule for the given allowlist
// function and email. Fails with MsgInvalidCredentials so the response
// cannot confirm whether the address exists on the allowlist.
func NewEmailAllowedRule(allowed func(email string) bool, email string) *EmailAllowedRule {
	r := &EmailAllowedRule{}
	r.Rule.SetContext(EmailAllowedContext{Allowed: allowed, Email: email})
	r.Rule.SetCondition(func(ctx any) bool {
		c, _ := ctx.(EmailAllowedContext)
		if c.Allowed == nil {
			return true // no allowlist configured: every email passes
		}
		if !c.Allowed(c.Email) {
			r.AddFailMessage(types.MsgInvalidCredentials)
			return false
		}
		return true
	})
	return r
}
