package auth

import (
	"strings"
	"testing"
)

func TestWebpageIncludesTitleContentAndScripts(t *testing.T) {
	title := "My Test Page"
	content := "<div>Content Body</div>"
	scripts := "console.log('hello from scripts');"

	page := webpage(title, content, scripts)
	html := page.ToHTML()

	if html == "" {
		t.Fatalf("expected non-empty HTML from webpage helper")
	}

	if !strings.Contains(html, title) {
		t.Fatalf("expected HTML to contain title %q, got %q", title, html)
	}

	if !strings.Contains(html, content) {
		t.Fatalf("expected HTML to contain content %q, got %q", content, html)
	}

	if !strings.Contains(html, scripts) {
		t.Fatalf("expected HTML to contain scripts %q, got %q", scripts, html)
	}

	// favicon data URL should be present
	if !strings.Contains(html, "data:image/x-icon;base64,") {
		t.Fatalf("expected HTML to contain favicon data URL, got %q", html)
	}
}
