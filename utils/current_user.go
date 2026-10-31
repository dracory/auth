package utils

import (
	"net/http"

	"github.com/dracory/auth/types"
	"github.com/dracory/req"
)

// ResolveCurrentUserID returns the authenticated user ID for the request.
// It first honours the middleware-provided context value (when the host
// wraps the route with WebAppendUserIdIfExistsMiddleware), then falls
// back to resolving the auth cookie through FuncUserFindByAuthToken so
// pages served by the package's own mux work without host middleware.
// Returns "" when the request is unauthenticated.
func ResolveCurrentUserID(r *http.Request, a types.AuthSharedInterface) string {
	if userID := a.GetCurrentUserID(r); userID != "" {
		return userID
	}

	token := AuthTokenRetrieveWithName(r, a.GetUseCookies(), a.GetCookieName())
	if token == "" {
		return ""
	}

	fn := a.GetFuncUserFindByAuthToken()
	if fn == nil {
		return ""
	}

	userID, err := fn(r.Context(), token, types.UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		return ""
	}
	return userID
}
