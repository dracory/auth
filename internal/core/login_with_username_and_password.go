package core

import (
	"context"

	"github.com/dracory/auth/internal/rules"
	"github.com/dracory/auth/types"
	authutils "github.com/dracory/auth/utils"
	"github.com/dracory/str"
)

type LoginWithUsernameAndPasswordResult struct {
	ErrorMessage   string
	SuccessMessage string
	Token          string
}

func LoginWithUsernameAndPassword(
	ctx context.Context,
	a types.AuthPasswordInterface,
	email string,
	password string,
	options types.UserAuthOptions,
) LoginWithUsernameAndPasswordResult {
	var response LoginWithUsernameAndPasswordResult

	if email == "" {
		response.ErrorMessage = types.MsgEmailRequired
		return response
	}

	if password == "" {
		response.ErrorMessage = types.MsgPasswordRequired
		return response
	}

	if check := rules.NewEmailFormatRule(email); check.Fails() {
		response.ErrorMessage = check.FailMessageFirst()
		return response
	}

	loginFn := a.GetFuncUserLogin()
	storeFn := a.GetFuncUserStoreAuthToken()
	logger := a.GetLogger()
	hooks := a.GetObservabilityHooks()

	// Anti-enumeration: when the optional FuncUserExists hook is configured,
	// check existence first so an unknown email still pays the cost of a
	// bcrypt compare — same wall time as a wrong password — rather than
	// short-circuiting on a missing user.
	if checker, ok := a.(interface {
		GetFuncUserExists() func(ctx context.Context, email string, options types.UserAuthOptions) (bool, error)
	}); ok {
		if existsFn := checker.GetFuncUserExists(); existsFn != nil {
			exists, existsErr := existsFn(ctx, email, options)
			if existsErr != nil && logger != nil {
				// Fail-open is deliberate (a flaky existence check must not
				// block logins) — but it must not be invisible.
				logger.Warn("user existence check failed, proceeding with login",
					"error", existsErr,
					"email", email,
				)
			}
			if existsErr == nil && !exists {
				authutils.DummyPasswordCompare(password)
				response.ErrorMessage = types.MsgInvalidCredentials
				hooks.RecordLoginAttempt(types.LoginMethodPassword, false, nil)
				return response
			}
		}
	}

	userID, err := loginFn(ctx, email, password, options)

	if err != nil {
		response.ErrorMessage = types.MsgInvalidCredentials
		if logger != nil {
			logger.Error("login with username and password failed",
				"error", err,
				"email", email,
				"ip", options.UserIp,
				"user_agent", options.UserAgent,
			)
		}
		hooks.RecordLoginAttempt(types.LoginMethodPassword, false, err)
		return response
	}

	if userID == "" {
		response.ErrorMessage = types.MsgInvalidCredentials
		hooks.RecordLoginAttempt(types.LoginMethodPassword, false, nil)
		return response
	}

	// Session-backed auth preferred: when FuncSessionCreate is configured,
	// create a real session and use its key as the credential placed in the
	// auth cookie. Otherwise fall back to the legacy token write.
	var token string
	if sessionCreator, ok := a.(interface {
		GetFuncSessionCreate() func(ctx context.Context, userID string, options types.UserAuthOptions) (types.SessionInterface, error)
	}); ok && sessionCreator.GetFuncSessionCreate() != nil {
		session, errSession := sessionCreator.GetFuncSessionCreate()(ctx, userID, options)
		// A nil session or an empty key would leave the client holding a
		// credential that authenticates nothing — treat as create failure.
		if errSession != nil || session == nil || session.GetKey() == "" {
			response.ErrorMessage = types.MsgFailedToProcess
			if logger != nil {
				logger.Error("session create failed",
					"error", errSession,
					"error_code", "SESSION_CREATE_FAILED",
					"email", email,
					"user_id", userID,
					"ip", options.UserIp,
					"user_agent", options.UserAgent,
				)
			}
			hooks.RecordLoginAttempt(types.LoginMethodPassword, false, errSession)
			return response
		}
		token = session.GetKey()
	} else {
		// Reuse the shared login/verification gamma from utils to avoid
		// duplicating the character set. We keep length at 32 for auth tokens.
		tokenGenerated, errRandom := str.RandomFromGamma(32, authutils.LoginCodeGamma(false))
		if errRandom != nil {
			response.ErrorMessage = types.MsgFailedToGenerateCode
			if logger != nil {
				logger.Error("auth token generation failed",
					"error", errRandom,
					"error_code", "CODE_GENERATION_FAILED",
					"email", email,
					"ip", options.UserIp,
					"user_agent", options.UserAgent,
				)
			}
			hooks.RecordLoginAttempt(types.LoginMethodPassword, false, errRandom)
			return response
		}
		token = tokenGenerated

		errSession := storeFn(ctx, token, userID, options)

		if errSession != nil {
			response.ErrorMessage = types.MsgFailedToProcess
			if logger != nil {
				logger.Error("auth token store failed",
					"error", errSession,
					"error_code", "TOKEN_STORE_FAILED",
					"email", email,
					"user_id", userID,
					"ip", options.UserIp,
					"user_agent", options.UserAgent,
				)
			}
			hooks.RecordLoginAttempt(types.LoginMethodPassword, false, errSession)
			return response
		}
	}

	response.SuccessMessage = types.MsgLoginSuccess
	response.Token = token
	hooks.RecordLoginAttempt(types.LoginMethodPassword, true, nil)
	hooks.RecordSessionCreated(userID)
	return response
}
