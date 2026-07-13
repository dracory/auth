package api_login_code_verify

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/auth/types"
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

	// Logger is used to log internal errors. If nil, slog.Default() is used.
	Logger *slog.Logger
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
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}

	result, perr := LoginCodeVerify(r.Context(), r, deps)
	if perr != nil {
		if perr.Err != nil {
			logger.Error("login code verify failed",
				slog.String("error", perr.Err.Error()),
				slog.String("code", string(perr.Code)),
			)
		}
		switch perr.Code {
		case LoginCodeVerifyErrorCodeValidation,
			LoginCodeVerifyErrorCodeCodeExpired:
			api.Respond(w, r, api.Error(perr.Message))
			return
		default:
			api.Respond(w, r, api.Error(types.MsgVerificationCodeExpired))
			return
		}
	}

	if deps.AuthenticateViaUsername == nil {
		api.Respond(w, r, api.Error(types.MsgFailedToProcess))
		return
	}

	deps.AuthenticateViaUsername(w, r, result.Email)
}

// ApiLoginCodeVerifyWithAuth is a convenience wrapper that allows callers to
// pass a types.AuthSharedInterface (such as authImplementation) instead of
// manually wiring Dependencies. It constructs the Dependencies struct using
// the interface accessors and preserves the existing behaviour.
func ApiLoginCodeVerifyWithAuth(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
	deps := Dependencies{
		DisableRateLimit: a.GetDisableRateLimit(),
		TemporaryKeyGet:  a.GetFuncTemporaryKeyGet(),
		Logger:           a.GetLogger(),
	}

	deps.AuthenticateViaUsername = func(w http.ResponseWriter, r *http.Request, email string) {
		a.AuthenticateViaUsername(w, r, email, "", "")
	}

	ApiLoginCodeVerify(w, r, deps)
}

// LoginCodeVerify encapsulates the core business logic for verifying a
// passwordless login code. It mirrors the original validation rules and error
// messages but does not perform authentication or write HTTP responses.
func LoginCodeVerify(ctx context.Context, r *http.Request, deps Dependencies) (*LoginCodeVerifyResult, *LoginCodeVerifyError) {
	verificationCode := req.GetStringTrimmed(r, "verification_code")

	if verificationCode == "" {
		return nil, &LoginCodeVerifyError{
			Code:    LoginCodeVerifyErrorCodeValidation,
			Message: types.MsgVerificationCodeRequired,
		}
	}

	if len(verificationCode) != utils.LoginCodeLength(deps.DisableRateLimit) {
		return nil, &LoginCodeVerifyError{
			Code:    LoginCodeVerifyErrorCodeValidation,
			Message: types.MsgVerificationCodeInvalidLength,
		}
	}

	if !str.ContainsOnly(verificationCode, utils.LoginCodeGamma(deps.DisableRateLimit)) {
		return nil, &LoginCodeVerifyError{
			Code:    LoginCodeVerifyErrorCodeValidation,
			Message: types.MsgVerificationCodeInvalidCharacters,
		}
	}

	if deps.TemporaryKeyGet == nil {
		return nil, &LoginCodeVerifyError{
			Code:    LoginCodeVerifyErrorCodeCodeExpired,
			Message: types.MsgVerificationCodeExpired,
		}
	}

	email, errCode := deps.TemporaryKeyGet(verificationCode)
	if errCode != nil {
		return nil, &LoginCodeVerifyError{
			Code:    LoginCodeVerifyErrorCodeCodeExpired,
			Message: types.MsgVerificationCodeExpired,
		}
	}

	return &LoginCodeVerifyResult{Email: email}, nil
}
