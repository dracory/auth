package middlewares

import (
	"context"
	"net/http"

	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

// WebAuthOrRedirectMiddleware checks that an authentication token
// exists, and then finds the userID based on it. On success appends
// the user ID to the context. On failure it will redirect the user
// to the login endpoint to reauthenticate.
//
// If you need to only find if the authentication token is successful
// without redirection please use the WebAppendUserIdIfExistsMiddleware
// which does exactly that without side effects
func WebAuthOrRedirectMiddleware(next http.Handler, a types.AuthSharedInterface) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authToken := utils.AuthTokenRetrieve(r, a.GetUseCookies())

		if authToken == "" {
			http.Redirect(w, r, a.LinkLogin(), http.StatusTemporaryRedirect)
			return
		}

		userID, err := a.GetFuncUserFindByAuthToken()(r.Context(), authToken, types.UserAuthOptions{
			UserIp:    req.GetIP(r),
			UserAgent: r.UserAgent(),
		})

		if err != nil {
			http.Redirect(w, r, a.LinkLogin(), http.StatusTemporaryRedirect)
			return
		}

		if userID == "" {
			http.Redirect(w, r, a.LinkLogin(), http.StatusTemporaryRedirect)
			return
		}

		ctx := context.WithValue(r.Context(), types.AuthenticatedUserID{}, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
