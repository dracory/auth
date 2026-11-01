// Package api_magic_link_callback provides the single-use GET callback that
// redeems a magic login link.
package api_magic_link_callback

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/dracory/auth/internal/rules"
	"github.com/dracory/auth/types"
	"github.com/dracory/auth/utils"
	"github.com/dracory/req"
)

// Dependencies defines dependencies for the magic-link callback.
type Dependencies struct {
	// TemporaryKeyGet/Set/Delete resolve and consume the token. The token
	// is single-use: it is invalidated on first hit before any other check.
	TemporaryKeyGet    func(key string) (string, error)
	TemporaryKeySet    func(key string, value string, expiresSeconds int) error
	TemporaryKeyDelete func(key string) error

	// AuthenticateViaUsername hands off to the host's post-auth/session
	// pipeline. Called with the email extracted from the token payload.
	// In-flow (non-handler) callers get nil and use the returned result.
	AuthenticateViaUsername func(w http.ResponseWriter, r *http.Request, email string, firstName string, lastName string)

	// OnSuccess redirect override; when nil the request's `return` payload
	// (safe-relative only) is used, else the referrer/root.
	OnSuccess func(w http.ResponseWriter, r *http.Request, email string, payload utils.MagicLinkPayload)

	// CreateUserOnRegister defers user creation to the register form: when
	// the verified email has no account, the callback redirects to
	// RegisterURL with a verified-email token instead of calling
	// AuthenticateViaUsername (which would create a session-first user).
	CreateUserOnRegister bool

	// UserExists reports whether the email already has an account. Used by
	// CreateUserOnRegister; nil falls back to AuthenticateViaUsername.
	UserExists func(ctx context.Context, email string, options types.UserAuthOptions) (bool, error)

	// RegisterURL is the register page the verified-email token redirects
	// to (e.g. "/auth/register"). Only used with CreateUserOnRegister.
	RegisterURL string

	// EmailCaptured is an optional lead-capture hook fired after the token
	// validates, with verified=true.
	EmailCaptured func(ctx context.Context, email string, verified bool, options types.UserAuthOptions) error

	Logger *slog.Logger
}

var callbackMu sync.Mutex

// ApiMagicLinkCallback redeems a token: GET .../magiclink-callback?token=.
// Mountable independently of the login method so apps can reuse it for
// password-reset and other verify-by-email flows.
func ApiMagicLinkCallback(w http.ResponseWriter, r *http.Request, deps Dependencies) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := req.GetStringTrimmed(r, "token")
	if token == "" {
		http.Error(w, types.MsgLinkNotValidOrExpired, http.StatusBadRequest)
		return
	}

	payload, ok := consumeToken(r.Context(), r, deps, token)
	if !ok {
		// Single generic error for unknown/expired/used/mismatched tokens.
		http.Error(w, types.MsgLinkNotValidOrExpired, http.StatusBadRequest)
		return
	}

	options := types.UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	}

	// Lead capture: the email just proved ownership of the inbox. The
	// hook is best-effort — wiring sites log failures.
	if deps.EmailCaptured != nil {
		_ = deps.EmailCaptured(r.Context(), payload.Email, true, options)
	}

	// Register-first policy: verified-but-new emails get a verified-email
	// token and land on the register form instead of being logged in as a
	// freshly created (incomplete) user; degraded fallbacks log loudly
	// (see RegisterFirstRule).
	if rf := rules.NewRegisterFirstRule(
		deps.CreateUserOnRegister, deps.UserExists, deps.TemporaryKeySet,
		r.Context(), payload.Email, options); !rf.Fails() {
		vt, errToken := utils.IssueVerifiedEmailToken(deps.TemporaryKeySet, payload.Email, 0)
		if errToken != nil {
			logger := deps.Logger
			if logger == nil {
				logger = slog.Default()
			}
			logger.Error("magic link callback: verified-email token issue failed",
				slog.String("error", errToken.Error()))
		} else {
			registerURL := deps.RegisterURL
			if registerURL == "" {
				registerURL = "/register"
			}
			separator := "?"
			if strings.Contains(registerURL, "?") {
				separator = "&"
			}
			http.Redirect(w, r, registerURL+separator+"verified_key="+url.QueryEscape(vt), http.StatusTemporaryRedirect)
			return
		}
	} else if rf.Degraded() {
		logger := deps.Logger
		if logger == nil {
			logger = slog.Default()
		}
		logger.Warn("magic link callback: " + rf.Reason())
	}

	if deps.AuthenticateViaUsername != nil {
		// Forward the remember-me choice captured at send time: the
		// post-auth pipeline reads r.FormValue("remember"), and for a GET
		// callback that resolves to the URL query.
		if payload.Remember {
			// The form was already parsed when "token" was read above —
			// rewriting RawQuery now would leave r.Form stale, so the
			// value is injected into the parsed form instead.
			if err := r.ParseForm(); err == nil {
				r.Form.Set("remember", "true")
			}
		}
		deps.AuthenticateViaUsername(w, r, payload.Email, "", "")
		return
	}

	// No auth callback configured: still redirect safely so the endpoint is
	// usable for pure verification flows.
	if deps.OnSuccess != nil {
		deps.OnSuccess(w, r, payload.Email, *payload)
		return
	}
	redirect := "/"
	if utils.IsSafeReturnPath(payload.Return) {
		redirect = payload.Return
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

// consumeToken looks up and deletes the token (single-use), then validates
// the stored value: constant-time token compare and optional IP binding.
func consumeToken(ctx context.Context, r *http.Request, deps Dependencies, token string) (*utils.MagicLinkPayload, bool) {
	if deps.TemporaryKeyGet == nil {
		return nil, false
	}

	callbackMu.Lock()
	defer callbackMu.Unlock()

	cacheKey := utils.MagicLinkKeyPrefix + token
	raw, err := deps.TemporaryKeyGet(cacheKey)
	if err != nil || raw == "" {
		return nil, false
	}

	var payload utils.MagicLinkPayload
	if json.Unmarshal([]byte(raw), &payload) != nil || payload.Token == "" {
		return nil, false
	}

	// Defense in depth: the stored token must match the presented one.
	if subtle.ConstantTimeCompare([]byte(payload.Token), []byte(token)) != 1 {
		return nil, false
	}

	// Optional IP binding (only when send recorded an IP). Checked BEFORE
	// invalidation: a link triggered from a different IP — URL scanner,
	// corporate proxy, mobile network switch — must not burn the token.
	if payload.IP != "" && payload.IP != req.GetIP(r) {
		return nil, false
	}

	// All checks passed — consume: single-use, later hits get an empty slot.
	if invErr := utils.InvalidateTempKey(deps.TemporaryKeyDelete, deps.TemporaryKeySet, cacheKey); invErr != nil {
		logger := deps.Logger
		if logger == nil {
			logger = slog.Default()
		}
		logger.Error("magic link callback: failed to invalidate token",
			slog.String("error", invErr.Error()))
	}

	return &payload, true
}
