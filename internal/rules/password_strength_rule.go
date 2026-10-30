package rules

import (
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/rule"
)

// PasswordStrengthRule validates that a submitted password satisfies the
// configured password strength requirements.
type PasswordStrengthRule struct {
	rule.Rule
}

// PasswordStrengthContext carries inputs for PasswordStrengthRule.
type PasswordStrengthContext struct {
	Password string
	Config   *types.PasswordStrengthConfig
}

// NewPasswordStrengthRule creates a PasswordStrengthRule for the given password
// and password strength configuration.
func NewPasswordStrengthRule(password string, cfg *types.PasswordStrengthConfig) *PasswordStrengthRule {
	r := &PasswordStrengthRule{}
	r.Rule.SetContext(PasswordStrengthContext{Password: password, Config: cfg})
	r.Rule.SetCondition(func(ctx any) bool {
		c, _ := ctx.(PasswordStrengthContext)
		if err := utils.ValidatePasswordStrength(c.Password, c.Config); err != nil {
			r.AddFailMessage(err.Error())
			return false
		}
		return true
	})
	return r
}
