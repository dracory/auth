package auth

import (
	"context"
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/req"
)

// ApiAuthOrErrorMiddleware checks that an authentication token
// exists, and then finds the userID based on it. On success appends
// the user ID to the context. On failure it will return an
// unauthenticated JSON response.
func (a authImplementation) ApiAuthOrErrorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authToken := AuthTokenRetrieve(r, a.useCookies)

		if authToken == "" {
			api.Respond(w, r, api.Unauthenticated("auth token is required"))
			return
		}

		userID, err := a.funcUserFindByAuthToken(r.Context(), authToken, UserAuthOptions{
			UserIp:    req.GetIP(r),
			UserAgent: r.UserAgent(),
		})

		if err != nil {
			api.Respond(w, r, api.Unauthenticated("auth token is required"))
			return
		}

		if userID == "" {
			api.Respond(w, r, api.Unauthenticated("user id is required"))
			return
		}

		ctx := context.WithValue(r.Context(), AuthenticatedUserID{}, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
