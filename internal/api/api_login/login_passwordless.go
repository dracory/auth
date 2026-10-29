package api_login

import (
	"context"
	"errors"
	"net/http"

	"github.com/dracory/auth/types"
	"github.com/dracory/req"

	"github.com/dracory/auth/utils"
)

// LoginPasswordlessErrorCode categorizes the possible error sources in the passwordless login flow.
type LoginPasswordlessErrorCode string

// LoginPasswordlessError represents a structured error for the passwordless login flow.
// Message is intended to be user-facing for validation errors. For internal failures,
// callers are expected to wrap Err into their own error types.
type LoginPasswordlessError struct {
	Code    LoginPasswordlessErrorCode
	Message string
	Err     error
}

func (e *LoginPasswordlessError) Error() string {
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

// LoginPasswordlessResult represents a successful passwordless login operation.
type LoginPasswordlessResult struct {
	SuccessMessage string
	// Nonce identifies the temporary-key under which "email:code" is stored.
	// The client must present it alongside the code on verification.
	Nonce string
}

// LoginPasswordless contains the core business logic of the passwordless login API.
// It performs input validation, code generation, temporary token storage and prepares
// the email contents. It does *not* perform logging; callers are responsible for
// logging and mapping structured errors to their own error types.
func loginPasswordless(ctx context.Context, r *http.Request, deps LoginPasswordlessDeps) (*LoginPasswordlessResult, *LoginPasswordlessError) {
	email := req.GetStringTrimmed(r, "email")

	if email == "" {
		return nil, &LoginPasswordlessError{
			Code:    LoginPasswordlessErrorCodeValidation,
			Message: types.MsgEmailRequired,
		}
	}

	if msg := utils.ValidateEmailFormat(email); msg != "" {
		return nil, &LoginPasswordlessError{
			Code:    LoginPasswordlessErrorCodeValidation,
			Message: msg,
		}
	}

	verificationCode, err := utils.GenerateVerificationCode(deps.DisableRateLimit)
	if err != nil {
		return nil, &LoginPasswordlessError{
			Code: LoginPasswordlessErrorCodeCodeGeneration,
			Err:  err,
		}
	}

	nonce, err := utils.GenerateNonce()
	if err != nil {
		return nil, &LoginPasswordlessError{
			Code: LoginPasswordlessErrorCodeCodeGeneration,
			Err:  err,
		}
	}

	if deps.TemporaryKeySet == nil {
		return nil, &LoginPasswordlessError{
			Code: LoginPasswordlessErrorCodeTokenStore,
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
		return nil, &LoginPasswordlessError{
			Code: LoginPasswordlessErrorCodeTokenStore,
			Err:  errTemp,
		}
	}

	// Seed the per-nonce attempt counter alongside the code; verify enforces
	// utils.MaxCodeAttempts before invalidating the code.
	if errTemp := deps.TemporaryKeySet(utils.LoginCodeAttemptsKeyPrefix+nonce, "0", expires); errTemp != nil {
		return nil, &LoginPasswordlessError{
			Code: LoginPasswordlessErrorCodeTokenStore,
			Err:  errTemp,
		}
	}

	if deps.EmailTemplate == nil || deps.EmailSend == nil {
		return nil, &LoginPasswordlessError{
			Code: LoginPasswordlessErrorCodeEmailSend,
			Err:  errors.New("email template or sender is not configured"),
		}
	}

	emailContent := deps.EmailTemplate(ctx, email, verificationCode)

	if errEmail := deps.EmailSend(ctx, email, "Login Code", emailContent); errEmail != nil {
		return nil, &LoginPasswordlessError{
			Code: LoginPasswordlessErrorCodeEmailSend,
			Err:  errEmail,
		}
	}

	return &LoginPasswordlessResult{
		SuccessMessage: types.MsgLoginCodeSent,
		Nonce:          nonce,
	}, nil
}
