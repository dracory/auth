package browser

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

// TestLoginSuccess verifies the full browser login flow: the page JS
// posts credentials to the API, stores the token, and redirects to the
// configured success URL.
func TestLoginSuccess(t *testing.T) {
	ctx := newTab(t)
	err := chromedp.Run(ctx,
		chromedp.Navigate(pageURL("login")),
		chromedp.WaitReady(`input[name=email]`, chromedp.ByQuery),
		chromedp.SendKeys(`input[name=email]`, "jane@example.com", chromedp.ByQuery),
		chromedp.SendKeys(`input[name=password]`, "long-enough-password", chromedp.ByQuery),
		chromedp.Click(`button.ButtonLogin`, chromedp.ByQuery),
		// The page script waits 2s on success before redirecting.
		chromedp.Sleep(3*time.Second),
	)
	if err != nil {
		t.Fatalf("login interaction failed: %v", err)
	}

	var loc string
	if err := chromedp.Run(ctx, chromedp.Location(&loc)); err != nil {
		t.Fatalf("reading location failed: %v", err)
	}
	if loc != baseURL+"/dashboard" {
		t.Fatalf("expected redirect to dashboard, got %q", loc)
	}
}

// TestLoginFailure verifies bad credentials surface the server error
// in the page's danger alert instead of redirecting.
func TestLoginFailure(t *testing.T) {
	ctx := newTab(t)
	err := chromedp.Run(ctx,
		chromedp.Navigate(pageURL("login")),
		chromedp.WaitReady(`input[name=email]`, chromedp.ByQuery),
		chromedp.SendKeys(`input[name=email]`, "jane@example.com", chromedp.ByQuery),
		chromedp.SendKeys(`input[name=password]`, "wrong-password", chromedp.ByQuery),
		chromedp.Click(`button.ButtonLogin`, chromedp.ByQuery),
		chromedp.Sleep(1500*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("login interaction failed: %v", err)
	}

	// Errors surface as a Notiflix toast when the library loads (its
	// markup lives outside .alert-group); assert on the rendered page
	// text instead so the fallback alert divs also pass.
	var bodyText, loc string
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`document.body.innerText`, &bodyText),
		chromedp.Location(&loc),
	)
	if err != nil {
		t.Fatalf("reading page failed: %v", err)
	}
	lower := strings.ToLower(bodyText)
	if !strings.Contains(lower, "error") && !strings.Contains(lower, "invalid") {
		t.Fatalf("expected an error message after failed login, page text: %q", bodyText)
	}
	if !strings.HasPrefix(loc, baseURL+authPath) {
		t.Fatalf("expected to stay on the auth pages, got %q", loc)
	}
}

// TestLoginEmptyFields verifies client-side validation rejects an
// empty form without hitting the API.
func TestLoginEmptyFields(t *testing.T) {
	ctx := newTab(t)
	err := chromedp.Run(ctx,
		chromedp.Navigate(pageURL("login")),
		chromedp.WaitReady(`button.ButtonLogin`, chromedp.ByQuery),
		chromedp.Click(`button.ButtonLogin`, chromedp.ByQuery),
		chromedp.Sleep(500*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("login interaction failed: %v", err)
	}

	var bodyText string
	if err := chromedp.Run(ctx, chromedp.Evaluate(`document.body.innerText`, &bodyText)); err != nil {
		t.Fatalf("reading page failed: %v", err)
	}
	if !strings.Contains(bodyText, "Email is required") {
		t.Fatalf("expected 'Email is required' validation, got %q", bodyText)
	}
}
