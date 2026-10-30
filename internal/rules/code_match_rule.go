package rules

import (
	"crypto/subtle"

	"github.com/dracory/auth/types"
	"github.com/dracory/rule"
)

// CodeMatchRule compares a stored code payload against the expected value
// in constant time — the compare the code-verify flows perform after the
// attempt-limit check. An empty stored value always fails, and the compare
// never leaks which bytes matched.
type CodeMatchRule struct {
	rule.Rule
}

// CodeMatchContext carries the inputs to CodeMatchRule.
type CodeMatchContext struct {
	Stored   string
	Expected string
}

// NewCodeMatchRule creates a CodeMatchRule for the given stored and
// expected values. Fails with MsgVerificationCodeExpired so the response
// cannot distinguish "no code" from "wrong code".
func NewCodeMatchRule(stored string, expected string) *CodeMatchRule {
	r := &CodeMatchRule{}
	r.Rule.SetContext(CodeMatchContext{Stored: stored, Expected: expected})
	r.Rule.SetCondition(func(ctx any) bool {
		c, _ := ctx.(CodeMatchContext)
		if c.Stored == "" || subtle.ConstantTimeCompare([]byte(c.Stored), []byte(c.Expected)) != 1 {
			r.AddFailMessage(types.MsgVerificationCodeExpired)
			return false
		}
		return true
	})
	return r
}
