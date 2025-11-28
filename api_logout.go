package auth

import (
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/req"
)

func (a authImplementation) apiLogout(w http.ResponseWriter, r *http.Request) {
	authToken := AuthTokenRetrieve(r, a.useCookies)

	if authToken == "" {
		api.Respond(w, r, api.Success("logout success"))
	}

	userID, errToken := a.funcUserFindByAuthToken(r.Context(), authToken, UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	if errToken != nil {
		authErr := NewLogoutError(errToken)
		logger := a.GetLogger()
		logger.Error("logout token lookup failed",
			"error", authErr.InternalErr,
			"error_code", authErr.Code,
			"ip", req.GetIP(r),
			"user_agent", r.UserAgent(),
			"endpoint", "api_logout",
		)
		api.Respond(w, r, api.Error(authErr.Message))
		return
	}

	if userID != "" {
		errLogout := a.funcUserLogout(r.Context(), userID, UserAuthOptions{
			UserIp:    req.GetIP(r),
			UserAgent: r.UserAgent(),
		})

		if errLogout != nil {
			authErr := NewLogoutError(errLogout)
			logger := a.GetLogger()
			logger.Error("user logout failed",
				"error", authErr.InternalErr,
				"error_code", authErr.Code,
				"user_id", userID,
				"ip", req.GetIP(r),
				"user_agent", r.UserAgent(),
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
