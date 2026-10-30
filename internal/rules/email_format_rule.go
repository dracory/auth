package rules

import (
	"github.com/dracory/auth/utils"
	"github.com/dracory/rule"
)

// EmailFormatRule validates that a submitted email address has a valid format.
type EmailFormatRule struct {
	rule.Rule
}

// NewEmailFormatRule creates an EmailFormatRule for the given email address.
func NewEmailFormatRule(email string) *EmailFormatRule {
	r := &EmailFormatRule{}
	r.Rule.SetContext(email)
	r.Rule.SetCondition(func(ctx any) bool {
		e, _ := ctx.(string)
		if msg := utils.ValidateEmailFormat(e); msg != "" {
			r.AddFailMessage(msg)
			return false
		}
		return true
	})
	return r
}
