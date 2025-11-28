package api_password_restore

const (
	PasswordRestoreErrorCodeNone          PasswordRestoreErrorCode = ""
	PasswordRestoreErrorCodeValidation    PasswordRestoreErrorCode = "validation"
	PasswordRestoreErrorCodeUserLookup    PasswordRestoreErrorCode = "user_lookup"
	PasswordRestoreErrorCodeCodeGenerate  PasswordRestoreErrorCode = "code_generation"
	PasswordRestoreErrorCodeTokenStore    PasswordRestoreErrorCode = "token_store"
	PasswordRestoreErrorCodeEmailSend     PasswordRestoreErrorCode = "email_send"
	PasswordRestoreErrorCodeInternalError PasswordRestoreErrorCode = "internal"
)
