package auth

// AuthError represents a structured authentication error with a code,
// user-facing message, and internal error details for logging.
type AuthError struct {
	Code        string
	Message     string // User-facing message
	InternalErr error  // For logging only, never exposed to users
}

// Error implements the error interface, returning the user-facing message.
func (e AuthError) Error() string {
	return e.Message
}

// Error codes for consistent error handling
const (
	ErrCodeEmailSendFailed      = "EMAIL_SEND_FAILED"
	ErrCodeTokenStoreFailed     = "TOKEN_STORE_FAILED"
	ErrCodeValidationFailed     = "VALIDATION_FAILED"
	ErrCodeAuthenticationFailed = "AUTHENTICATION_FAILED"
	ErrCodeRegistrationFailed   = "REGISTRATION_FAILED"
	ErrCodeLogoutFailed         = "LOGOUT_FAILED"
	ErrCodeInternalError        = "INTERNAL_ERROR"
	ErrCodeCodeGenerationFailed = "CODE_GENERATION_FAILED"
	ErrCodeSerializationFailed  = "SERIALIZATION_FAILED"
	ErrCodePasswordResetFailed  = "PASSWORD_RESET_FAILED"
)

// NewEmailSendError creates an AuthError for email send failures.
func NewEmailSendError(err error) AuthError {
	return AuthError{
		Code:        ErrCodeEmailSendFailed,
		Message:     "Failed to send email. Please try again later",
		InternalErr: err,
	}
}

// NewTokenStoreError creates an AuthError for token store failures.
func NewTokenStoreError(err error) AuthError {
	return AuthError{
		Code:        ErrCodeTokenStoreFailed,
		Message:     "Failed to process request. Please try again later",
		InternalErr: err,
	}
}

// NewCodeGenerationError creates an AuthError for code generation failures.
func NewCodeGenerationError(err error) AuthError {
	return AuthError{
		Code:        ErrCodeCodeGenerationFailed,
		Message:     "Failed to generate verification code. Please try again later",
		InternalErr: err,
	}
}

// NewSerializationError creates an AuthError for data serialization failures.
func NewSerializationError(err error) AuthError {
	return AuthError{
		Code:        ErrCodeSerializationFailed,
		Message:     "Failed to process request. Please try again later",
		InternalErr: err,
	}
}

// NewAuthenticationError creates an AuthError for authentication failures.
func NewAuthenticationError(err error) AuthError {
	return AuthError{
		Code:        ErrCodeAuthenticationFailed,
		Message:     "Authentication failed",
		InternalErr: err,
	}
}

// NewRegistrationError creates an AuthError for registration failures.
func NewRegistrationError(err error) AuthError {
	return AuthError{
		Code:        ErrCodeRegistrationFailed,
		Message:     "Registration failed. Please try again later",
		InternalErr: err,
	}
}

// NewLogoutError creates an AuthError for logout failures.
func NewLogoutError(err error) AuthError {
	return AuthError{
		Code:        ErrCodeLogoutFailed,
		Message:     "Logout failed. Please try again later",
		InternalErr: err,
	}
}

// NewPasswordResetError creates an AuthError for password reset failures.
func NewPasswordResetError(err error) AuthError {
	return AuthError{
		Code:        ErrCodePasswordResetFailed,
		Message:     "Password reset failed. Please try again later",
		InternalErr: err,
	}
}

// NewInternalError creates an AuthError for generic internal errors.
func NewInternalError(err error) AuthError {
	return AuthError{
		Code:        ErrCodeInternalError,
		Message:     "Internal server error. Please try again later",
		InternalErr: err,
	}
}
