package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	authtypes "github.com/dracory/auth/types"
	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/req"
	"github.com/dracory/str"
)

// RegisterCodeVerifyDeps defines the dependencies required for verifying a
// registration code and creating the user account.
type RegisterCodeVerifyDeps struct {
	DisableRateLimit bool

	TemporaryKeyGet func(key string) (string, error)

	PasswordStrength *authtypes.PasswordStrengthConfig

	Passwordless bool

	PasswordlessUserRegister func(ctx context.Context, email, firstName, lastName string) error
	UserRegister             func(ctx context.Context, email, password, firstName, lastName string) error
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

// RegisterCodeVerify encapsulates the core business logic for verifying a
// registration code, performing optional password validation and creating the
// user. It does not log or write HTTP responses.
func RegisterCodeVerify(ctx context.Context, r *http.Request, deps RegisterCodeVerifyDeps) (*RegisterCodeVerifyResult, *RegisterCodeVerifyError) {
	verificationCode := req.GetStringTrimmed(r, "verification_code")

	// Input validation mirrors the original behaviour/messages.
	if verificationCode == "" {
		return nil, &RegisterCodeVerifyError{
			Code:    RegisterCodeVerifyErrorCodeValidation,
			Message: "Verification code is required field",
		}
	}

	if len(verificationCode) != authutils.LoginCodeLength(deps.DisableRateLimit) {
		return nil, &RegisterCodeVerifyError{
			Code:    RegisterCodeVerifyErrorCodeValidation,
			Message: "Verification code is invalid length",
		}
	}

	if !str.ContainsOnly(verificationCode, authutils.LoginCodeGamma(deps.DisableRateLimit)) {
		return nil, &RegisterCodeVerifyError{
			Code:    RegisterCodeVerifyErrorCodeValidation,
			Message: "Verification code contains invalid characters",
		}
	}

	if deps.TemporaryKeyGet == nil {
		return nil, &RegisterCodeVerifyError{
			Code: RegisterCodeVerifyErrorCodeCodeExpired,
			Err:  errors.New("temporary key store is not configured"),
		}
	}

	registerJSON, errCode := deps.TemporaryKeyGet(verificationCode)
	if errCode != nil {
		return nil, &RegisterCodeVerifyError{
			Code:    RegisterCodeVerifyErrorCodeCodeExpired,
			Message: "Verification code has expired",
			Err:     errCode,
		}
	}

	// Unmarshal the stored JSON (string) into a map, as in the original
	// implementation, to preserve behaviour.
	registerMap := map[string]any{}
	if errJSON := json.Unmarshal([]byte(registerJSON), &registerMap); errJSON != nil {
		return nil, &RegisterCodeVerifyError{
			Code:    RegisterCodeVerifyErrorCodeDeserialize,
			Message: "Serialized format is malformed",
			Err:     errJSON,
		}
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
