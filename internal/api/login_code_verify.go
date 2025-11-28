package api

import (
	"context"
	"errors"
	"net/http"

	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/req"
	"github.com/dracory/str"
)

// LoginCodeVerifyDeps defines dependencies for verifying a passwordless
// login code.
type LoginCodeVerifyDeps struct {
	DisableRateLimit bool

	TemporaryKeyGet func(key string) (string, error)
}

// LoginCodeVerifyErrorCode categorizes error sources.
type LoginCodeVerifyErrorCode string

const (
	LoginCodeVerifyErrorCodeNone        LoginCodeVerifyErrorCode = ""
	LoginCodeVerifyErrorCodeValidation  LoginCodeVerifyErrorCode = "validation"
	LoginCodeVerifyErrorCodeCodeExpired LoginCodeVerifyErrorCode = "code_expired"
)

// LoginCodeVerifyError represents a structured error in the login code
// verification flow.
type LoginCodeVerifyError struct {
	Code    LoginCodeVerifyErrorCode
	Message string
	Err     error
}

func (e *LoginCodeVerifyError) Error() string {
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

// LoginCodeVerifyResult represents a successful verification.
type LoginCodeVerifyResult struct {
	Email string
}

// LoginCodeVerify encapsulates the core business logic for verifying a
// passwordless login code. It mirrors the original validation rules and error
// messages but does not perform authentication or write HTTP responses.
func LoginCodeVerify(ctx context.Context, r *http.Request, deps LoginCodeVerifyDeps) (*LoginCodeVerifyResult, *LoginCodeVerifyError) {
	verificationCode := req.GetStringTrimmed(r, "verification_code")

	if verificationCode == "" {
		return nil, &LoginCodeVerifyError{
			Code:    LoginCodeVerifyErrorCodeValidation,
			Message: "Verification code is required field",
		}
	}

	if len(verificationCode) != authutils.LoginCodeLength(deps.DisableRateLimit) {
		return nil, &LoginCodeVerifyError{
			Code:    LoginCodeVerifyErrorCodeValidation,
			Message: "Verification code is invalid length",
		}
	}

	if !str.ContainsOnly(verificationCode, authutils.LoginCodeGamma(deps.DisableRateLimit)) {
		return nil, &LoginCodeVerifyError{
			Code:    LoginCodeVerifyErrorCodeValidation,
			Message: "Verification code contains invalid characters",
		}
	}

	if deps.TemporaryKeyGet == nil {
		return nil, &LoginCodeVerifyError{
			Code:    LoginCodeVerifyErrorCodeCodeExpired,
			Message: "Verification code has expired",
			Err:     errors.New("temporary key store is not configured"),
		}
	}

	email, errCode := deps.TemporaryKeyGet(verificationCode)
	if errCode != nil {
		return nil, &LoginCodeVerifyError{
			Code:    LoginCodeVerifyErrorCodeCodeExpired,
			Message: "Verification code has expired",
			Err:     errCode,
		}
	}

	return &LoginCodeVerifyResult{Email: email}, nil
}
