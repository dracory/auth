package api_login

const (
	LoginPasswordlessErrorCodeNone           LoginPasswordlessErrorCode = ""
	LoginPasswordlessErrorCodeValidation     LoginPasswordlessErrorCode = "validation"
	LoginPasswordlessErrorCodeCodeGeneration LoginPasswordlessErrorCode = "code_generation"
	LoginPasswordlessErrorCodeTokenStore     LoginPasswordlessErrorCode = "token_store"
	LoginPasswordlessErrorCodeEmailSend      LoginPasswordlessErrorCode = "email_send"
)
