package rules

import (
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/rule"
)

// EmailSendThrottleRule enforces the per-email send quota — the check that
// every code-send path (login, register, password restore) performs before
// generating a code, so the endpoint cannot be used for email bombing.
// A nil get function (unconfigured store) fails open like a store read error.
type EmailSendThrottleRule struct {
	rule.Rule
}

// EmailSendThrottleContext carries the inputs to EmailSendThrottleRule.
type EmailSendThrottleContext struct {
	Get   func(key string) (string, error)
	Email string
}

// NewEmailSendThrottleRule creates an EmailSendThrottleRule for the given
// store getter and email.
func NewEmailSendThrottleRule(get func(key string) (string, error), email string) *EmailSendThrottleRule {
	r := &EmailSendThrottleRule{}
	r.Rule.SetContext(EmailSendThrottleContext{Get: get, Email: email})
	r.Rule.SetCondition(func(ctx any) bool {
		c, _ := ctx.(EmailSendThrottleContext)
		if c.Get == nil {
			return true // unconfigured store: no throttle possible
		}
		if !utils.CanSendEmail(c.Get, c.Email) {
			r.AddFailMessage(types.MsgTooManySends)
			return false
		}
		return true
	})
	return r
}
