package rules

import (
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/rule"
)

// SafeReturnPathRule validates that a return path is a safe relative path —
// never an external URL. Callers may use the failure to drop the path
// silently (as magic-link send does) or to surface a rejection message;
// an empty path passes since there is nothing to propagate.
type SafeReturnPathRule struct {
	rule.Rule
}

// NewSafeReturnPathRule creates a SafeReturnPathRule for the given path.
func NewSafeReturnPathRule(returnPath string) *SafeReturnPathRule {
	r := &SafeReturnPathRule{}
	r.Rule.SetContext(returnPath)
	r.Rule.SetCondition(func(ctx any) bool {
		p, _ := ctx.(string)
		if p == "" {
			return true
		}
		if !utils.IsSafeReturnPath(p) {
			r.AddFailMessage(types.MsgUnsafeReturnPath)
			return false
		}
		return true
	})
	return r
}
