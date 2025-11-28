package auth

import (
	"net/http"

	"github.com/dracory/api"
	apilogin "github.com/dracory/auth/internal/api"
	"github.com/dracory/req"
	"github.com/dracory/str"
)

// apiLoginCodeVerify used for passwordless login code verification
func (a authImplementation) apiLoginCodeVerify(w http.ResponseWriter, r *http.Request) {
	deps := apilogin.LoginCodeVerifyDeps{
		DisableRateLimit: a.disableRateLimit,
		TemporaryKeyGet:  a.funcTemporaryKeyGet,
	}

	result, perr := apilogin.LoginCodeVerify(r.Context(), r, deps)
	if perr != nil {
		// Original implementation did not log; it only returned validation/expiry
		// messages. Preserve that behaviour here.
		switch perr.Code {
		case apilogin.LoginCodeVerifyErrorCodeValidation,
			apilogin.LoginCodeVerifyErrorCodeCodeExpired:
			api.Respond(w, r, api.Error(perr.Message))
			return
		default:
			api.Respond(w, r, api.Error("Verification code has expired"))
			return
		}
	}

	// On success, perform authentication via username as before.
	a.authenticateViaUsername(w, r, result.Email, "", "")
}

// authenticateViaEmail used for passwordless login and registration
// username is an email in passwordless auth
// firstName is used only in username and password auth
// lastName is used only in username and password auth
func (a authImplementation) authenticateViaUsername(w http.ResponseWriter, r *http.Request, username string, firstName string, lastName string) {
	var userID string
	var errUser error
	if a.passwordless {
		userID, errUser = a.passwordlessFuncUserFindByEmail(r.Context(), username, UserAuthOptions{
			UserIp:    req.GetIP(r),
			UserAgent: r.UserAgent(),
		})
	} else {
		userID, errUser = a.funcUserFindByUsername(r.Context(), username, firstName, lastName, UserAuthOptions{
			UserIp:    req.GetIP(r),
			UserAgent: r.UserAgent(),
		})
	}

	if errUser != nil {
		api.Respond(w, r, api.Error("Invalid credentials"))
		return
	}

	if userID == "" {
		api.Respond(w, r, api.Error("Invalid credentials"))
		return
	}

	token, errRandomFromGamma := str.RandomFromGamma(32, "BCDFGHJKLMNPQRSTVXYZ")

	if errRandomFromGamma != nil {
		authErr := NewCodeGenerationError(errRandomFromGamma)
		logger := a.GetLogger()
		logger.Error("login auth token generation failed",
			"error", authErr.InternalErr,
			"error_code", authErr.Code,
			"user_id", userID,
			"ip", req.GetIP(r),
			"user_agent", r.UserAgent(),
			"endpoint", "authenticate_via_username",
		)
		api.Respond(w, r, api.Error(authErr.Message))
		return
	}

	errSession := a.funcUserStoreAuthToken(r.Context(), token, userID, UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	})

	if errSession != nil {
		authErr := NewTokenStoreError(errSession)
		logger := a.GetLogger()
		logger.Error("login auth token store failed",
			"error", authErr.InternalErr,
			"error_code", authErr.Code,
			"user_id", userID,
			"ip", req.GetIP(r),
			"user_agent", r.UserAgent(),
			"endpoint", "authenticate_via_username",
		)
		api.Respond(w, r, api.Error(authErr.Message))
		return
	}

	if a.useCookies {
		a.setAuthCookie(w, r, token)
	}

	api.Respond(w, r, api.SuccessWithData("login success", map[string]interface{}{
		"token": token,
	}))
}
