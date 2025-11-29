package api_logout

import (
	"context"
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

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

// ApiLogout is the HTTP-level helper that wires request/response handling
// to the core ApiLogout business logic using the provided dependencies.
func ApiLogout(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	if deps.AuthTokenRetrieve == nil {
		api.Respond(w, r, api.Error("Internal server error. Please try again later"))
		return
	}

	token := deps.AuthTokenRetrieve(r, deps.UseCookies)
	logoutErr := logout(r.Context(), token, deps)
	if logoutErr != nil {
		switch logoutErr.Code {
		case LogoutErrorCodeTokenLookup,
			LogoutErrorCodeUserLogout:
			api.Respond(w, r, api.Error("Logout failed. Please try again later"))
			return
		default:
			api.Respond(w, r, api.Error("Internal server error. Please try again later"))
			return
		}
	}

	if deps.UseCookies && deps.RemoveAuthCookie != nil {
		deps.RemoveAuthCookie(w, r)
	}

	api.Respond(w, r, api.Success("logout success"))
}

// ApiLogoutWithAuth is a convenience wrapper that allows callers to pass a
// types.AuthSharedInterface (such as authImplementation) instead of manually
// wiring Dependencies. It constructs the Dependencies struct using the
// interface accessors and preserves the existing behaviour.
func ApiLogoutWithAuth(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
	deps := Dependencies{
		UseCookies: a.GetUseCookies(),
		AuthTokenRetrieve: func(r *http.Request, useCookies bool) string {
			return utils.AuthTokenRetrieve(r, useCookies)
		},
		RemoveAuthCookie: func(w http.ResponseWriter, r *http.Request) {
			a.RemoveAuthCookie(w, r)
		},
	}

	if fn := a.GetFuncUserFindByAuthToken(); fn != nil {
		deps.UserFromToken = func(ctx context.Context, token string) (string, error) {
			return fn(ctx, token, types.UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		}
	}

	if fn := a.GetFuncUserLogout(); fn != nil {
		deps.LogoutUser = func(ctx context.Context, userID string) error {
			return fn(ctx, userID, types.UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		}
	}

	ApiLogout(w, r, deps)
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
func logout(ctx context.Context, token string, dependencies Dependencies) *LogoutError {
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
