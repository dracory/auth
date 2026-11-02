package api_login

const (
	LoginOtpErrorCodeNone           LoginOtpErrorCode = ""
	LoginOtpErrorCodeValidation     LoginOtpErrorCode = "validation"
	LoginOtpErrorCodeCodeGeneration LoginOtpErrorCode = "code_generation"
	LoginOtpErrorCodeTokenStore     LoginOtpErrorCode = "token_store"
	LoginOtpErrorCodeEmailSend      LoginOtpErrorCode = "email_send"
	LoginOtpErrorCodeTooManySends   LoginOtpErrorCode = "too_many_sends"
)
