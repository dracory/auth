// Package middlewares: remember-me exchange middleware.
//
// If the auth cookie is absent/expired but a remember cookie holds a valid
// marked session, the middleware exchanges it: rotates the remember
// session (delete old, issue new) and mints a short-lived auth session.
// Both directions are guarded — a marked remember key in the auth cookie
// is rejected (and deleted), and an unmarked session in the remember
// cookie is rejected.
package middlewares

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

// RememberMeConfig wires the remember-me exchange middleware.
type RememberMeConfig struct {
	// SessionStore looks up and rotates sessions.
	SessionStore types.SessionStoreInterface

	// AuthCookieName is the auth session cookie name.
	AuthCookieName string
	// RememberCookieName is the remember token cookie name.
	RememberCookieName string

	// IssueRememberSession creates the rotated marked session (same hook
	// as the pipeline's PostAuthHooks.IssueRememberSession).
	IssueRememberSession func(ctx context.Context, userID string, options types.UserAuthOptions) (types.SessionInterface, error)

	// IssueAuthSession mints the short-lived auth session exchanged from
	// the remember session — the auth cookie must carry ITS key, never the
	// (now-deleted) remember key.
	IssueAuthSession func(ctx context.Context, userID string, options types.UserAuthOptions) (types.SessionInterface, error)

	// RememberMeDays is the rotated cookie/session TTL (default 30).
	RememberMeDays int
	// SecureCookies controls the Secure flag on re-issued cookies.
	SecureCookies bool
	// AuthCookieSet writes the auth cookie for the minted session.
	AuthCookieSet func(w http.ResponseWriter, r *http.Request, token string)

	// Logger logs exchange failures (rotation/mint errors are otherwise
	// silent fail-closed paths).
	Logger *slog.Logger
}

// RememberMeMiddleware wraps next with remember-me exchange logic.
func RememberMeMiddleware(cfg RememberMeConfig, next http.Handler) http.Handler {
	if cfg.RememberCookieName == "" {
		cfg.RememberCookieName = utils.DefaultRememberCookieName
	}
	if cfg.RememberMeDays <= 0 {
		cfg.RememberMeDays = 30
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cfg.SessionStore == nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx := r.Context()

		// Auth cookie present: check it's not a misused remember key.
		if c, err := r.Cookie(cfg.AuthCookieName); err == nil && c.Value != "" {
			sess, _ := cfg.SessionStore.SessionFindByKey(ctx, c.Value)
			if sess != nil && sess.GetValue(utils.RememberSessionMarker) == "" {
				next.ServeHTTP(w, r) // normal session — proceed
				return
			}
			// Misuse A: a remember key presented as an auth credential —
			// delete it so it can't mint 30-day access.
			if sess != nil {
				_ = cfg.SessionStore.SessionDeleteByKey(ctx, c.Value)
			}
		}

		// Remember cookie present?
		rc, err := r.Cookie(cfg.RememberCookieName)
		if err != nil || rc.Value == "" {
			next.ServeHTTP(w, r)
			return
		}
		rsess, err := cfg.SessionStore.SessionFindByKey(ctx, rc.Value)
		if err != nil || rsess == nil ||
			rsess.GetValue(utils.RememberSessionMarker) == "" {
			// Misuse B: unmarked or unknown session in the remember cookie.
			utils.RememberCookieRemove(w, cfg.RememberCookieName)
			next.ServeHTTP(w, r)
			return
		}

		// Rotate: consume the old remember session, mint a new marked one
		// plus a short-lived auth session.
		_ = cfg.SessionStore.SessionDeleteByKey(ctx, rc.Value)
		options := types.UserAuthOptions{UserIp: req.GetIP(r), UserAgent: r.UserAgent()}
		if cfg.IssueRememberSession != nil {
			newRsess, rErr := cfg.IssueRememberSession(ctx, rsess.GetUserID(), options)
			if rErr == nil && newRsess != nil && newRsess.GetKey() != "" {
				utils.RememberCookieSet(w, cfg.RememberCookieName, newRsess.GetKey(), cfg.RememberMeDays, cfg.SecureCookies)
			} else if cfg.Logger != nil {
				cfg.Logger.Error("remember-me: rotation failed",
					slog.Any("error", rErr),
					slog.String("user_id", rsess.GetUserID()))
			}
		}
		// Mint a fresh auth session — the cookie MUST hold the new session's
		// key, never the deleted remember key. Nil hook or error leaves the
		// user unauthenticated (fail-closed by design).
		if cfg.IssueAuthSession != nil && cfg.AuthCookieSet != nil {
			authSess, aErr := cfg.IssueAuthSession(ctx, rsess.GetUserID(), options)
			if aErr == nil && authSess != nil && authSess.GetKey() != "" {
				cfg.AuthCookieSet(w, r, authSess.GetKey())
			} else if cfg.Logger != nil {
				cfg.Logger.Error("remember-me: auth session mint failed",
					slog.Any("error", aErr),
					slog.String("user_id", rsess.GetUserID()))
			}
		}
		next.ServeHTTP(w, r)
	})
}
