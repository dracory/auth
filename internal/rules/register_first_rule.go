package rules

import (
	"context"

	"github.com/dracory/auth/types"
	"github.com/dracory/rule"
)

// RegisterFirstRule decides whether a verified email should be sent to
// the register form (verified-email token → create on submit) instead of
// being authenticated immediately. It applies only when the
// register-first policy is enabled AND every required dependency is
// configured AND the email has no existing account.
//
// When the policy is enabled but cannot be enforced — missing deps or a
// failed existence check — the rule fails in a *degraded* state: callers
// fall back to session-first registration and should log the Reason
// loudly, because the fallback produces exactly the incomplete user row
// the policy exists to prevent. An existing account fails normally
// (normal login), without degradation.
//
// Shared by the OTP code-verify and magic-link callback flows so the
// policy decision is named, tested, and identical in both places.
type RegisterFirstRule struct {
	rule.Rule
	degraded bool
	reason   string
}

// RegisterFirstContext carries the inputs to RegisterFirstRule.
type RegisterFirstContext struct {
	// Enabled is the resolved CreateUserOnRegister policy flag.
	Enabled bool
	// UserExists reports whether the email already has an account.
	UserExists func(ctx context.Context, email string, options types.UserAuthOptions) (bool, error)
	// TemporaryKeySet backs the verified-email token store.
	TemporaryKeySet func(key string, value string, expiresSeconds int) error
	// Ctx is the request context for the existence check.
	Ctx context.Context
	// Email is the verified address being decided on.
	Email string
	// Options carries request attribution for the existence check.
	Options types.UserAuthOptions
}

// NewRegisterFirstRule creates a RegisterFirstRule. Evaluate once with
// Fails(); when it fails, Degraded + Reason describe warn-worthy
// policy-degradation fallbacks.
func NewRegisterFirstRule(
	enabled bool,
	userExists func(ctx context.Context, email string, options types.UserAuthOptions) (bool, error),
	tempKeySet func(key string, value string, expiresSeconds int) error,
	ctx context.Context,
	email string,
	options types.UserAuthOptions,
) *RegisterFirstRule {
	r := &RegisterFirstRule{}
	r.Rule.SetContext(RegisterFirstContext{
		Enabled:         enabled,
		UserExists:      userExists,
		TemporaryKeySet: tempKeySet,
		Ctx:             ctx,
		Email:           email,
		Options:         options,
	})
	r.Rule.SetCondition(func(any) bool {
		if !enabled {
			return false
		}
		if userExists == nil {
			r.degraded = true
			r.reason = "CreateUserOnRegister set but UserExists is not configured — falling back to session-first registration"
			return false
		}
		if tempKeySet == nil {
			r.degraded = true
			r.reason = "CreateUserOnRegister set but TemporaryKeySet is not configured — falling back to session-first registration"
			return false
		}
		exists, err := userExists(ctx, email, options)
		if err != nil {
			r.degraded = true
			r.reason = "CreateUserOnRegister set but UserExists failed — falling back to session-first registration: " + err.Error()
			return false
		}
		return !exists
	})
	return r
}

// Degraded reports whether the last Fails() evaluation failed because
// the policy could not be enforced (missing deps or check error) rather
// than because it does not apply. Callers should log Reason at Warn.
func (r *RegisterFirstRule) Degraded() bool {
	return r.degraded
}

// Reason returns the degradation explanation set during the last Fails()
// evaluation, or "" when the rule applied or was simply not enabled.
func (r *RegisterFirstRule) Reason() string {
	return r.reason
}
