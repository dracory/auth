package api_register

import (
	"context"
	"encoding/json"
	"errors"
	"html"
	"net/http"

	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

// RegisterPasswordlessInitDeps defines the dependencies required for the
// passwordless registration init (sending verification code).
type RegisterPasswordlessInitDeps struct {
	DisableRateLimit bool

	TemporaryKeySet func(key string, value string, expiresSeconds int) error
	ExpiresSeconds  int

	EmailTemplate func(ctx context.Context, email string, verificationCode string) string
	EmailSend     func(ctx context.Context, email string, subject string, body string) error
}

// RegisterPasswordlessInitErrorCode categorizes possible error sources.
type RegisterPasswordlessInitErrorCode string

const (
	RegisterPasswordlessInitErrorCodeNone           RegisterPasswordlessInitErrorCode = ""
	RegisterPasswordlessInitErrorCodeValidation     RegisterPasswordlessInitErrorCode = "validation"
	RegisterPasswordlessInitErrorCodeCodeGeneration RegisterPasswordlessInitErrorCode = "code_generation"
	RegisterPasswordlessInitErrorCodeSerialization  RegisterPasswordlessInitErrorCode = "serialization"
	RegisterPasswordlessInitErrorCodeTokenStore     RegisterPasswordlessInitErrorCode = "token_store"
	RegisterPasswordlessInitErrorCodeEmailSend      RegisterPasswordlessInitErrorCode = "email_send"
	RegisterPasswordlessInitErrorCodeInternal       RegisterPasswordlessInitErrorCode = "internal"
)

// RegisterPasswordlessInitError is a structured error for the registration init flow.
type RegisterPasswordlessInitError struct {
	Code    RegisterPasswordlessInitErrorCode
	Message string
	Err     error
}

func (e *RegisterPasswordlessInitError) Error() string {
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

// RegisterPasswordlessInitResult represents a successful registration init.
type RegisterPasswordlessInitResult struct {
	SuccessMessage string
}

// RegisterPasswordlessInit contains the core business logic of the passwordless
// registration init API. It validates input, generates the verification code,
// stores a temporary JSON payload and sends the email. It intentionally does
// not perform any logging; callers are responsible for mapping structured
// errors to their own logging and HTTP responses.
func RegisterPasswordlessInit(ctx context.Context, r *http.Request, deps RegisterPasswordlessInitDeps) (*RegisterPasswordlessInitResult, *RegisterPasswordlessInitError) {
	email := req.GetStringTrimmed(r, "email")
	firstName := html.EscapeString(req.GetStringTrimmed(r, "first_name"))
	lastName := html.EscapeString(req.GetStringTrimmed(r, "last_name"))

	if firstName == "" {
		return nil, &RegisterPasswordlessInitError{
			Code:    RegisterPasswordlessInitErrorCodeValidation,
			Message: "First name is required field",
		}
	}

	if lastName == "" {
		return nil, &RegisterPasswordlessInitError{
			Code:    RegisterPasswordlessInitErrorCodeValidation,
			Message: "Last name is required field",
		}
	}

	if email == "" {
		return nil, &RegisterPasswordlessInitError{
			Code:    RegisterPasswordlessInitErrorCodeValidation,
			Message: "Email is required field",
		}
	}

	if msg := authutils.ValidateEmailFormat(email); msg != "" {
		return nil, &RegisterPasswordlessInitError{
			Code:    RegisterPasswordlessInitErrorCodeValidation,
			Message: msg,
		}
	}

	verificationCode, err := authutils.GenerateVerificationCode(deps.DisableRateLimit)
	if err != nil {
		return nil, &RegisterPasswordlessInitError{
			Code: RegisterPasswordlessInitErrorCodeCodeGeneration,
			Err:  err,
		}
	}

	payload, errJSON := json.Marshal(map[string]string{
		"email":      email,
		"first_name": firstName,
		"last_name":  lastName,
	})
	if errJSON != nil {
		return nil, &RegisterPasswordlessInitError{
			Code: RegisterPasswordlessInitErrorCodeSerialization,
			Err:  errJSON,
		}
	}

	if deps.TemporaryKeySet == nil {
		return nil, &RegisterPasswordlessInitError{
			Code: RegisterPasswordlessInitErrorCodeTokenStore,
			Err:  errors.New("temporary key store is not configured"),
		}
	}

	expires := deps.ExpiresSeconds
	if expires <= 0 {
		// fallback to one hour
		expires = 3600
	}

	if errTemp := deps.TemporaryKeySet(verificationCode, string(payload), expires); errTemp != nil {
		return nil, &RegisterPasswordlessInitError{
			Code: RegisterPasswordlessInitErrorCodeTokenStore,
			Err:  errTemp,
		}
	}

	if deps.EmailTemplate == nil || deps.EmailSend == nil {
		return nil, &RegisterPasswordlessInitError{
			Code: RegisterPasswordlessInitErrorCodeInternal,
			Err:  errors.New("email template or sender is not configured"),
		}
	}

	emailContent := deps.EmailTemplate(ctx, email, verificationCode)

	if errEmail := deps.EmailSend(ctx, email, "Registration Code", emailContent); errEmail != nil {
		return nil, &RegisterPasswordlessInitError{
			Code: RegisterPasswordlessInitErrorCodeEmailSend,
			Err:  errEmail,
		}
	}

	return &RegisterPasswordlessInitResult{
		SuccessMessage: "Registration code was sent successfully",
	}, nil
}
