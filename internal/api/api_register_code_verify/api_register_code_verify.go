package api_register_code_verify

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sync"

	"github.com/dracory/api"
	types "github.com/dracory/auth/types"
	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/req"
	"github.com/dracory/str"
)

// Dependencies defines the dependencies required for verifying a registration
// code, creating the user account, and authenticating the user.
type Dependencies struct {
	DisableRateLimit bool

	TemporaryKeyGet func(key string) (string, error)

	// TemporaryKeySet is optional; when present the consumed code is
	// invalidated after a successful verification (single-use).
	TemporaryKeySet func(key string, value string, expiresSeconds int) error

	// TemporaryKeyDelete is optional; when present it is preferred over
	// TemporaryKeySet for invalidating consumed codes.
	TemporaryKeyDelete func(key string) error

	PasswordStrength *types.PasswordStrengthConfig

	Passwordless bool

	PasswordlessUserRegister func(ctx context.Context, email, firstName, lastName string) error
	UserRegister             func(ctx context.Context, email, password, firstName, lastName string) error

	// AuthenticateViaUsername is called on successful registration to
	// authenticate the user and produce the final HTTP response.
	AuthenticateViaUsername func(w http.ResponseWriter, r *http.Request, email, firstName, lastName string)

	// Logger is used to log internal errors. If nil, slog.Default() is used.
	Logger *slog.Logger
}

// RegisterCodeVerifyErrorCode categorizes error sources in the registration
// code verification flow.
type RegisterCodeVerifyErrorCode string

const (
	RegisterCodeVerifyErrorCodeNone               RegisterCodeVerifyErrorCode = ""
	RegisterCodeVerifyErrorCodeValidation         RegisterCodeVerifyErrorCode = "validation"
	RegisterCodeVerifyErrorCodeCodeExpired        RegisterCodeVerifyErrorCode = "code_expired"
	RegisterCodeVerifyErrorCodeDeserialize        RegisterCodeVerifyErrorCode = "deserialize"
	RegisterCodeVerifyErrorCodePasswordValidation RegisterCodeVerifyErrorCode = "password_validation"
	RegisterCodeVerifyErrorCodeRegister           RegisterCodeVerifyErrorCode = "register"
)

// RegisterCodeVerifyError represents a structured error.
type RegisterCodeVerifyError struct {
	Code    RegisterCodeVerifyErrorCode
	Message string
	Err     error
}

func (e *RegisterCodeVerifyError) Error() string {
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

// RegisterCodeVerifyResult represents a successful verification.
type RegisterCodeVerifyResult struct {
	Email     string
	FirstName string
	LastName  string
}

// ApiRegisterCodeVerify is the HTTP-level helper that wires
// request/response handling to the core RegisterCodeVerify business logic
// using the provided dependencies.
func ApiRegisterCodeVerify(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}

	result, perr := RegisterCodeVerify(r.Context(), r, deps)
	if perr != nil {
		if perr.Err != nil {
			logger.Error("register code verify failed",
				slog.String("error", perr.Err.Error()),
				slog.String("code", string(perr.Code)),
			)
		}
		switch perr.Code {
		case RegisterCodeVerifyErrorCodeValidation,
			RegisterCodeVerifyErrorCodeCodeExpired,
			RegisterCodeVerifyErrorCodeDeserialize:
			api.Respond(w, r, api.Error(perr.Message))
			return
		case RegisterCodeVerifyErrorCodePasswordValidation:
			// Preserve behaviour of returning the validation error string.
			if perr.Err != nil {
				api.Respond(w, r, api.Error(perr.Err.Error()))
			} else {
				api.Respond(w, r, api.Error(types.MsgPasswordValidationFailed))
			}
			return
		case RegisterCodeVerifyErrorCodeRegister:
			// Map to the same user-facing message as NewRegistrationError.
			api.Respond(w, r, api.Error(types.MsgRegistrationFailedGeneric))
			return
		default:
			// Map to the same user-facing message pattern as NewInternalError.
			api.Respond(w, r, api.Error(types.MsgInternalServer))
			return
		}
	}

	if deps.AuthenticateViaUsername == nil {
		api.Respond(w, r, api.Error(types.MsgFailedToProcess))
		return
	}

	// Delegate final authentication and response to the provided callback so
	// that existing behaviour (including token generation and cookies) is
	// preserved.
	deps.AuthenticateViaUsername(w, r, result.Email, result.FirstName, result.LastName)
}

// ApiRegisterCodeVerifyWithAuth is a convenience wrapper that allows callers
// to pass a types.AuthSharedInterface (such as authImplementation) instead of
// manually wiring Dependencies. It constructs the Dependencies struct using
// the interface accessors and preserves the existing behaviour.
func ApiRegisterCodeVerifyWithAuth(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
	deps := Dependencies{
		DisableRateLimit: a.GetDisableRateLimit(),
		TemporaryKeyGet:  a.GetFuncTemporaryKeyGet(),
		TemporaryKeySet:  a.GetFuncTemporaryKeySet(),
		PasswordStrength: a.GetPasswordStrength(),
		Passwordless:     a.IsPasswordless(),
		Logger:           a.GetLogger(),
	}
	if deleter, ok := a.(interface {
		GetFuncTemporaryKeyDelete() func(key string) error
	}); ok {
		deps.TemporaryKeyDelete = deleter.GetFuncTemporaryKeyDelete()
	}

	if fn := a.GetPasswordlessUserRegister(); fn != nil {
		deps.PasswordlessUserRegister = func(ctx context.Context, email, firstName, lastName string) error {
			return fn(ctx, email, firstName, lastName, types.UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		}
	}

	if fn := a.GetFuncUserRegister(); fn != nil {
		deps.UserRegister = func(ctx context.Context, email, password, firstName, lastName string) error {
			return fn(ctx, email, password, firstName, lastName, types.UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		}
	}

	deps.AuthenticateViaUsername = func(w http.ResponseWriter, r *http.Request, email, firstName, lastName string) {
		a.AuthenticateViaUsername(w, r, email, firstName, lastName)
	}

	ApiRegisterCodeVerify(w, r, deps)
}

// RegisterCodeVerify encapsulates the core business logic for verifying a
// registration code, performing optional password validation and creating the
// user. It does not log or write HTTP responses.
func RegisterCodeVerify(ctx context.Context, r *http.Request, deps Dependencies) (*RegisterCodeVerifyResult, *RegisterCodeVerifyError) {
	verificationCode := req.GetStringTrimmed(r, "verification_code")
	nonce := req.GetStringTrimmed(r, "nonce")

	// Input validation mirrors the original behaviour/messages.
	if verificationCode == "" {
		return nil, &RegisterCodeVerifyError{
			Code:    RegisterCodeVerifyErrorCodeValidation,
			Message: types.MsgVerificationCodeRequired,
		}
	}

	if nonce == "" {
		return nil, &RegisterCodeVerifyError{
			Code:    RegisterCodeVerifyErrorCodeValidation,
			Message: types.MsgNonceRequired,
		}
	}

	if !authutils.IsValidNonce(nonce) {
		return nil, &RegisterCodeVerifyError{
			Code:    RegisterCodeVerifyErrorCodeValidation,
			Message: types.MsgNonceInvalid,
		}
	}

	if len(verificationCode) != authutils.LoginCodeLength(deps.DisableRateLimit) {
		return nil, &RegisterCodeVerifyError{
			Code:    RegisterCodeVerifyErrorCodeValidation,
			Message: types.MsgVerificationCodeInvalidLength,
		}
	}

	if !str.ContainsOnly(verificationCode, authutils.LoginCodeGamma(deps.DisableRateLimit)) {
		return nil, &RegisterCodeVerifyError{
			Code:    RegisterCodeVerifyErrorCodeValidation,
			Message: types.MsgVerificationCodeInvalidCharacters,
		}
	}

	if deps.TemporaryKeyGet == nil {
		return nil, &RegisterCodeVerifyError{
			Code: RegisterCodeVerifyErrorCodeCodeExpired,
			Err:  errors.New("temporary key store is not configured"),
		}
	}

	// The nonce — not the code — is the storage key. The stored payload
	// carries the code for constant-time comparison.
	// verifyMu serializes get-compare-invalidate to prevent concurrent
	// reuse of the same nonce (TOCTOU, best-effort within one process).
	cacheKey := authutils.RegisterCodeKeyPrefix + nonce
	verifyMu.Lock()
	defer verifyMu.Unlock()

	registerJSON, errCode := deps.TemporaryKeyGet(cacheKey)
	if errCode != nil || registerJSON == "" {
		return nil, &RegisterCodeVerifyError{
			Code:    RegisterCodeVerifyErrorCodeCodeExpired,
			Message: types.MsgVerificationCodeExpired,
		}
	}

	// Unmarshal the stored JSON (string) into a map, as in the original
	// implementation, to preserve behaviour.
	registerMap := map[string]any{}
	if errJSON := json.Unmarshal([]byte(registerJSON), &registerMap); errJSON != nil {
		return nil, &RegisterCodeVerifyError{
			Code:    RegisterCodeVerifyErrorCodeDeserialize,
			Message: types.MsgSerializedFormatMalformed,
			Err:     errJSON,
		}
	}

	storedCode, _ := registerMap["code"].(string)
	if storedCode == "" || subtle.ConstantTimeCompare([]byte(storedCode), []byte(verificationCode)) != 1 {
		return nil, &RegisterCodeVerifyError{
			Code:    RegisterCodeVerifyErrorCodeCodeExpired,
			Message: types.MsgVerificationCodeExpired,
		}
	}

	// Single-use: invalidate the consumed code; prefer a real delete.
	if err := authutils.InvalidateTempKey(deps.TemporaryKeyDelete, deps.TemporaryKeySet, cacheKey); err != nil {
		logger := deps.Logger
		if logger == nil {
			logger = slog.Default()
		}
		logger.Error("register code verify: failed to invalidate consumed code",
			slog.String("error", err.Error()))
	}

	email := ""
	if val, ok := registerMap["email"]; ok {
		if s, ok := val.(string); ok {
			email = s
		}
	}

	firstName := ""
	if val, ok := registerMap["first_name"]; ok {
		if s, ok := val.(string); ok {
			firstName = s
		}
	}

	lastName := ""
	if val, ok := registerMap["last_name"]; ok {
		if s, ok := val.(string); ok {
			lastName = s
		}
	}

	password := ""
	if val, ok := registerMap["password"]; ok {
		if s, ok := val.(string); ok {
			password = s
		}
	}

	// Perform registration
	var errRegister error

	if deps.Passwordless {
		if deps.PasswordlessUserRegister == nil {
			return nil, &RegisterCodeVerifyError{
				Code: RegisterCodeVerifyErrorCodeRegister,
				Err:  errors.New("passwordless user register function is not configured"),
			}
		}
		errRegister = deps.PasswordlessUserRegister(ctx, email, firstName, lastName)
	} else {
		// Username/password flow with strength validation
		if deps.PasswordStrength != nil {
			if err := authutils.ValidatePasswordStrength(password, deps.PasswordStrength); err != nil {
				return nil, &RegisterCodeVerifyError{
					Code: RegisterCodeVerifyErrorCodePasswordValidation,
					Err:  err,
				}
			}
		}

		if deps.UserRegister == nil {
			return nil, &RegisterCodeVerifyError{
				Code: RegisterCodeVerifyErrorCodeRegister,
				Err:  errors.New("user register function is not configured"),
			}
		}

		errRegister = deps.UserRegister(ctx, email, password, firstName, lastName)
	}

	if errRegister != nil {
		return nil, &RegisterCodeVerifyError{
			Code: RegisterCodeVerifyErrorCodeRegister,
			Err:  errRegister,
		}
	}

	return &RegisterCodeVerifyResult{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}, nil
}

// verifyMu serializes the get-compare-invalidate sequence in
// RegisterCodeVerify (see comment at the lookup site).
var verifyMu sync.Mutex
