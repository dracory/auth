package rules

import (
	"fmt"

	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/rule"
)

// CodeAttemptsLimitRule enforces the per-nonce attempt limit shared by the
// code-verify flows (login and register). The check runs before the payload
// lookup so a known nonce cannot be used to probe the code space until
// expiry. When the limit is exceeded the rule invalidates both the payload
// key and the counter key — the nonce is burned.
//
// Fail-open on read: a store error or a missing/corrupt counter resets to
// 0 — intentional, so codes issued before this feature (or an evicted
// counter key) still verify.
type CodeAttemptsLimitRule struct {
	rule.Rule
	attempts int
}

// CodeAttemptsLimitContext carries the inputs to CodeAttemptsLimitRule.
type CodeAttemptsLimitContext struct {
	// Get reads a key from the temporary key store.
	Get func(key string) (string, error)
	// Delete removes a key; preferred invalidation method when set.
	Delete func(key string) error
	// Set writes a key. Required for the limit to work: callers increment
	// the counter via Set, so a nil Set means the counter never grows and
	// the rule always observes 0 — MaxCodeAttempts silently becomes a
	// no-op. Also used as the invalidation fallback
	// ("empty value, 0 expiry" convention) when Delete is nil.
	Set func(key string, value string, expiresSeconds int) error
	// PayloadKey is the temporary key holding the code payload.
	PayloadKey string
	// AttemptsKey is the temporary key holding the attempt counter.
	AttemptsKey string
}

// NewCodeAttemptsLimitRule creates a CodeAttemptsLimitRule for the given
// store functions and keys.
func NewCodeAttemptsLimitRule(
	get func(key string) (string, error),
	deleteFn func(key string) error,
	set func(key string, value string, expiresSeconds int) error,
	payloadKey string,
	attemptsKey string,
) *CodeAttemptsLimitRule {
	r := &CodeAttemptsLimitRule{}
	r.Rule.SetContext(CodeAttemptsLimitContext{
		Get:         get,
		Delete:      deleteFn,
		Set:         set,
		PayloadKey:  payloadKey,
		AttemptsKey: attemptsKey,
	})
	r.Rule.SetCondition(func(ctx any) bool {
		c, _ := ctx.(CodeAttemptsLimitContext)
		if c.Get == nil {
			return true // unconfigured store: no limit possible
		}
		attempts := 0
		if attemptsStr, err := c.Get(c.AttemptsKey); err == nil {
			_, _ = fmt.Sscanf(attemptsStr, "%d", &attempts)
		}
		r.attempts = attempts
		if attempts >= utils.MaxCodeAttempts {
			_ = utils.InvalidateTempKey(c.Delete, c.Set, c.PayloadKey)
			_ = utils.InvalidateTempKey(c.Delete, c.Set, c.AttemptsKey)
			r.AddFailMessage(types.MsgTooManyAttempts)
			return false
		}
		return true
	})
	return r
}

// Attempts returns the counter value observed during the last evaluation —
// callers use it to write the incremented counter before the code compare.
// Returns 0 when the rule has not been evaluated yet.
func (r *CodeAttemptsLimitRule) Attempts() int {
	return r.attempts
}
