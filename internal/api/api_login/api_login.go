package api_login

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/auth/internal/core"
	"github.com/dracory/auth/types"
	"github.com/dracory/req"
)

// LoginPasswordlessDeps defines the dependencies required for the passwordless login flow.
// It is intentionally decoupled from the authImplementation type to avoid import cycles.
type LoginPasswordlessDeps struct {
	DisableRateLimit bool

	TemporaryKeyGet func(key string) (string, error)
	TemporaryKeySet func(key string, value string, expiresSeconds int) error

	ExpiresSeconds int

	EmailTemplate func(ctx context.Context, email string, verificationCode string) string
	EmailSend     func(ctx context.Context, email string, subject string, body string) error

	// EmailEnqueue, when set, is preferred over EmailSend: only an opaque
	// delivery key enters the queue (see utils.DeliverSecret).
	EmailEnqueue func(ctx context.Context, deliveryKey string) error
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
			if err.Err != nil {
				logger.Error("passwordless login failed",
					slog.String("error", err.Err.Error()),
					slog.String("code", string(err.Code)),
				)
			}
			switch err.Code {
			case LoginPasswordlessErrorCodeValidation:
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

	if dependencies.LoginWithUsernameAndPassword == nil {
		api.Respond(w, r, api.Error(types.MsgInternalServer))
		return
	}

	email := req.GetStringTrimmed(r, "email")
	password := req.GetStringTrimmed(r, "password")
	ip := req.GetIP(r)
	userAgent := r.UserAgent()

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

	deps := Dependencies{
		Passwordless: a.IsPasswordless(),
		Logger:       a.GetLogger(),
		PasswordlessDependencies: LoginPasswordlessDeps{
			DisableRateLimit: a.GetDisableRateLimit(),
			TemporaryKeyGet:  a.GetFuncTemporaryKeyGet(),
			TemporaryKeySet:  a.GetFuncTemporaryKeySet(),
			EmailEnqueue:     emailEnqueue,
			ExpiresSeconds:   0, // let business logic apply default
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

	ApiLogin(w, r, deps)
}
