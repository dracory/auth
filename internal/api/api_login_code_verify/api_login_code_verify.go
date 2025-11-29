package api_login_code_verify

import (
	"context"
	"errors"
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/auth/utils"
	"github.com/dracory/req"
	"github.com/dracory/str"
)

// Dependencies defines dependencies for verifying a passwordless
// login code and authenticating the user.
type Dependencies struct {
	DisableRateLimit bool

	TemporaryKeyGet func(key string) (string, error)

	// AuthenticateViaUsername is called on successful code verification
	// to perform authentication (token generation, cookies, etc.) and send
	// the final HTTP response.
	AuthenticateViaUsername func(w http.ResponseWriter, r *http.Request, email string)
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

// ApiLoginCodeVerify is the HTTP-level helper that wires request/response
// handling to the core LoginCodeVerify business logic using the provided
// dependencies.
func ApiLoginCodeVerify(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	result, perr := LoginCodeVerify(r.Context(), r, deps)
	if perr != nil {
		switch perr.Code {
		case LoginCodeVerifyErrorCodeValidation,
			LoginCodeVerifyErrorCodeCodeExpired:
			api.Respond(w, r, api.Error(perr.Message))
			return
		default:
			api.Respond(w, r, api.Error("Verification code has expired"))
			return
		}
	}

	if deps.AuthenticateViaUsername == nil {
		api.Respond(w, r, api.Error("Failed to process request. Please try again later"))
		return
	}

	deps.AuthenticateViaUsername(w, r, result.Email)
}

// LoginCodeVerify encapsulates the core business logic for verifying a
// passwordless login code. It mirrors the original validation rules and error
// messages but does not perform authentication or write HTTP responses.
func LoginCodeVerify(ctx context.Context, r *http.Request, deps Dependencies) (*LoginCodeVerifyResult, *LoginCodeVerifyError) {
	verificationCode := req.GetStringTrimmed(r, "verification_code")

	if verificationCode == "" {
		return nil, &LoginCodeVerifyError{
			Code:    LoginCodeVerifyErrorCodeValidation,
			Message: "Verification code is required field",
		}
	}

	if len(verificationCode) != utils.LoginCodeLength(deps.DisableRateLimit) {
		return nil, &LoginCodeVerifyError{
			Code:    LoginCodeVerifyErrorCodeValidation,
			Message: "Verification code is invalid length",
		}
	}

	if !str.ContainsOnly(verificationCode, utils.LoginCodeGamma(deps.DisableRateLimit)) {
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
