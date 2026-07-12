package api_impersonate_stop

import (
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

func ApiImpersonateStop(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	currentAuthToken := utils.AuthTokenRetrieve(r, deps.UseCookies)
	if currentAuthToken == "" {
		api.Respond(w, r, api.Error(types.MsgNotImpersonating))
		return
	}

	originalAdminToken, err := deps.TemporaryKeyGet(types.ImpersonationKeyPrefix + currentAuthToken)
	if err != nil {
		api.Respond(w, r, api.Error(types.MsgImpersonationFailed))
		return
	}
	if originalAdminToken == "" {
		api.Respond(w, r, api.Error(types.MsgNotImpersonating))
		return
	}

	options := types.UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	}

	targetUserID, err := deps.UserFindByAuthToken(r.Context(), currentAuthToken, options)
	if err != nil || targetUserID == "" {
		api.Respond(w, r, api.Error(types.MsgImpersonationFailed))
		return
	}

	adminUserID, err := deps.UserFindByAuthToken(r.Context(), originalAdminToken, options)
	if err != nil || adminUserID == "" {
		api.Respond(w, r, api.Error(types.MsgImpersonationFailed))
		return
	}

	if deps.UserLogout != nil {
		_ = deps.UserLogout(r.Context(), targetUserID, options)
	}

	if deps.UseCookies && deps.SetAuthCookie != nil {
		deps.SetAuthCookie(w, r, originalAdminToken)
	}

	_ = deps.TemporaryKeySet(types.ImpersonationKeyPrefix+currentAuthToken, "", 1)

	if deps.ImpersonationStop != nil {
		_ = deps.ImpersonationStop(r.Context(), adminUserID, targetUserID)
	}

	if deps.ObservabilityHooks != nil {
		deps.ObservabilityHooks.RecordImpersonationStop(adminUserID, targetUserID)
	}

	api.Respond(w, r, api.Success(types.MsgImpersonationStopped))
}

func ApiImpersonateStopWithAuth(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
	deps := Dependencies{
		UserFindByAuthToken: a.GetFuncUserFindByAuthToken(),
		UserLogout:          a.GetFuncUserLogout(),
		TemporaryKeyGet:     a.GetFuncTemporaryKeyGet(),
		TemporaryKeySet:     a.GetFuncTemporaryKeySet(),
		UseCookies:          a.GetUseCookies(),
		ObservabilityHooks:  a.GetObservabilityHooks(),
		ImpersonationStop:   a.GetFuncImpersonationStop(),
		SetAuthCookie: func(w http.ResponseWriter, r *http.Request, token string) {
			a.SetAuthCookie(w, r, token)
		},
	}

	ApiImpersonateStop(w, r, deps)
}
