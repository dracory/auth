// Package api_magic_link_callback provides the single-use GET callback that
// redeems a magic login link.
package api_magic_link_callback

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

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

	if deps.AuthenticateViaUsername != nil {
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
