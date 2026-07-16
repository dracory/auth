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
	fmt.Println("=== AUTHKNIGHT CALLBACK START ===")
	fmt.Printf("  Request URI: %s\n", r.RequestURI)
	fmt.Printf("  URL Path: %s\n", r.URL.Path)
	fmt.Printf("  Query: %v\n", r.URL.Query())
	fmt.Printf("  BaseURL: %s\n", deps.BaseURL)
	fmt.Printf("  RegisterURL: %s\n", deps.RegisterURL)
	fmt.Printf("  UrlRedirectOnSuccess: %s\n", deps.UrlRedirectOnSuccess)
	fmt.Printf("  UseCookies: %v\n", deps.UseCookies)
	fmt.Printf("  SetAuthCookie is nil: %v\n", deps.SetAuthCookie == nil)
	fmt.Printf("  UserFindByEmail is nil: %v\n", deps.UserFindByEmail == nil)
	fmt.Printf("  TemporaryKeySet is nil: %v\n", deps.TemporaryKeySet == nil)

	logger := deps.Logger
	if logger == nil {
		logger = slog.Default()
	}

	onceToken := req.GetStringTrimmed(r, "once")
	fmt.Printf("  onceToken: '%s'\n", onceToken)
	if onceToken == "" {
		fmt.Println("  ERROR: missing once parameter")
		logger.Error("authknight callback: missing once parameter")
		http.Redirect(w, r, deps.UrlRedirectOnSuccess, http.StatusTemporaryRedirect)
		return
	}

	baseURL := deps.BaseURL
	if baseURL == "" {
		baseURL = "https://authknight.com"
	}
	fmt.Printf("  baseURL (resolved): %s\n", baseURL)

	timeout := deps.HTTPTimeout
	if timeout <= 0 {
		timeout = DefaultHTTPTimeout
	}
	fmt.Printf("  timeout: %v\n", timeout)

	fmt.Println("  Calling verifyOnceToken...")
	email, err := verifyOnceToken(r.Context(), baseURL, onceToken, timeout)
	fmt.Printf("  verifyOnceToken result: email='%s' err=%v\n", email, err)
	if err != nil {
		fmt.Printf("  ERROR: token verification failed: %v\n", err)
		logger.Error("authknight callback: token verification failed",
			slog.String("error", err.Error()),
		)
		http.Redirect(w, r, deps.UrlRedirectOnSuccess, http.StatusTemporaryRedirect)
		return
	}

	if email == "" {
		fmt.Println("  ERROR: empty email in response")
		logger.Error("authknight callback: empty email in response")
		http.Redirect(w, r, deps.UrlRedirectOnSuccess, http.StatusTemporaryRedirect)
		return
	}

	fmt.Printf("  Email verified: %s\n", email)

	options := types.UserAuthOptions{
		UserIp:    req.GetIP(r),
		UserAgent: r.UserAgent(),
	}

	fmt.Println("  Calling UserFindByEmail...")
	userID, errFind := deps.UserFindByEmail(r.Context(), email, options)
	fmt.Printf("  UserFindByEmail result: userID='%s' errFind=%v\n", userID, errFind)
	if errFind == nil && userID != "" {
		fmt.Println("  Existing user - creating session")
		token, errToken := str.RandomFromGamma(TokenLength, TokenGamma)
		if errToken != nil {
			fmt.Printf("  ERROR: token generation failed: %v\n", errToken)
			logger.Error("authknight callback: token generation failed",
				slog.String("error", errToken.Error()),
			)
			http.Redirect(w, r, deps.UrlRedirectOnSuccess, http.StatusTemporaryRedirect)
			return
		}

		fmt.Println("  Storing auth token...")
		if errStore := deps.UserStoreAuthToken(r.Context(), token, userID, options); errStore != nil {
			fmt.Printf("  ERROR: auth token store failed: %v\n", errStore)
			logger.Error("authknight callback: auth token store failed",
				slog.String("error", errStore.Error()),
			)
			http.Redirect(w, r, deps.UrlRedirectOnSuccess, http.StatusTemporaryRedirect)
			return
		}

		if deps.UseCookies && deps.SetAuthCookie != nil {
			fmt.Println("  Setting auth cookie...")
			deps.SetAuthCookie(w, r, token)
		} else {
			fmt.Printf("  Skipping cookie: UseCookies=%v SetAuthCookie==nil=%v\n", deps.UseCookies, deps.SetAuthCookie == nil)
		}

		redirectURL := deps.UrlRedirectOnSuccess
		if deps.RedirectURL != nil {
			if custom := deps.RedirectURL(r.Context(), userID); custom != "" {
				redirectURL = custom
			}
		}

		fmt.Printf("  Redirecting existing user to: %s\n", redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	fmt.Println("  New user - redirecting to register")
	tempKey, errKey := str.RandomFromGamma(TokenLength, TokenGamma)
	if errKey != nil {
		fmt.Printf("  ERROR: temp key generation failed: %v\n", errKey)
		logger.Error("authknight callback: temp key generation failed",
			slog.String("error", errKey.Error()),
		)
		http.Redirect(w, r, deps.UrlRedirectOnSuccess, http.StatusTemporaryRedirect)
		return
	}

	fmt.Printf("  Generated tempKey: %s\n", tempKey)
	expiresSeconds := int(DefaultTempKeyExpiration.Seconds())
	fmt.Println("  Storing temp key...")
	if errSet := deps.TemporaryKeySet(tempKey, email, expiresSeconds); errSet != nil {
		fmt.Printf("  ERROR: temp key store failed: %v\n", errSet)
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
	fmt.Printf("  Redirecting new user to: %s\n", redirectURL)
	fmt.Println("=== AUTHKNIGHT CALLBACK END ===")
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// ApiAuthKnightCallbackWithAuth is a convenience wrapper that allows callers to pass
// a types.AuthAuthKnightInterface instead of manually wiring Dependencies.
func ApiAuthKnightCallbackWithAuth(w http.ResponseWriter, r *http.Request, a types.AuthAuthKnightInterface) {
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
			fmt.Println("  [SetAuthCookie] calling a.SetAuthCookie...")
			a.SetAuthCookie(w, r, token)
			fmt.Println("  [SetAuthCookie] done")
		},
		Logger: a.GetLogger(),
	}

	fmt.Println("[CALLBACK_WITH_AUTH] deps wired, calling ApiAuthKnightCallback")
	ApiAuthKnightCallback(w, r, deps)
}

// verifyOnceToken calls AuthKnight's /api/who endpoint to verify the once token
// and returns the validated email.
func verifyOnceToken(ctx context.Context, baseURL, onceToken string, timeout time.Duration) (string, error) {
	whoURL := baseURL + AuthKnightWhoEndpoint
	fmt.Printf("  [verifyOnceToken] whoURL=%s\n", whoURL)

	form := url.Values{}
	form.Set("once", onceToken)

	httpReq, errReq := http.NewRequestWithContext(ctx, http.MethodPost, whoURL, strings.NewReader(form.Encode()))
	if errReq != nil {
		fmt.Printf("  [verifyOnceToken] ERROR creating request: %v\n", errReq)
		return "", fmt.Errorf("failed to create request: %w", errReq)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: timeout}

	fmt.Println("  [verifyOnceToken] sending POST to AuthKnight...")
	resp, errDo := client.Do(httpReq)
	if errDo != nil {
		fmt.Printf("  [verifyOnceToken] ERROR API call failed: %v\n", errDo)
		return "", fmt.Errorf("authknight API call failed: %w", errDo)
	}
	defer resp.Body.Close()

	fmt.Printf("  [verifyOnceToken] response status code: %d\n", resp.StatusCode)

	body, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		fmt.Printf("  [verifyOnceToken] ERROR reading body: %v\n", errRead)
		return "", fmt.Errorf("failed to read response body: %w", errRead)
	}

	fmt.Printf("  [verifyOnceToken] response body: %s\n", string(body))

	var whoResp authKnightWhoResponse
	if errJSON := json.Unmarshal(body, &whoResp); errJSON != nil {
		fmt.Printf("  [verifyOnceToken] ERROR parsing JSON: %v\n", errJSON)
		return "", fmt.Errorf("failed to parse response: %w", errJSON)
	}

	fmt.Printf("  [verifyOnceToken] parsed: status='%s' message='%s' email='%s'\n", whoResp.Status, whoResp.Message, whoResp.Data.Email)

	if whoResp.Status != "success" {
		fmt.Println("  [verifyOnceToken] ERROR: status not success")
		return "", fmt.Errorf("authknight verification failed: status=%s message=%s", whoResp.Status, whoResp.Message)
	}

	fmt.Printf("  [verifyOnceToken] success, returning email='%s'\n", whoResp.Data.Email)
	return whoResp.Data.Email, nil
}
