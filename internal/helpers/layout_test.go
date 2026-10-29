package helpers

import (
	"strings"
	"testing"
)

func TestLayoutWrapsContentWithFontAndStyle(t *testing.T) {
	content := "<p>Hello World</p>"
	html := Layout(content)

	if html == "" {
		t.Fatalf("expected non-empty layout output")
	}

	if strings.Contains(html, "fonts.bunny.net") {
		t.Fatalf("layout must not request external font hosts (breaks host CSP), got %q", html)
	}
	if !strings.Contains(html, "font-family: Nunito, sans-serif;") {
		t.Fatalf("expected layout to keep the Nunito font-family fallback, got %q", html)
	}

	if !strings.Contains(html, "background: #f8fafc;") {
		t.Fatalf("expected layout to include background style, got %q", html)
	}

	if !strings.Contains(html, content) {
		t.Fatalf("expected layout to include content %q, got %q", content, html)
	}
}
