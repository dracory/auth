package browser

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
)

// navigate loads url in the tab and waits for the page body to be
// ready — the pages are server-rendered so the DOM is complete at
// that point (no AJAX hydration like the blogadmin Vue pages).
func navigate(t *testing.T, ctx context.Context, url string) {
	t.Helper()
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("navigate %s failed: %v", url, err)
	}
}

// count returns the number of elements matching the CSS selector.
func count(t *testing.T, ctx context.Context, sel string) int {
	t.Helper()
	var n int
	err := chromedp.Run(ctx,
		chromedp.Evaluate(fmt.Sprintf(`document.querySelectorAll(%q).length`, sel), &n),
	)
	if err != nil {
		t.Fatalf("query %q failed: %v", sel, err)
	}
	return n
}

// mustExist asserts every selector matches at least one element.
// Selectors carry the full expected markup — classes, attributes,
// types — so a missing class or dropped attribute fails the test.
func mustExist(t *testing.T, ctx context.Context, sels ...string) {
	t.Helper()
	for _, sel := range sels {
		if n := count(t, ctx, sel); n == 0 {
			t.Errorf("expected at least one element matching %q", sel)
		}
	}
}

// mustNotExist asserts every selector matches zero elements.
func mustNotExist(t *testing.T, ctx context.Context, sels ...string) {
	t.Helper()
	for _, sel := range sels {
		if n := count(t, ctx, sel); n != 0 {
			t.Errorf("expected no elements matching %q, found %d", sel, n)
		}
	}
}

// bodyText returns the rendered page text.
func bodyText(t *testing.T, ctx context.Context) string {
	t.Helper()
	var text string
	if err := chromedp.Run(ctx, chromedp.Evaluate(`document.body.innerText`, &text)); err != nil {
		t.Fatalf("reading body text failed: %v", err)
	}
	return text
}

// mustContain asserts the rendered page text contains every string.
func mustContain(t *testing.T, ctx context.Context, texts ...string) {
	t.Helper()
	body := bodyText(t, ctx)
	for _, s := range texts {
		if !strings.Contains(body, s) {
			t.Errorf("expected page text to contain %q", s)
		}
	}
}
