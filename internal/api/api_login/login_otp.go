package api_login

import (
	"context"
	"errors"
	"net/http"

	"github.com/dracory/auth/internal/rules"
	"github.com/dracory/auth/types"
	"github.com/dracory/req"
	"github.com/dracory/rule"

	"github.com/dracory/auth/utils"
)

// LoginOtpErrorCode categorizes the possible error sources in the otp login flow.
type LoginOtpErrorCode string

// LoginOtpError represents a structured error for the otp login flow.
// Message is intended to be user-facing for validation errors. For internal failures,
// callers are expected to wrap Err into their own error types.
type LoginOtpError struct {
	Code    LoginOtpErrorCode
	Message string
	Err     error
}

func (e *LoginOtpError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Code)
}

// LoginOtpResult represents a successful otp login operation.
type LoginOtpResult struct {
	SuccessMessage string
	// Nonce identifies the temporary-key under which "email:code" is stored.
	// The client must present it alongside the code on verification.
	Nonce string
}

// LoginOtp contains the core business logic of the otp login API.
// It performs input validation, code generation, temporary token storage and prepares
// the email contents. It does *not* perform logging; callers are responsible for
// logging and mapping structured errors to their own error types.
func loginOtp(ctx context.Context, r *http.Request, deps LoginOtpDeps) (*LoginOtpResult, *LoginOtpError) {
	email := req.GetStringTrimmed(r, "email")

	for _, check := range []rule.RuleInterface{
		rules.NewRequiredFieldRule(email, types.MsgEmailRequired),
		rules.NewEmailFormatRule(email),
	} {
		if check.Fails() {
			return nil, &LoginOtpError{
				Code:    LoginOtpErrorCodeValidation,
				Message: check.FailMessageFirst(),
			}
		}
	}

	// Lead capture on submission — the email is not yet verified. The
	// hook is best-effort — wiring sites log failures.
	if deps.EmailCaptured != nil {
		_ = deps.EmailCaptured(ctx, email, false, types.UserAuthOptions{
			UserIp:    req.GetIP(r),
			UserAgent: r.UserAgent(),
		})
	}

	// Per-email send throttle — prevents the endpoint being used for email
	// bombing; enforced before code generation so a blocked email never
	// consumes a code or send quota.
	if check := rules.NewEmailSendThrottleRule(deps.TemporaryKeyGet, email, deps.EmailSendThrottle); check.Fails() {
		return nil, &LoginOtpError{
			Code:    LoginOtpErrorCodeTooManySends,
			Message: check.FailMessageFirst(),
		}
	}

	verificationCode, err := utils.GenerateVerificationCode(deps.DisableRateLimit)
	if err != nil {
		return nil, &LoginOtpError{
			Code: LoginOtpErrorCodeCodeGeneration,
			Err:  err,
		}
	}

	nonce, err := utils.GenerateNonce()
	if err != nil {
		return nil, &LoginOtpError{
			Code: LoginOtpErrorCodeCodeGeneration,
			Err:  err,
		}
	}

	if deps.TemporaryKeySet == nil {
		return nil, &LoginOtpError{
			Code: LoginOtpErrorCodeTokenStore,
			Err:  errors.New("temporary key store is not configured"),
		}
	}

	expires := deps.ExpiresSeconds
	if expires <= 0 {
		// Fallback to one hour if the caller forgets to configure this; this keeps
		// the business logic package independent from the auth package constants
		// while still providing a safe default.
		expires = 3600
	}

	// The nonce — not the code — is the storage key, so a guessed code alone
	// cannot be redeemed; the client must also present the 128-bit nonce.
	if errTemp := deps.TemporaryKeySet(utils.LoginCodeKeyPrefix+nonce, email+":"+verificationCode, expires); errTemp != nil {
		return nil, &LoginOtpError{
			Code: LoginOtpErrorCodeTokenStore,
			Err:  errTemp,
		}
	}

	// Seed the per-nonce attempt counter alongside the code; verify enforces
	// utils.MaxCodeAttempts before invalidating the code.
	if errTemp := deps.TemporaryKeySet(utils.LoginCodeAttemptsKeyPrefix+nonce, "0", expires); errTemp != nil {
		return nil, &LoginOtpError{
			Code: LoginOtpErrorCodeTokenStore,
			Err:  errTemp,
		}
	}

	if deps.EmailTemplate == nil || deps.EmailSend == nil {
		return nil, &LoginOtpError{
			Code: LoginOtpErrorCodeEmailSend,
			Err:  errors.New("email template or sender is not configured"),
		}
	}

	emailContent := deps.EmailTemplate(ctx, email, verificationCode)

	// Queue-safe delivery preferred: when EmailEnqueue is set only the
	// opaque delivery key enters the queue; the code stays in the TTL-bound
	// temporary-key store for the worker to resolve.
	errEmail := utils.DeliverSecret(ctx, deps.TemporaryKeyGet, deps.TemporaryKeySet,
		deps.EmailEnqueue, deps.EmailSend, email, verificationCode,
		"Login Code", emailContent, expires)
	if errEmail != nil {
		return nil, &LoginOtpError{
			Code: LoginOtpErrorCodeEmailSend,
			Err:  errEmail,
		}
	}

	// Consume send quota only after the email was handed off successfully —
	// failed sends must not count against the user.
	if deps.TemporaryKeyGet != nil {
		_ = utils.RecordEmailSend(deps.TemporaryKeyGet, deps.TemporaryKeySet, email, deps.EmailSendThrottle)
	}

	return &LoginOtpResult{
		SuccessMessage: types.MsgLoginCodeSent,
		Nonce:          nonce,
	}, nil
}
