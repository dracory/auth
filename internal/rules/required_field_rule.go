package rules

import "github.com/dracory/rule"

// RequiredFieldRule validates that a submitted field is non-empty.
// The fail message is supplied by the caller (per-field copy differs).
type RequiredFieldRule struct {
	rule.Rule
}

// NewRequiredFieldRule creates a RequiredFieldRule that fails with the
// given message when value is empty.
func NewRequiredFieldRule(value string, message string) *RequiredFieldRule {
	r := &RequiredFieldRule{}
	r.Rule.SetContext(value)
	r.Rule.SetCondition(func(ctx any) bool {
		v, _ := ctx.(string)
		if v == "" {
			r.AddFailMessage(message)
			return false
		}
		return true
	})
	return r
}
