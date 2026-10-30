package rules

import (
	"crypto/subtle"

	"github.com/dracory/auth/types"
	"github.com/dracory/rule"
)

// PasswordsMatchRule validates that a password and its confirmation are
// identical, compared in constant time.
type PasswordsMatchRule struct {
	rule.Rule
}

// PasswordsMatchContext carries the inputs to PasswordsMatchRule.
type PasswordsMatchContext struct {
	Password        string
	PasswordConfirm string
}

// NewPasswordsMatchRule creates a PasswordsMatchRule for the given password
// pair.
func NewPasswordsMatchRule(password string, passwordConfirm string) *PasswordsMatchRule {
	r := &PasswordsMatchRule{}
	r.Rule.SetContext(PasswordsMatchContext{Password: password, PasswordConfirm: passwordConfirm})
	r.Rule.SetCondition(func(ctx any) bool {
		c, _ := ctx.(PasswordsMatchContext)
		if subtle.ConstantTimeCompare([]byte(c.Password), []byte(c.PasswordConfirm)) != 1 {
			r.AddFailMessage(types.MsgPasswordsDoNotMatch)
			return false
		}
		return true
	})
	return r
}
