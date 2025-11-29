package utils

import (
	"strings"
	"testing"
)

func TestWebpage(t *testing.T) {
	title := "Test Page Title"
	content := "<div>Test Content</div>"
	scripts := "console.log('test');"

	page := Webpage(title, content, scripts)
	html := page.ToHTML()

	if html == "" {
		t.Fatalf("expected non-empty HTML from Webpage")
	}

	if !strings.Contains(html, title) {
		t.Errorf("expected HTML to contain title %q", title)
	}

	if !strings.Contains(html, content) {
		t.Errorf("expected HTML to contain content %q", content)
	}

	if !strings.Contains(html, scripts) {
		t.Errorf("expected HTML to contain scripts %q", scripts)
	}

	// Verify Bootstrap CSS is included
	if !strings.Contains(html, "bootstrap") {
		t.Errorf("expected HTML to include Bootstrap CSS")
	}

	// Verify favicon is included
	if !strings.Contains(html, "data:image/x-icon;base64,") {
		t.Errorf("expected HTML to include favicon data URL")
	}
}
