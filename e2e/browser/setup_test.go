// Package browser contains end-to-end tests that drive the auth UI
// pages with a real headless Chrome via chromedp (CDP). The auth
// handler is served in-process by httptest on a random port — no
// binary build needed.
//
// This complements the HTTP-level flow tests in e2e/: those exercise
// the API end to end via httptest.NewRecorder, while these prove the
// server-rendered pages and their embedded JS actually work in a
// browser (form validation, fetch wiring, redirects).
package browser

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/dracory/auth"
	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/types"
)

var (
	baseURL  string
	allocCtx context.Context
)

const authPath = "/auth"

func TestMain(m *testing.M) {
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	baseURL = srv.URL

	cfg := testutils.NewUsernameAndPasswordConfigForTest()
	cfg.Endpoint = baseURL + authPath
	cfg.UrlRedirectOnSuccess = baseURL + "/dashboard"
	cfg.FuncUserLogin = func(ctx context.Context, username, password string, options types.UserAuthOptions) (string, error) {
		if username == "jane@example.com" && password == "long-enough-password" {
			return "user-1", nil
		}
		return "", nil
	}
	cfg.FuncUserFindByUsername = func(ctx context.Context, username, firstName, lastName string, options types.UserAuthOptions) (string, error) {
		return "user-1", nil
	}

	a, err := auth.NewUsernameAndPasswordAuth(cfg)
	if err != nil {
		panic("e2e/browser: failed to create auth: " + err.Error())
	}

	mux.Handle(authPath+"/", a.Router())
	mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body>dashboard</body></html>"))
	})

	// One shared browser process for the entire test run.
	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Headless)
	if bin := chromeBin(); bin != "" {
		opts = append(opts, chromedp.ExecPath(bin))
	}
	var cancel context.CancelFunc
	allocCtx, cancel = chromedp.NewExecAllocator(context.Background(), opts...)

	code := m.Run()

	cancel()
	srv.Close()
	os.Exit(code)
}

// newTab opens a fresh tab in the shared browser and returns its
// context. Each tab gets a 30s deadline; cleanup closes the tab.
func newTab(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := chromedp.NewContext(allocCtx)
	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	t.Cleanup(func() { cancelTimeout(); cancel() })
	return ctx
}

// pageURL builds a page URL under the auth mount, e.g. /auth/login.
func pageURL(page string) string {
	return baseURL + authPath + "/" + page
}

// chromeBin resolves a headless Chromium binary. Order: CHROME_BIN env,
// Playwright's downloaded headless shell, empty (chromedp finds Chrome
// on PATH / standard install locations).
func chromeBin() string {
	if b := os.Getenv("CHROME_BIN"); b != "" {
		return b
	}
	patterns := []string{
		// Windows
		filepath.Join(os.Getenv("LOCALAPPDATA"),
			"ms-playwright/chromium_headless_shell-*/chrome-headless-shell-win64/chrome-headless-shell.exe"),
		// WSL: Windows-side Playwright shell via interop
		"/mnt/*/Users/*/AppData/Local/ms-playwright/chromium_headless_shell-*/chrome-headless-shell-win64/chrome-headless-shell.exe",
		// Linux
		filepath.Join(os.Getenv("HOME"),
			".cache/ms-playwright/chromium_headless_shell-*/chrome-headless-shell-linux64/chrome-headless-shell"),
	}
	for _, pattern := range patterns {
		if matches, _ := filepath.Glob(pattern); len(matches) > 0 {
			return matches[len(matches)-1]
		}
	}
	return ""
}
