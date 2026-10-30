// Remember-me exchange middleware.
//
// If the auth cookie is absent/expired but the remember cookie holds a
// valid marked session, the middleware exchanges it: the used remember
// session is shortened to a grace window (not deleted — concurrent browser
// restores all carry the same cookie) and a fresh marked session plus a
// short-lived auth session are issued.
//
// Two misuse directions are guarded:
//   - A marked remember key presented in the AUTH cookie is stripped from
//     the request and the cookie expired — a remember token must never act
//     as a long-lived auth credential.
//   - An unmarked or unknown key in the REMEMBER cookie is rejected and
//     the cookie removed.
//
// On a successful exchange the request's Cookie header is rewritten, so
// downstream auth middleware sees the minted session on the same request.
package middlewares

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

// SessionUpdater is an optional extension of types.SessionStoreInterface
// for stores that can update a session in place. It is used to shorten a
// consumed remember session to its grace window; stores without it fall
// back to immediate deletion.
type SessionUpdater interface {
	SessionUpdate(ctx context.Context, session types.SessionInterface) error
}

// RememberMeConfig wires the remember-me exchange middleware.
type RememberMeConfig struct {
	// Enabled is the master switch for the restore path. The auth-cookie
	// misuse check still runs when false — remember sessions issued before
	// the feature was switched off must never act as auth credentials.
	Enabled bool

	// SessionStore looks up and rotates sessions.
	SessionStore types.SessionStoreInterface

	// AuthCookieName is the auth session cookie name.
	AuthCookieName string
	// RememberCookieName is the remember token cookie name.
	RememberCookieName string

	// IssueRememberSession creates a marked long-lived session (same hook
	// as the pipeline's PostAuthHooks.IssueRememberSession).
	IssueRememberSession func(ctx context.Context, userID string, options types.UserAuthOptions) (types.SessionInterface, error)

	// IssueAuthSession mints the short-lived auth session exchanged from
	// the remember session — the auth cookie must carry ITS key, never the
	// (now-consumed) remember key.
	IssueAuthSession func(ctx context.Context, userID string, options types.UserAuthOptions) (types.SessionInterface, error)

	// UserIsActive is an optional account-state check; when it returns
	// false the remember session is revoked (deleted + cookie expired).
	// Nil allows every user.
	UserIsActive func(ctx context.Context, userID string) bool

	// RememberMeDays is the rotated cookie/session TTL (default 30).
	RememberMeDays int

	// RotationGrace is how long a consumed remember session stays valid
	// after rotation (default 60s). Concurrent restores absorb within the
	// window; it is only ever shortened, never extended.
	RotationGrace time.Duration

	// SecureCookies controls the Secure flag on re-issued cookies.
	SecureCookies bool

	// AuthCookieSet writes the auth cookie for the minted session. Nil
	// falls back to a plain HttpOnly Set-Cookie.
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
	if cfg.RotationGrace <= 0 {
		cfg.RotationGrace = 60 * time.Second
	}
	if cfg.AuthCookieSet == nil {
		cfg.AuthCookieSet = func(w http.ResponseWriter, r *http.Request, token string) {
			http.SetCookie(w, &http.Cookie{
				Name:     cfg.AuthCookieName,
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				Secure:   cfg.SecureCookies,
				SameSite: http.SameSiteLaxMode,
			})
		}
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
			if sess != nil && !sessionExpired(sess) && sess.GetValue(utils.RememberSessionMarker) == "" {
				next.ServeHTTP(w, r) // normal session — proceed
				return
			}
			// Misuse: a remember key presented as an auth credential is
			// stripped from the request and the cookie expired — the
			// session itself is left for the rotation flow below (it may
			// still be a legitimately held remember token).
			if sess != nil && sess.GetValue(utils.RememberSessionMarker) != "" {
				removeRequestCookie(r, cfg.AuthCookieName)
				expireCookie(w, cfg.AuthCookieName, cfg.SecureCookies)
			}
		}

		if !cfg.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Remember cookie present?
		rc, err := r.Cookie(cfg.RememberCookieName)
		if err != nil || rc.Value == "" {
			next.ServeHTTP(w, r)
			return
		}
		rsess, err := cfg.SessionStore.SessionFindByKey(ctx, rc.Value)
		if err != nil || rsess == nil || sessionExpired(rsess) ||
			rsess.GetValue(utils.RememberSessionMarker) == "" {
			// Misuse B: unmarked, expired, or unknown session in the
			// remember cookie.
			utils.RememberCookieRemove(w, cfg.RememberCookieName)
			next.ServeHTTP(w, r)
			return
		}

		// Deactivated or deleted account: revoke the remember credential.
		if cfg.UserIsActive != nil && !cfg.UserIsActive(ctx, rsess.GetUserID()) {
			_ = cfg.SessionStore.SessionDeleteByKey(ctx, rc.Value)
			utils.RememberCookieRemove(w, cfg.RememberCookieName)
			next.ServeHTTP(w, r)
			return
		}

		options := types.UserAuthOptions{UserIp: req.GetIP(r), UserAgent: r.UserAgent()}

		// Mint a fresh auth session — the cookie MUST hold the new session's
		// key, never the consumed remember key. The request cookie header is
		// rewritten so downstream auth middleware authenticates on this same
		// request (AddCookie would leave the stale client cookie first in
		// line). Nil hook or error leaves the user unauthenticated
		// (fail-closed by design).
		if cfg.IssueAuthSession != nil {
			authSess, aErr := cfg.IssueAuthSession(ctx, rsess.GetUserID(), options)
			if aErr == nil && authSess != nil && authSess.GetKey() != "" {
				cfg.AuthCookieSet(w, r, authSess.GetKey())
				setRequestCookie(r, cfg.AuthCookieName, authSess.GetKey())
			} else if cfg.Logger != nil {
				cfg.Logger.Error("remember-me: auth session mint failed",
					slog.Any("error", aErr),
					slog.String("user_id", rsess.GetUserID()))
			}
		}

		// Rotate: shorten the used remember session to the grace window —
		// never extend — then issue a fresh marked session. Stores without
		// in-place update fall back to immediate deletion.
		graceEnd := time.Now().Add(cfg.RotationGrace)
		rotate := true
		if rsess.GetExpiresAt().After(graceEnd) {
			rsess.SetExpiresAt(graceEnd)
			if updater, ok := cfg.SessionStore.(SessionUpdater); ok {
				if err := updater.SessionUpdate(ctx, rsess); err != nil {
					// Fail closed: without the shortening the old key keeps
					// its full TTL, so issuing a fresh cookie would widen
					// the replay window instead of narrowing it.
					if cfg.Logger != nil {
						cfg.Logger.Error("remember-me: grace shortening failed",
							slog.Any("error", err))
					}
					rotate = false
				}
			} else {
				_ = cfg.SessionStore.SessionDeleteByKey(ctx, rc.Value)
			}
		}
		if rotate && cfg.IssueRememberSession != nil {
			newRsess, rErr := cfg.IssueRememberSession(ctx, rsess.GetUserID(), options)
			if rErr == nil && newRsess != nil && newRsess.GetKey() != "" {
				utils.RememberCookieSet(w, cfg.RememberCookieName, newRsess.GetKey(), cfg.RememberMeDays, cfg.SecureCookies)
			} else if cfg.Logger != nil {
				cfg.Logger.Error("remember-me: rotation failed",
					slog.Any("error", rErr),
					slog.String("user_id", rsess.GetUserID()))
			}
		}
		next.ServeHTTP(w, r)
	})
}

// sessionExpired reports whether the session's expiry has passed. Sessions
// with no expiry are treated as expired (fail closed).
func sessionExpired(s types.SessionInterface) bool {
	return !s.GetExpiresAt().After(time.Now())
}

// setRequestCookie rebuilds the request Cookie header with name=value,
// replacing any existing entry of the same name — r.Cookie resolves the
// first match, so appending would leave a stale duplicate in place.
func setRequestCookie(r *http.Request, name, value string) {
	removeRequestCookie(r, name)
	r.AddCookie(&http.Cookie{Name: name, Value: value}) // #nosec G124 -- in-request cookie injection, not sent to browser
}

// removeRequestCookie rebuilds the request Cookie header without the named
// cookie. Remaining cookies are re-serialized via Cookie.String so quoting
// and escaping survive the rebuild.
func removeRequestCookie(r *http.Request, name string) {
	var kept []string
	for _, c := range r.Cookies() {
		if c.Name != name {
			if s := c.String(); s != "" {
				kept = append(kept, s)
			}
		}
	}
	r.Header.Set("Cookie", strings.Join(kept, "; "))
}

// expireCookie emits an expiring Set-Cookie for name.
func expireCookie(w http.ResponseWriter, name string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "none",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   -1,
	})
}
