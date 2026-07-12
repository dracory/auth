package api_impersonate_start

import (
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/req"
	"github.com/dracory/str"
)

const tokenGamma = "BCDFGHJKLMNPQRSTVXYZ"
const tokenLength = 32

func ApiImpersonateStart(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	authToken := utils.AuthTokenRetrieve(r, deps.UseCookies)
	if authToken == "" {
		api.Respond(w, r, api.Unauthenticated(types.MsgTokenRequired))
		return
	}

	options := types.UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	}

	adminUserID, err := deps.UserFindByAuthToken(r.Context(), authToken, options)
	if err != nil || adminUserID == "" {
		api.Respond(w, r, api.Unauthenticated(types.MsgInvalidCredentials))
		return
	}

	targetUserID := req.GetStringTrimmed(r, "user_id")
	if targetUserID == "" {
		api.Respond(w, r, api.Error(types.MsgTokenRequired))
		return
	}

	if deps.TemporaryKeyGet != nil {
		existing, _ := deps.TemporaryKeyGet(impersonationKeyPrefix + authToken)
		if existing != "" {
			api.Respond(w, r, api.Error(types.MsgAlreadyImpersonating))
			return
		}
	}

	allowed, err := deps.CanImpersonate(r.Context(), adminUserID, targetUserID)
	if err != nil {
		api.Respond(w, r, api.Error(types.MsgImpersonationFailed))
		return
	}
	if !allowed {
		api.Respond(w, r, api.Error(types.MsgImpersonationForbidden))
		return
	}

	newToken, err := str.RandomFromGamma(tokenLength, tokenGamma)
	if err != nil {
		api.Respond(w, r, api.Error(types.MsgFailedToGenerateCode))
		return
	}

	if deps.TemporaryKeySet != nil {
		err = deps.TemporaryKeySet(
			impersonationKeyPrefix+newToken,
			authToken,
			int(DefaultAuthTokenExpiration.Seconds()),
		)
		if err != nil {
			api.Respond(w, r, api.Error(types.MsgImpersonationFailed))
			return
		}
	}

	if deps.UserStoreAuthToken != nil {
		err = deps.UserStoreAuthToken(r.Context(), newToken, targetUserID, options)
		if err != nil {
			api.Respond(w, r, api.Error(types.MsgImpersonationFailed))
			return
		}
	}

	if deps.UseCookies && deps.SetAuthCookie != nil {
		deps.SetAuthCookie(w, r, newToken)
	}

	if deps.ImpersonationStart != nil {
		_ = deps.ImpersonationStart(r.Context(), adminUserID, targetUserID)
	}

	if deps.ObservabilityHooks != nil {
		deps.ObservabilityHooks.RecordImpersonationStart(adminUserID, targetUserID)
		deps.ObservabilityHooks.RecordSessionCreated(targetUserID)
	}

	api.Respond(w, r, api.SuccessWithData(types.MsgImpersonationStarted, map[string]any{
		"token": newToken,
	}))
}

func ApiImpersonateStartWithAuth(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
	deps := Dependencies{
		CanImpersonate:      a.GetFuncCanImpersonate(),
		UserFindByAuthToken: a.GetFuncUserFindByAuthToken(),
		UserStoreAuthToken:  a.GetFuncUserStoreAuthToken(),
		TemporaryKeyGet:     a.GetFuncTemporaryKeyGet(),
		TemporaryKeySet:     a.GetFuncTemporaryKeySet(),
		UseCookies:          a.GetUseCookies(),
		ObservabilityHooks:  a.GetObservabilityHooks(),
		ImpersonationStart:  a.GetFuncImpersonationStart(),
		SetAuthCookie: func(w http.ResponseWriter, r *http.Request, token string) {
			a.SetAuthCookie(w, r, token)
		},
	}

	ApiImpersonateStart(w, r, deps)
}
