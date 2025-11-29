package api_authenticate_via_username

import (
	"context"
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/auth/types"
	"github.com/dracory/req"
	"github.com/dracory/str"
)

// Dependencies defines the dependencies required for authenticating a user by
// username (or email) and returning an auth token.
type Dependencies struct {
	Passwordless bool

	PasswordlessUserFindByEmail func(ctx context.Context, email string) (string, error)
	UserFindByUsername          func(ctx context.Context, username, firstName, lastName string) (string, error)

	UserStoreAuthToken func(ctx context.Context, token, userID string) error

	UseCookies    bool
	SetAuthCookie func(w http.ResponseWriter, r *http.Request, token string)
}

// AuthenticateErrorCode categorizes error sources in the authentication flow.
type AuthenticateErrorCode string

const (
	AuthenticateErrorCodeNone       AuthenticateErrorCode = ""
	AuthenticateErrorCodeUserLookup AuthenticateErrorCode = "user_lookup"
	AuthenticateErrorCodeCodeGen    AuthenticateErrorCode = "code_generation"
	AuthenticateErrorCodeTokenStore AuthenticateErrorCode = "token_store"
)

// AuthenticateError represents a structured error in the authentication flow.
type AuthenticateError struct {
	Code    AuthenticateErrorCode
	Message string
	Err     error
}

func (e *AuthenticateError) Error() string {
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

// AuthenticateResult represents a successful authentication.
type AuthenticateResult struct {
	Token string
}

// ApiAuthenticateViaUsername is the HTTP-level helper that wires
// request/response handling to the core AuthenticateViaUsername business
// logic using the provided dependencies.
func ApiAuthenticateViaUsername(w http.ResponseWriter, r *http.Request, username, firstName, lastName string, deps Dependencies) {
	result, aerr := AuthenticateViaUsername(r.Context(), username, firstName, lastName, deps)
	if aerr != nil {
		// All errors map directly to their user-facing messages.
		api.Respond(w, r, api.Error(aerr.Message))
		return
	}

	if deps.UseCookies && deps.SetAuthCookie != nil {
		deps.SetAuthCookie(w, r, result.Token)
	}

	api.Respond(w, r, api.SuccessWithData("login success", map[string]any{
		"token": result.Token,
	}))
}

// ApiAuthenticateViaUsernameWithAuth is a convenience wrapper that allows
// callers to pass a types.AuthSharedInterface (such as authImplementation)
// instead of manually wiring Dependencies. It constructs the Dependencies
// struct using the interface accessors and preserves the existing behaviour.
func ApiAuthenticateViaUsernameWithAuth(w http.ResponseWriter, r *http.Request, username, firstName, lastName string, a types.AuthSharedInterface) {
	deps := Dependencies{
		Passwordless: a.IsPasswordless(),
		UseCookies:   a.GetUseCookies(),
	}

	if fn := a.GetPasswordlessUserFindByEmail(); fn != nil {
		deps.PasswordlessUserFindByEmail = func(ctx context.Context, email string) (string, error) {
			return fn(ctx, email, types.UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		}
	}

	if fn := a.GetFuncUserFindByUsername(); fn != nil {
		deps.UserFindByUsername = func(ctx context.Context, username, firstName, lastName string) (string, error) {
			return fn(ctx, username, firstName, lastName, types.UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		}
	}

	if fn := a.GetFuncUserStoreAuthToken(); fn != nil {
		deps.UserStoreAuthToken = func(ctx context.Context, token, userID string) error {
			return fn(ctx, token, userID, types.UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		}
	}

	deps.SetAuthCookie = func(w http.ResponseWriter, r *http.Request, token string) {
		a.SetAuthCookie(w, r, token)
	}

	ApiAuthenticateViaUsername(w, r, username, firstName, lastName, deps)
}

// AuthenticateViaUsername contains the core business logic for authenticating
// a user given a username (or email) and optional first/last name. It does not
// log or write HTTP responses.
func AuthenticateViaUsername(ctx context.Context, username, firstName, lastName string, deps Dependencies) (*AuthenticateResult, *AuthenticateError) {
	var userID string
	var errUser error

	if deps.Passwordless {
		if deps.PasswordlessUserFindByEmail == nil {
			return nil, &AuthenticateError{
				Code:    AuthenticateErrorCodeUserLookup,
				Message: "Invalid credentials",
			}
		}
		userID, errUser = deps.PasswordlessUserFindByEmail(ctx, username)
	} else {
		if deps.UserFindByUsername == nil {
			return nil, &AuthenticateError{
				Code:    AuthenticateErrorCodeUserLookup,
				Message: "Invalid credentials",
			}
		}
		userID, errUser = deps.UserFindByUsername(ctx, username, firstName, lastName)
	}

	if errUser != nil {
		return nil, &AuthenticateError{
			Code:    AuthenticateErrorCodeUserLookup,
			Message: "Invalid credentials",
			Err:     errUser,
		}
	}

	if userID == "" {
		return nil, &AuthenticateError{
			Code:    AuthenticateErrorCodeUserLookup,
			Message: "Invalid credentials",
		}
	}

	token, errRandomFromGamma := str.RandomFromGamma(32, "BCDFGHJKLMNPQRSTVXYZ")
	if errRandomFromGamma != nil {
		return nil, &AuthenticateError{
			Code:    AuthenticateErrorCodeCodeGen,
			Message: "Failed to generate verification code. Please try again later",
			Err:     errRandomFromGamma,
		}
	}

	if deps.UserStoreAuthToken == nil {
		return nil, &AuthenticateError{
			Code:    AuthenticateErrorCodeTokenStore,
			Message: "Failed to process request. Please try again later",
		}
	}

	if errSession := deps.UserStoreAuthToken(ctx, token, userID); errSession != nil {
		return nil, &AuthenticateError{
			Code:    AuthenticateErrorCodeTokenStore,
			Message: "Failed to process request. Please try again later",
			Err:     errSession,
		}
	}

	return &AuthenticateResult{Token: token}, nil
}
