package api_logout

import "context"

// LogoutErrorCode categorizes error sources in the logout flow.
type LogoutErrorCode string

const (
	LogoutErrorCodeNone        LogoutErrorCode = ""
	LogoutErrorCodeTokenLookup LogoutErrorCode = "token_lookup"
	LogoutErrorCodeUserLogout  LogoutErrorCode = "user_logout"
)

// LogoutError represents a structured error for the logout flow.
type LogoutError struct {
	Code   LogoutErrorCode
	Err    error
	UserID string
}

func (e *LogoutError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return string(e.Code)
}

// Logout contains the core business logic for logging out a user based on
// an authentication token. It does not interact with HTTP, cookies or logs;
// these responsibilities remain with the caller.
//
// Behaviour:
//   - Token lookup is always delegated to UserFromToken, even for empty
//     tokens. This mirrors the original apiLogout semantics where tests may
//     override token validation behaviour.
//   - If token lookup fails, a LogoutError with CodeTokenLookup is returned.
//   - If a user ID is resolved and LogoutUser fails, a LogoutError with
//     CodeUserLogout is returned.
//   - Otherwise, nil is returned to indicate success.
func ApiLogout(ctx context.Context, token string, dependencies Dependencies) *LogoutError {
	if dependencies.UserFromToken == nil {
		return &LogoutError{Code: LogoutErrorCodeTokenLookup}
	}

	userID, errToken := dependencies.UserFromToken(ctx, token)
	if errToken != nil {
		return &LogoutError{Code: LogoutErrorCodeTokenLookup, Err: errToken}
	}

	if userID == "" {
		// Token is valid but not associated with a user; treat as success.
		return nil
	}

	if dependencies.LogoutUser == nil {
		return &LogoutError{Code: LogoutErrorCodeUserLogout, UserID: userID}
	}

	if errLogout := dependencies.LogoutUser(ctx, userID); errLogout != nil {
		return &LogoutError{Code: LogoutErrorCodeUserLogout, Err: errLogout, UserID: userID}
	}

	return nil
}
