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

	if !strings.Contains(html, "min-vh-100") {
		t.Fatalf("expected layout to vertically center the column, got %q", html)
	}
	if !strings.Contains(html, "bg-light") {
		t.Fatalf("expected layout to use the light background utility, got %q", html)
	}
	if !strings.Contains(html, "col-12") {
		t.Fatalf("expected layout to wrap content in a centered column, got %q", html)
	}

	if !strings.Contains(html, content) {
		t.Fatalf("expected layout to include content %q, got %q", content, html)
	}
}
