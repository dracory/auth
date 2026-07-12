package middlewares

import (
	"context"
	"net/http"

	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
)

const impersonationKeyPrefix = "imp:"

func ImpersonationAwareMiddleware(next http.Handler, a types.AuthSharedInterface) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authToken := utils.AuthTokenRetrieve(r, a.GetUseCookies())
		if authToken == "" {
			next.ServeHTTP(w, r)
			return
		}

		tempKeyGet := a.GetFuncTemporaryKeyGet()
		if tempKeyGet == nil {
			next.ServeHTTP(w, r)
			return
		}

		originalToken, err := tempKeyGet(impersonationKeyPrefix + authToken)
		if err != nil || originalToken == "" {
			next.ServeHTTP(w, r)
			return
		}

		userFindByAuthToken := a.GetFuncUserFindByAuthToken()
		if userFindByAuthToken == nil {
			next.ServeHTTP(w, r)
			return
		}

		adminUserID, err := userFindByAuthToken(r.Context(), originalToken, types.UserAuthOptions{
			UserIp:    r.RemoteAddr,
			UserAgent: r.UserAgent(),
		})
		if err != nil || adminUserID == "" {
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), types.ImpersonatorUserID{}, adminUserID)
		ctx = context.WithValue(ctx, types.IsImpersonating{}, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
