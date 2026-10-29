package core

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

// PostAuthDeps wires the post-auth pipeline. AuthCookieSet is the module's
// default cookie writer when hooks.SetAuthCookie is nil.
type PostAuthDeps struct {
	Hooks         types.PostAuthHooks
	Options       types.UserAuthOptions
	Logger        *slog.Logger
	AuthCookieSet func(w http.ResponseWriter, r *http.Request, token string)

	// Remember-me: when enabled and the request opted in, the pipeline
	// issues a marked long-lived session via hooks.IssueRememberSession and
	// sets the remember cookie.
	RememberMeEnabled  bool
	RememberMeDays     int
	RememberCookieName string
	SecureCookies      bool
}

// PostAuthResult is what the caller hands back to the client.
type PostAuthResult struct {
	UserID   string
	Token    string // session key or auth token — for API/cookie use
	Redirect string
	Session  types.SessionInterface
}

// PostAuth runs the shared post-authentication pipeline every login method
// funnels through: allowlist → find-or-create user → active check →
// session create → auth cookie → safe redirect. All hooks optional.
func PostAuth(ctx context.Context, w http.ResponseWriter, r *http.Request, email string, backURL string, deps PostAuthDeps) (*PostAuthResult, error) {
	hooks := deps.Hooks
	options := deps.Options
	if options.UserIp == "" {
		options.UserIp = req.GetIP(r)
	}
	if options.UserAgent == "" {
		options.UserAgent = r.UserAgent()
	}

	// 1. Allowlist — must run here (middleware may not cover auth routes).
	if hooks.EmailAllowed != nil {
		if allowed, denyMsg := hooks.EmailAllowed(ctx, email); !allowed {
			if denyMsg == "" {
				denyMsg = types.MsgInvalidCredentials
			}
			return nil, errors.New(denyMsg)
		}
	}

	// 2. Find-or-create user.
	var userID string
	if hooks.UserFindByEmailOrCreate != nil {
		uid, err := hooks.UserFindByEmailOrCreate(ctx, email, options)
		if err != nil {
			if deps.Logger != nil {
				deps.Logger.Error("post-auth: user find-or-create failed",
					slog.String("error", err.Error()),
					slog.String("email", email))
			}
			return nil, errors.New(types.MsgFailedToProcess)
		}
		userID = uid
	}

	// 3. Arbitrary registration rules — first failure short-circuits.
	if hooks.RegisterRules != nil {
		for _, rule := range hooks.RegisterRules(ctx, email) {
			if rule != nil && !rule.Passes() {
				return nil, errors.New(rule.FailMessage())
			}
		}
	}

	// 4. Active check.
	if hooks.UserIsActive != nil && userID != "" {
		if !hooks.UserIsActive(ctx, userID, options) {
			return nil, errors.New(types.MsgUserInactive)
		}
	}

	// 5. Session create → credential.
	var token string
	var session types.SessionInterface
	if hooks.SessionCreate != nil && userID != "" {
		s, err := hooks.SessionCreate(ctx, userID, options)
		if err != nil || s == nil || s.GetKey() == "" {
			return nil, errors.New(types.MsgFailedToProcess)
		}
		session = s
		token = s.GetKey()
	}

	// 6. Auth cookie.
	cookieSetter := deps.AuthCookieSet
	if hooks.SetAuthCookie != nil {
		cookieSetter = hooks.SetAuthCookie
	}
	if cookieSetter != nil && token != "" {
		cookieSetter(w, r, token)
	}

	// 6b. Remember-me: marked long-lived session + second cookie, only when
	// enabled AND the user opted in (`remember=true`).
	// Note: opt-in reads the form field — JSON-body login endpoints can't
	// trigger it unless callers also forward the field.
	if deps.RememberMeEnabled && hooks.IssueRememberSession != nil && userID != "" &&
		(r.FormValue("remember") == "true" || r.FormValue("remember") == "on") {
		rs, rErr := hooks.IssueRememberSession(ctx, userID, options)
		if rErr == nil && rs != nil && rs.GetKey() != "" {
			days := deps.RememberMeDays
			if days <= 0 {
				days = 30
			}
			name := deps.RememberCookieName
			if name == "" {
				name = utils.DefaultRememberCookieName
			}
			utils.RememberCookieSet(w, name, rs.GetKey(), days, deps.SecureCookies)
		} else if deps.Logger != nil {
			deps.Logger.Error("post-auth: remember session issue failed",
				slog.Any("error", rErr))
		}
	}

	// 7. Redirect: safe backURL first, else hook, else root.
	redirect := "/"
	if utils.IsSafeBackURL(backURL, r.Host) {
		redirect = backURL
	} else if hooks.RedirectURL != nil {
		if u := hooks.RedirectURL(ctx, userID); u != "" {
			redirect = u
		}
	}

	return &PostAuthResult{UserID: userID, Token: token, Redirect: redirect, Session: session}, nil
}
