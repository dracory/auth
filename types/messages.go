package types

// Error messages for validation and internal failures.
// These constants centralize all user-facing strings so applications
// can override or localize them without modifying core business logic.
const (
	// Validation errors
	MsgEmailRequired                     = "Email is required field"
	MsgPasswordRequired                  = "Password is required field"
	MsgFirstNameRequired                 = "First name is required field"
	MsgLastNameRequired                  = "Last name is required field"
	MsgTokenRequired                     = "Token is required field"
	MsgVerificationCodeRequired          = "Verification code is required field"
	MsgPasswordsDoNotMatch               = "Passwords do not match"
	MsgEmailInvalid                      = "Email is invalid"
	MsgVerificationCodeInvalidLength     = "Verification code is invalid length"
	MsgVerificationCodeInvalidCharacters = "Verification code contains invalid characters"
	MsgVerificationCodeExpired           = "Verification code has expired"
	MsgSerializedFormatMalformed         = "Serialized format is malformed"

	// Auth errors
	MsgInvalidCredentials = "Invalid credentials"
	MsgUserNotFound       = "User not found"

	// Operation errors
	MsgRegistrationFailed          = "registration failed."
	MsgRegistrationFailedFn        = "registration failed. FuncUserRegister function not defined"
	MsgRegistrationFailedEmailTpl  = "registration failed. FuncEmailTemplateRegisterCode function not defined"
	MsgRegistrationFailedEmailSend = "registration failed. FuncEmailSend function not defined"
	MsgRegistrationFailedGeneric   = "Registration failed. Please try again later"
	MsgPasswordResetFailed         = "Password reset failed. Please try again later"
	MsgLogoutFailed                = "Logout failed. Please try again later"
	MsgPasswordValidationFailed    = "Password validation failed"

	// Generic internal errors
	MsgInternalServer        = "Internal server error. Please try again later"
	MsgFailedToProcess       = "Failed to process request. Please try again later"
	MsgFailedToGenerateCode  = "Failed to generate verification code. Please try again later"
	MsgFailedToSendEmail     = "Failed to send email. Please try again later"
	MsgLinkNotValidOrExpired = "Link not valid or expired"
	MsgTooManyRequests       = "Too many requests. Please try again later."

	// Email subjects
	EmailSubjectRegistrationCode = "Registration Code"

	// Success messages
	MsgLoginSuccess          = "login success"
	MsgRegistrationSuccess   = "registration success"
	MsgRegistrationCodeSent  = "Registration code was sent successfully"
	MsgLoginCodeSent         = "Login code was sent successfully"
	MsgPasswordResetLinkSent = "Password reset link was sent to your e-mail"
	MsgPasswordResetSuccess  = "Password has been reset successfully"

	// Impersonation messages
	MsgImpersonationStarted    = "impersonation started"
	MsgImpersonationStopped    = "impersonation stopped"
	MsgNotImpersonating        = "not currently impersonating"
	MsgAlreadyImpersonating    = "already impersonating — stop first"
	MsgImpersonationNotEnabled = "impersonation is not enabled"
	MsgImpersonationForbidden  = "impersonation is not allowed"
	MsgImpersonationFailed     = "impersonation failed. Please try again later"
)
