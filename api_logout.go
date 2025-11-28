package auth

import (
	"context"
	"net/http"

	"github.com/dracory/api"
	apilogout "github.com/dracory/auth/internal/api"
	"github.com/dracory/req"
)

func (a authImplementation) apiLogout(w http.ResponseWriter, r *http.Request) {
	authToken := AuthTokenRetrieve(r, a.useCookies)
	deps := apilogout.LogoutDeps{
		UserFromToken: func(ctx context.Context, token string) (string, error) {
			return a.funcUserFindByAuthToken(ctx, token, UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		},
		LogoutUser: func(ctx context.Context, userID string) error {
			return a.funcUserLogout(ctx, userID, UserAuthOptions{
				UserIp:    req.GetIP(r),
				UserAgent: r.UserAgent(),
			})
		},
	}

	logoutErr := apilogout.Logout(r.Context(), authToken, deps)
	if logoutErr != nil {
		logger := a.GetLogger()
		ip := req.GetIP(r)
		userAgent := r.UserAgent()
		switch logoutErr.Code {
		case apilogout.LogoutErrorCodeTokenLookup:
			authErr := NewLogoutError(logoutErr.Err)
			logger.Error("logout token lookup failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_logout",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		case apilogout.LogoutErrorCodeUserLogout:
			authErr := NewLogoutError(logoutErr.Err)
			logger.Error("user logout failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"user_id", logoutErr.UserID,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_logout",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		default:
			authErr := NewInternalError(logoutErr.Err)
			logger := a.GetLogger()
			logger.Error("logout internal error",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"ip", ip,
				"user_agent", userAgent,
				"endpoint", "api_logout",
			)
			api.Respond(w, r, api.Error(authErr.Message))
			return
		}
	}

	if a.useCookies {
		a.removeAuthCookie(w, r)
	}

	api.Respond(w, r, api.Success("logout success"))
}
