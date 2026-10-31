package api_authknight_callback

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dracory/auth/internal/core"
	"github.com/dracory/auth/types"
	"github.com/dracory/req"
	"github.com/dracory/str"
)

// authKnightWhoResponse represents the response from AuthKnight's /api/who endpoint.
type authKnightWhoResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Email   string `json:"email"`
		BackURL string `json:"back_url"`
	} `json:"data"`
}

// ApiAuthKnightCallback handles the AuthKnight callback redirect.
// It extracts the "once" parameter, verifies it with AuthKnight's API,
// and either creates a session (existing user) or redirects to the register page (new user).
func ApiAuthKnightCallback(w http.ResponseWriter, r *http.Request, deps Dependencies) {

	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}

	onceToken := req.GetStringTrimmed(r, "once")
	if onceToken == "" {
		logger.Error("authknight callback: missing once parameter")
		http.Redirect(w, r, deps.UrlRedirectOnSuccess, http.StatusTemporaryRedirect)
		return
	}

	baseURL := deps.BaseURL
	if baseURL == "" {
		baseURL = "https://authknight.com"
	}

	timeout := deps.HTTPTimeout
	if timeout <= 0 {
		timeout = DefaultHTTPTimeout
	}

	email, err := verifyOnceToken(r.Context(), baseURL, onceToken, timeout)
	if err != nil {
		logger.Error("authknight callback: token verification failed",
			slog.String("error", err.Error()),
		)
		http.Redirect(w, r, deps.UrlRedirectOnSuccess, http.StatusTemporaryRedirect)
		return
	}

	if email == "" {
		logger.Error("authknight callback: empty email in response")
		http.Redirect(w, r, deps.UrlRedirectOnSuccess, http.StatusTemporaryRedirect)
		return
	}


	options := types.UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	}

	userID, errFind := deps.UserFindByEmail(r.Context(), email, options)
	if errFind == nil && userID != "" {
		token, errToken := str.RandomFromGamma(TokenLength, TokenGamma)
		if errToken != nil {
			logger.Error("authknight callback: token generation failed",
				slog.String("error", errToken.Error()),
			)
			http.Redirect(w, r, deps.UrlRedirectOnSuccess, http.StatusTemporaryRedirect)
			return
		}

		if errStore := deps.UserStoreAuthToken(r.Context(), token, userID, options); errStore != nil {
			logger.Error("authknight callback: auth token store failed",
				slog.String("error", errStore.Error()),
			)
			http.Redirect(w, r, deps.UrlRedirectOnSuccess, http.StatusTemporaryRedirect)
			return
		}

		if deps.UseCookies && deps.SetAuthCookie != nil {
			deps.SetAuthCookie(w, r, token)
		} else {
		}

		redirectURL := deps.UrlRedirectOnSuccess
		if deps.RedirectURL != nil {
			if custom := deps.RedirectURL(r.Context(), userID); custom != "" {
				redirectURL = custom
			}
		}

		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	tempKey, errKey := str.RandomFromGamma(TokenLength, TokenGamma)
	if errKey != nil {
		logger.Error("authknight callback: temp key generation failed",
			slog.String("error", errKey.Error()),
		)
		http.Redirect(w, r, deps.UrlRedirectOnSuccess, http.StatusTemporaryRedirect)
		return
	}

	expiresSeconds := int(DefaultTempKeyExpiration.Seconds())
	if errSet := deps.TemporaryKeySet(tempKey, email, expiresSeconds); errSet != nil {
		logger.Error("authknight callback: temp key store failed",
			slog.String("error", errSet.Error()),
		)
		http.Redirect(w, r, deps.UrlRedirectOnSuccess, http.StatusTemporaryRedirect)
		return
	}

	registerURL := deps.RegisterURL
	if registerURL == "" {
		registerURL = "/register"
	}

	separator := "?"
	if strings.Contains(registerURL, "?") {
		separator = "&"
	}
	redirectURL := fmt.Sprintf("%s%sak_key=%s", registerURL, separator, url.QueryEscape(tempKey))
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// ApiAuthKnightCallbackWithAuth is a convenience wrapper that allows callers to pass
// a types.AuthAuthKnightInterface instead of manually wiring Dependencies.
func ApiAuthKnightCallbackWithAuth(w http.ResponseWriter, r *http.Request, a types.AuthAuthKnightInterface) {
	// Post-auth pipeline preferred — same convention as
	// authenticateViaUsername: when ANY PostAuthHook is configured the
	// shared pipeline (allowlist → find-or-create → session → cookie →
	// redirect) replaces the legacy token-store path, so the session
	// implementation owns token creation instead of a random string.
	if hk, ok := any(a).(interface{ GetPostAuthHooks() types.PostAuthHooks }); ok {
		hooks := hk.GetPostAuthHooks()
		if hooks.EmailAllowed != nil || hooks.UserFindByEmailOrCreate != nil ||
			hooks.UserIsActive != nil || hooks.SessionCreate != nil ||
			hooks.SetAuthCookie != nil || hooks.RedirectURL != nil ||
			hooks.RegisterRules != nil {
			apiAuthKnightCallbackPostAuth(w, r, a, hooks)
			return
		}
	}
	deps := Dependencies{
		BaseURL:              AuthKnightDefaultBaseURL,
		HTTPTimeout:          a.GetAuthKnightHTTPTimeout(),
		UserFindByEmail:      a.GetAuthKnightUserFindByEmail(),
		UserStoreAuthToken:   a.GetFuncUserStoreAuthToken(),
		TemporaryKeySet:      a.GetFuncTemporaryKeySet(),
		TemporaryKeyGet:      a.GetFuncTemporaryKeyGet(),
		RedirectURL:          a.GetAuthKnightRedirectURL(),
		UrlRedirectOnSuccess: a.LinkRedirectOnSuccess(),
		RegisterURL:          a.LinkRegister(),
		UseCookies:           a.GetUseCookies(),
		SetAuthCookie: func(w http.ResponseWriter, r *http.Request, token string) {
			a.SetAuthCookie(w, r, token)
		},
		Logger: a.GetLogger(),
	}

	ApiAuthKnightCallback(w, r, deps)
}

// apiAuthKnightCallbackPostAuth runs the AuthKnight callback through the
// shared post-auth pipeline: verify the once token with AuthKnight, then
// let PostAuth handle allowlist, find-or-create, session, cookie, and
// redirect. New users are auto-created by the pipeline and land on the
// register-completion redirect its hooks produce.
func apiAuthKnightCallbackPostAuth(w http.ResponseWriter, r *http.Request, a types.AuthAuthKnightInterface, hooks types.PostAuthHooks) {
	logger := a.GetLogger()
	if logger == nil {
		logger = slog.Default()
	}

	fallback := a.LinkRedirectOnSuccess()
	onceToken := req.GetStringTrimmed(r, "once")
	if onceToken == "" {
		logger.Error("authknight callback: missing once parameter")
		http.Redirect(w, r, fallback, http.StatusTemporaryRedirect)
		return
	}

	timeout := a.GetAuthKnightHTTPTimeout()
	if timeout <= 0 {
		timeout = DefaultHTTPTimeout
	}

	email, err := verifyOnceToken(r.Context(), AuthKnightDefaultBaseURL, onceToken, timeout)
	if err != nil || email == "" {
		logger.Error("authknight callback: token verification failed")
		http.Redirect(w, r, fallback, http.StatusTemporaryRedirect)
		return
	}

	deps := core.PostAuthDeps{
		Hooks:   hooks,
		Options: types.UserAuthOptions{UserIp: req.GetIP(r), UserAgent: r.UserAgent()},
		Logger:  logger,
	}
	if g, ok := any(a).(interface{ GetRememberMeEnabled() bool }); ok {
		deps.RememberMeEnabled = g.GetRememberMeEnabled()
	}
	if g, ok := any(a).(interface{ GetRememberMeDays() int }); ok {
		deps.RememberMeDays = g.GetRememberMeDays()
	}
	if g, ok := any(a).(interface{ GetRememberCookieName() string }); ok {
		deps.RememberCookieName = g.GetRememberCookieName()
	}
	if g, ok := any(a).(interface{ GetSecureCookies() bool }); ok {
		deps.SecureCookies = g.GetSecureCookies()
	}

	result, err := core.PostAuth(r.Context(), w, r, email, "", deps)
	if err != nil {
		logger.Error("authknight callback: post-auth failed",
			slog.String("error", err.Error()))
		http.Redirect(w, r, fallback, http.StatusTemporaryRedirect)
		return
	}

	redirect := result.Redirect
	if redirect == "" {
		redirect = fallback
	}
	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
}

// verifyOnceToken calls AuthKnight's /api/who endpoint to verify the once token
// and returns the validated email.
func verifyOnceToken(ctx context.Context, baseURL, onceToken string, timeout time.Duration) (string, error) {
	whoURL := baseURL + AuthKnightWhoEndpoint

	form := url.Values{}
	form.Set("once", onceToken)

	httpReq, errReq := http.NewRequestWithContext(ctx, http.MethodPost, whoURL, strings.NewReader(form.Encode()))
	if errReq != nil {
		return "", fmt.Errorf("failed to create request: %w", errReq)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: timeout}

	resp, errDo := client.Do(httpReq)
	if errDo != nil {
		return "", fmt.Errorf("authknight API call failed: %w", errDo)
	}
	defer resp.Body.Close()


	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authknight API returned status %d", resp.StatusCode)
	}

	body, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return "", fmt.Errorf("failed to read response body: %w", errRead)
	}


	var whoResp authKnightWhoResponse
	if errJSON := json.Unmarshal(body, &whoResp); errJSON != nil {
		return "", fmt.Errorf("failed to parse response: %w", errJSON)
	}


	if whoResp.Status != "success" {
		return "", fmt.Errorf("authknight verification failed: status=%s message=%s", whoResp.Status, whoResp.Message)
	}

	return whoResp.Data.Email, nil
}
