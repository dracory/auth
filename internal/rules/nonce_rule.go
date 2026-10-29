package rules

import (
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/rule"
)

// NonceRule validates the submitted nonce: required and well-formed
// (32-char hex). Enforced before any temporary-key lookup so malformed
// input never reaches the store.
type NonceRule struct {
	rule.Rule
}

// NewNonceRule creates a NonceRule for the given nonce.
func NewNonceRule(nonce string) *NonceRule {
	r := &NonceRule{}
	r.Rule.SetContext(nonce)
	r.Rule.SetCondition(func(ctx any) bool {
		n, _ := ctx.(string)

		if n == "" {
			r.AddFailMessage(types.MsgNonceRequired)
			return false
		}
		if !utils.IsValidNonce(n) {
			r.AddFailMessage(types.MsgNonceInvalid)
			return false
		}
		return true
	})
	return r
}
