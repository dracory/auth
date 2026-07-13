package middlewares

import (
	"context"
	"net/http"

	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

func ImpersonationAwareMiddleware(next http.Handler, a types.AuthSharedInterface) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authToken := utils.AuthTokenRetrieveWithName(r, a.GetUseCookies(), a.GetCookieName())
		if authToken == "" {
			next.ServeHTTP(w, r)
			return
		}

		tempKeyGet := a.GetFuncTemporaryKeyGet()
		if tempKeyGet == nil {
			next.ServeHTTP(w, r)
			return
		}

		originalToken, err := tempKeyGet(types.ImpersonationKeyPrefix + authToken)
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
			UserIp:    req.GetIP(r),
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
