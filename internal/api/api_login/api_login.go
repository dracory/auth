package api_login

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/auth/internal/core"
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

// LoginPasswordlessDeps defines the dependencies required for the passwordless login flow.
// It is intentionally decoupled from the authImplementation type to avoid import cycles.
type LoginPasswordlessDeps struct {
	DisableRateLimit bool

	// EmailSendThrottle configures the per-email send quota; zero value
	// applies the package defaults.
	EmailSendThrottle utils.EmailSendThrottle

	TemporaryKeyGet func(key string) (string, error)
	TemporaryKeySet func(key string, value string, expiresSeconds int) error

	ExpiresSeconds int

	EmailTemplate func(ctx context.Context, email string, verificationCode string) string
	EmailSend     func(ctx context.Context, email string, subject string, body string) error

	// EmailEnqueue, when set, is preferred over EmailSend: only an opaque
	// delivery key enters the queue (see utils.DeliverSecret).
	EmailEnqueue func(ctx context.Context, deliveryKey string) error

	// EmailCaptured is an optional lead-capture hook fired after input
	// validation, with verified=false (email submitted, not yet proven).
	EmailCaptured func(ctx context.Context, email string, verified bool, options types.UserAuthOptions) error
}

// ApiLogin is the HTTP-level handler that combines passwordless and
// username+password login flows behind a shared interface.
func ApiLogin(w http.ResponseWriter, r *http.Request, dependencies Dependencies) {
	logger := dependencies.Logger
	if logger == nil {
		logger = slog.Default()
	}

	if dependencies.Passwordless {
		result, err := loginPasswordless(r.Context(), r, dependencies.PasswordlessDependencies)
		if err != nil {
			// Throttle/validation errors carry no wrapped Err — fall back to
			// the message so every failure is observable, not just internal ones.
			errText := err.Error()
			if err.Err != nil {
				errText = err.Err.Error()
			}
			logger.Error("passwordless login failed",
				slog.String("error", errText),
				slog.String("code", string(err.Code)),
			)
			switch err.Code {
			case LoginPasswordlessErrorCodeValidation:
				api.Respond(w, r, api.Error(err.Message))
				return
			case LoginPasswordlessErrorCodeTooManySends:
				api.Respond(w, r, api.Error(err.Message))
				return
			case LoginPasswordlessErrorCodeTokenStore:
				api.Respond(w, r, api.Error(types.MsgFailedToProcess))
				return
			case LoginPasswordlessErrorCodeEmailSend:
				api.Respond(w, r, api.Error(types.MsgFailedToSendEmail))
				return
			default:
				api.Respond(w, r, api.Error(types.MsgInternalServer))
				return
			}
		}

		api.Respond(w, r, api.SuccessWithData(result.SuccessMessage, map[string]any{
			"nonce": result.Nonce,
		}))
		return
	}

	email := req.GetStringTrimmed(r, "email")
	password := req.GetStringTrimmed(r, "password")
	ip := req.GetIP(r)
	userAgent := r.UserAgent()

	// Post-auth pipeline preferred: the shared hooks own session create,
	// auth cookie, remember-me, and the redirect — one cookie/session
	// writer for every login method, same as the OTP verify path.
	if dependencies.LoginWithUsernameAndPasswordPostAuth != nil {
		successMessage, token, redirect, errMessage := dependencies.LoginWithUsernameAndPasswordPostAuth(
			r.Context(), w, r, email, password, ip, userAgent)
		if errMessage != "" {
			api.Respond(w, r, api.Error(errMessage))
			return
		}
		api.Respond(w, r, api.SuccessWithData(successMessage, map[string]any{
			"token":    token,
			"redirect": redirect,
		}))
		return
	}

	if dependencies.LoginWithUsernameAndPassword == nil {
		api.Respond(w, r, api.Error(types.MsgInternalServer))
		return
	}

	successMessage, token, errMessage := dependencies.LoginWithUsernameAndPassword(r.Context(), email, password, ip, userAgent)
	if errMessage != "" {
		api.Respond(w, r, api.Error(errMessage))
		return
	}

	if dependencies.UseCookies && dependencies.SetAuthCookie != nil {
		dependencies.SetAuthCookie(w, r, token)
	}

	api.Respond(w, r, api.SuccessWithData(successMessage, map[string]any{
		"token": token,
	}))
}

// ApiLoginWithAuth is a convenience wrapper that allows callers to pass a
// types.AuthSharedInterface (such as authImplementation) instead of manually
// wiring Dependencies. It constructs the Dependencies struct using the
// interface accessors and preserves the existing behaviour.
func ApiLoginWithAuth(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) {
	a = types.UnwrapAuth(a)
	passwordAuth, ok := a.(types.AuthPasswordInterface)
	if !ok {
		if logger := a.GetLogger(); logger != nil {
			logger.Error("login requires AuthPasswordInterface")
		}
		http.Error(w, types.MsgInternalServer, http.StatusInternalServerError)
		return
	}

	// Optional-hook resolution via type assertion — adding GetFunc* methods
	// to AuthSharedInterface would break external implementers, so new
	// hooks ship as optional interfaces instead (same pattern as
	// FuncTemporaryKeyDelete, FuncUserExists, FuncSessionCreate).
	var emailEnqueue func(ctx context.Context, deliveryKey string) error
	if enq, ok := a.(interface {
		GetFuncEmailEnqueue() func(ctx context.Context, deliveryKey string) error
	}); ok {
		emailEnqueue = enq.GetFuncEmailEnqueue()
	}

	var emailCaptured func(ctx context.Context, email string, verified bool, options types.UserAuthOptions) error
	if capt, ok := a.(interface {
		GetFuncEmailCaptured() func(ctx context.Context, email string, verified bool, options types.UserAuthOptions) error
	}); ok {
		if fn := capt.GetFuncEmailCaptured(); fn != nil {
			emailCaptured = func(ctx context.Context, email string, verified bool, options types.UserAuthOptions) error {
				if err := fn(ctx, email, verified, options); err != nil {
					a.GetLogger().Error("email capture hook failed", slog.String("error", err.Error()))
					return err
				}
				return nil
			}
		}
	}

	deps := Dependencies{
		Passwordless: a.IsPasswordless(),
		Logger:       a.GetLogger(),
		PasswordlessDependencies: LoginPasswordlessDeps{
			DisableRateLimit:  a.GetDisableRateLimit(),
			EmailSendThrottle: utils.EmailSendThrottleFromAuth(a),
			TemporaryKeyGet:   a.GetFuncTemporaryKeyGet(),
			TemporaryKeySet:   a.GetFuncTemporaryKeySet(),
			EmailEnqueue:      emailEnqueue,
			EmailCaptured:     emailCaptured,
			ExpiresSeconds:    0, // let business logic apply default
			EmailTemplate: func(ctx context.Context, email string, verificationCode string) string {
				fn := a.GetPasswordlessFuncEmailTemplateLoginCode()
				if fn == nil {
					return ""
				}
				return fn(ctx, email, verificationCode, types.UserAuthOptions{
					UserIp:    req.GetIP(r),
					UserAgent: r.UserAgent(),
				})
			},
			EmailSend: func(ctx context.Context, email string, subject string, body string) error {
				fn := a.GetPasswordlessFuncEmailSend()
				if fn == nil {
					return nil
				}
				return fn(ctx, email, subject, body)
			},
		},
		LoginWithUsernameAndPassword: func(ctx context.Context, email, password, ip, userAgent string) (string, string, string) {
			res := core.LoginWithUsernameAndPassword(ctx, passwordAuth, email, password, types.UserAuthOptions{
				UserIp:    ip,
				UserAgent: userAgent,
			})
			return res.SuccessMessage, res.Token, res.ErrorMessage
		},
		UseCookies: a.GetUseCookies(),
		SetAuthCookie: func(w http.ResponseWriter, r *http.Request, token string) {
			a.SetAuthCookie(w, r, token)
		},
	}

	// When ANY PostAuthHook is configured, route password login through the
	// shared post-auth pipeline so session create, auth cookie, remember-me
	// and redirect behave identically to the OTP verify path. Same
	// any-hook trigger as authenticateViaUsername; the getter ships as an
	// optional interface (adding it to AuthSharedInterface would break
	// external implementers). IssueRememberSession is intentionally not a
	// trigger — alone it produces a tokenless session, and it only makes
	// sense once SessionCreate is present.
	var postAuthHooks types.PostAuthHooks
	if g, ok := a.(interface{ GetPostAuthHooks() types.PostAuthHooks }); ok {
		postAuthHooks = g.GetPostAuthHooks()
	}
	if hooks := postAuthHooks; hooks.EmailAllowed != nil ||
		hooks.UserFindByEmailOrCreate != nil || hooks.UserIsActive != nil ||
		hooks.SessionCreate != nil || hooks.SetAuthCookie != nil ||
		hooks.RedirectURL != nil || hooks.RegisterRules != nil ||
		hooks.RegistrationComplete != nil {
		deps.LoginWithUsernameAndPasswordPostAuth = func(
			ctx context.Context,
			w http.ResponseWriter, r *http.Request,
			email, password, ip, userAgent string,
		) (string, string, string, string) {
			obsHooks := a.GetObservabilityHooks()

			// Allowlist first — cheap gate that should not wait behind a
			// bcrypt compare, matching where it sits in the OTP pipeline.
			if postAuthHooks.EmailAllowed != nil {
				if allowed, denyMsg := postAuthHooks.EmailAllowed(ctx, email); !allowed {
					if denyMsg == "" {
						denyMsg = types.MsgInvalidCredentials
					}
					return "", "", "", denyMsg
				}
			}

			userID, errMessage := core.VerifyUsernameAndPassword(ctx, passwordAuth, email, password, types.UserAuthOptions{
				UserIp:    ip,
				UserAgent: userAgent,
			})
			if errMessage != "" {
				return "", "", "", errMessage
			}

			// Pin user resolution to the just-verified account: credential
			// login must never create users, so find-or-create is bypassed.
			// EmailAllowed is cleared too — it already ran above (before the
			// password check) so the pipeline would only re-run it.
			pwHooks := postAuthHooks
			pwHooks.EmailAllowed = nil
			pwHooks.UserFindByEmailOrCreate = func(ctx context.Context, _ string, _ types.UserAuthOptions) (string, error) {
				return userID, nil
			}

			// Honor UseCookies=false (token-only API hosts): suppress the
			// auth-cookie and remember-cookie writes entirely rather than
			// letting the pipeline emit cookies the host opted out of.
			if !a.GetUseCookies() {
				pwHooks.SetAuthCookie = nil
				pwHooks.IssueRememberSession = nil
			}

			paDeps := core.PostAuthDeps{
				Hooks:    pwHooks,
				Endpoint: a.GetEndpoint(),
				Options:  types.UserAuthOptions{UserIp: ip, UserAgent: userAgent},
				Logger:   a.GetLogger(),
				AuthCookieSet: func(w http.ResponseWriter, r *http.Request, token string) {
					if a.GetUseCookies() {
						a.SetAuthCookie(w, r, token)
					}
				},
			}
			if g, ok := a.(interface{ GetRememberMeEnabled() bool }); ok {
				paDeps.RememberMeEnabled = g.GetRememberMeEnabled()
			}
			if g, ok := a.(interface{ GetRememberMeDays() int }); ok {
				paDeps.RememberMeDays = g.GetRememberMeDays()
			}
			if g, ok := a.(interface{ GetRememberCookieName() string }); ok {
				paDeps.RememberCookieName = g.GetRememberCookieName()
			}
			if g, ok := a.(interface{ GetSecureCookies() bool }); ok {
				paDeps.SecureCookies = g.GetSecureCookies()
			}

			result, err := core.PostAuth(ctx, w, r, email, "", paDeps)
			if err != nil {
				// Legacy recorded session-create failures as failed attempts;
				// keep the symmetry for inactive/denied/pipeline errors too.
				obsHooks.RecordLoginAttempt(types.LoginMethodPassword, false, err)
				return "", "", "", err.Error()
			}

			obsHooks.RecordLoginAttempt(types.LoginMethodPassword, true, nil)
			if result.Session != nil {
				obsHooks.RecordSessionCreated(userID)
			}

			return types.MsgLoginSuccess, result.Token, result.Redirect, ""
		}
	}

	ApiLogin(w, r, deps)
}
