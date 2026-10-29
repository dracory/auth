// Package rules contains the reusable verification rules shared by the
// code-verify flows (login and register). Each rule wraps
// github.com/dracory/rule so the check logic lives separately from the
// HTTP/implementation plumbing.
package rules

import (
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/rule"
	"github.com/dracory/str"
)

// VerificationCodeRule validates the submitted verification code: required,
// correct length for the configured gamma, and only allowed characters.
type VerificationCodeRule struct {
	rule.Rule
}

// NewVerificationCodeRule creates a VerificationCodeRule for the given code.
// disableRateLimit selects the hardened code alphabet/length, same as at
// generation time.
func NewVerificationCodeRule(code string, disableRateLimit bool) *VerificationCodeRule {
	r := &VerificationCodeRule{}
	r.Rule.SetContext(code)
	r.Rule.SetCondition(func(ctx any) bool {
		c, _ := ctx.(string)

		if c == "" {
			r.AddFailMessage(types.MsgVerificationCodeRequired)
			return false
		}
		if len(c) != utils.LoginCodeLength(disableRateLimit) {
			r.AddFailMessage(types.MsgVerificationCodeInvalidLength)
			return false
		}
		if !str.ContainsOnly(c, utils.LoginCodeGamma(disableRateLimit)) {
			r.AddFailMessage(types.MsgVerificationCodeInvalidCharacters)
			return false
		}
		return true
	})
	return r
}
