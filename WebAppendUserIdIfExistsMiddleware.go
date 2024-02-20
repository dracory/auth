package auth

import (
	"context"
	"net/http"

	"github.com/gouniverse/utils"
)

// WebAppendUserIdIfExistsMiddleware appends the user ID to the context
// if an authentication token exists in the requests. This middleware does
// not have a side effect like for instance redirecting to the login
// endpoint. This is why it is important to be added to places which
// can be used by both guests and users (i.e. website pages), where authenticated
// users may have some extra privileges
//
// If you need to redirect the user if authentication token not found,
// or the user does not exist, take a look at the WebAuthOrRedirectMiddleware
// middleware, which does exactly that
func (a Auth) WebAppendUserIdIfExistsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authToken := AuthTokenRetrieve(r, a.useCookies)

		if authToken != "" {
			userID, err := a.funcUserFindByAuthToken(authToken, utils.IP(r), r.UserAgent())

			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), AuthenticatedUserID{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		next.ServeHTTP(w, r)
	})
}
