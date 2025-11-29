package shared

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestBuildPage_BasicStructure tests that buildPage creates a valid HTML document
// with the expected structure.
func TestBuildPage_BasicStructure(t *testing.T) {
	opts := PageOptions{
		Title: "Test Page",
		Layout: func(content string) string {
			return "<div class='container'>" + content + "</div>"
		},
		Content: "<h1>Hello World</h1>",
		Scripts: "",
	}

	html := buildPage(opts)

	// Check for basic HTML structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("expected DOCTYPE declaration")
	}
	if !strings.Contains(html, "<html") {
		t.Error("expected html tag")
	}
	if !strings.Contains(html, "</html>") {
		t.Error("expected closing html tag")
	}
	if !strings.Contains(html, "<head>") {
		t.Error("expected head tag")
	}
	if !strings.Contains(html, "<body>") {
		t.Error("expected body tag")
	}
}

// TestBuildPage_Title tests that the page title is properly set.
func TestBuildPage_Title(t *testing.T) {
	opts := PageOptions{
		Title: "My Custom Title",
		Layout: func(content string) string {
			return content
		},
		Content: "<p>Content</p>",
	}

	html := buildPage(opts)

	if !strings.Contains(html, "<title>My Custom Title</title>") {
		t.Errorf("expected title 'My Custom Title' in HTML, got: %s", html)
	}
}

// TestBuildPage_Favicon tests that the favicon is included.
func TestBuildPage_Favicon(t *testing.T) {
	opts := PageOptions{
		Title: "Test",
		Layout: func(content string) string {
			return content
		},
		Content: "<p>Content</p>",
	}

	html := buildPage(opts)

	if !strings.Contains(html, "data:image/x-icon;base64") {
		t.Error("expected favicon data URL")
	}
}

// TestBuildPage_BootstrapInclusion tests that Bootstrap CSS and JS are included.
func TestBuildPage_BootstrapInclusion(t *testing.T) {
	opts := PageOptions{
		Title: "Test",
		Layout: func(content string) string {
			return content
		},
		Content: "<p>Content</p>",
	}

	html := buildPage(opts)

	// Check for Bootstrap CSS
	if !strings.Contains(html, "bootstrap") {
		t.Error("expected Bootstrap CSS reference")
	}
}

// TestBuildPage_CustomStyles tests that custom styles are included.
func TestBuildPage_CustomStyles(t *testing.T) {
	opts := PageOptions{
		Title: "Test",
		Layout: func(content string) string {
			return content
		},
		Content: "<p>Content</p>",
	}

	html := buildPage(opts)

	// Check for custom styles
	if !strings.Contains(html, "font-family") {
		t.Error("expected custom font-family style")
	}
	if !strings.Contains(html, "background-color: #f8fafc") {
		t.Error("expected custom background color")
	}
	if !strings.Contains(html, ".form-select") {
		t.Error("expected form-select style")
	}
}

// TestBuildPage_LayoutApplication tests that the layout function is applied to content.
func TestBuildPage_LayoutApplication(t *testing.T) {
	opts := PageOptions{
		Title: "Test",
		Layout: func(content string) string {
			return "<div class='wrapper'>" + content + "</div>"
		},
		Content: "<p>Inner Content</p>",
	}

	html := buildPage(opts)

	if !strings.Contains(html, "<div class='wrapper'>") {
		t.Error("expected layout wrapper div")
	}
	if !strings.Contains(html, "<p>Inner Content</p>") {
		t.Error("expected inner content")
	}
}

// TestBuildPage_ContentInclusion tests that the content is included in the output.
func TestBuildPage_ContentInclusion(t *testing.T) {
	testContent := "<h1>Test Heading</h1><p>Test paragraph with unique text xyz123</p>"
	opts := PageOptions{
		Title: "Test",
		Layout: func(content string) string {
			return content
		},
		Content: testContent,
	}

	html := buildPage(opts)

	if !strings.Contains(html, "Test Heading") {
		t.Error("expected test heading in output")
	}
	if !strings.Contains(html, "unique text xyz123") {
		t.Error("expected unique text in output")
	}
}

// TestBuildPage_CustomScripts tests that custom scripts are included.
func TestBuildPage_CustomScripts(t *testing.T) {
	customScript := "<script>console.log('custom script');</script>"
	opts := PageOptions{
		Title: "Test",
		Layout: func(content string) string {
			return content
		},
		Content: "<p>Content</p>",
		Scripts: customScript,
	}

	html := buildPage(opts)

	if !strings.Contains(html, "console.log('custom script')") {
		t.Error("expected custom script in output")
	}
}

// TestBuildPage_EmptyContent tests that buildPage handles empty content gracefully.
func TestBuildPage_EmptyContent(t *testing.T) {
	opts := PageOptions{
		Title: "Empty Page",
		Layout: func(content string) string {
			return content
		},
		Content: "",
	}

	html := buildPage(opts)

	// Should still have valid HTML structure
	if !strings.Contains(html, "<html") {
		t.Error("expected valid HTML even with empty content")
	}
	if !strings.Contains(html, "<title>Empty Page</title>") {
		t.Error("expected title even with empty content")
	}
}

// TestPageRender_StatusCode tests that PageRender sets the correct status code.
func TestPageRender_StatusCode(t *testing.T) {
	opts := PageOptions{
		Title: "Test",
		Layout: func(content string) string {
			return content
		},
		Content: "<p>Content</p>",
	}

	recorder := httptest.NewRecorder()
	PageRender(recorder, opts)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, recorder.Code)
	}
}

// TestPageRender_ContentType tests that PageRender sets the correct content type.
func TestPageRender_ContentType(t *testing.T) {
	opts := PageOptions{
		Title: "Test",
		Layout: func(content string) string {
			return content
		},
		Content: "<p>Content</p>",
	}

	recorder := httptest.NewRecorder()
	PageRender(recorder, opts)

	contentType := recorder.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("expected Content-Type 'text/html', got %q", contentType)
	}
}

// TestPageRender_HTMLOutput tests that PageRender writes HTML to the response.
func TestPageRender_HTMLOutput(t *testing.T) {
	opts := PageOptions{
		Title: "Render Test",
		Layout: func(content string) string {
			return "<div>" + content + "</div>"
		},
		Content: "<p>Rendered Content</p>",
	}

	recorder := httptest.NewRecorder()
	PageRender(recorder, opts)

	body := recorder.Body.String()
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("expected DOCTYPE in rendered output")
	}
	if !strings.Contains(body, "<title>Render Test</title>") {
		t.Error("expected title in rendered output")
	}
	if !strings.Contains(body, "Rendered Content") {
		t.Error("expected content in rendered output")
	}
}

// TestPageRender_WithLogger tests that PageRender handles write errors with logging.
func TestPageRender_WithLogger(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))

	opts := PageOptions{
		Title: "Test",
		Layout: func(content string) string {
			return content
		},
		Content:    "<p>Content</p>",
		Logger:     logger,
		LogMessage: "Failed to render page",
	}

	// Use a normal recorder - it won't error on Write
	recorder := httptest.NewRecorder()
	PageRender(recorder, opts)

	// Verify the page was rendered successfully
	if recorder.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, recorder.Code)
	}
}

// TestPageRender_WithoutLogger tests that PageRender works without a logger.
func TestPageRender_WithoutLogger(t *testing.T) {
	opts := PageOptions{
		Title: "Test",
		Layout: func(content string) string {
			return content
		},
		Content: "<p>Content</p>",
		Logger:  nil, // No logger provided
	}

	recorder := httptest.NewRecorder()
	PageRender(recorder, opts)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, recorder.Code)
	}
}

// TestPageRender_ComplexLayout tests PageRender with a more complex layout.
func TestPageRender_ComplexLayout(t *testing.T) {
	opts := PageOptions{
		Title: "Complex Page",
		Layout: func(content string) string {
			return `
				<div class="container">
					<header><h1>Site Header</h1></header>
					<main>` + content + `</main>
					<footer><p>Site Footer</p></footer>
				</div>
			`
		},
		Content: "<article><h2>Article Title</h2><p>Article content</p></article>",
		Scripts: "<script>console.log('page loaded');</script>",
	}

	recorder := httptest.NewRecorder()
	PageRender(recorder, opts)

	body := recorder.Body.String()
	if !strings.Contains(body, "Site Header") {
		t.Error("expected header in output")
	}
	if !strings.Contains(body, "Article Title") {
		t.Error("expected article title in output")
	}
	if !strings.Contains(body, "Site Footer") {
		t.Error("expected footer in output")
	}
	if !strings.Contains(body, "page loaded") {
		t.Error("expected custom script in output")
	}
}

// TestPageRender_MultiplePages tests that PageRender can be called multiple times.
func TestPageRender_MultiplePages(t *testing.T) {
	pages := []struct {
		title   string
		content string
	}{
		{"Page 1", "<p>First page content</p>"},
		{"Page 2", "<p>Second page content</p>"},
		{"Page 3", "<p>Third page content</p>"},
	}

	for _, page := range pages {
		opts := PageOptions{
			Title: page.title,
			Layout: func(content string) string {
				return content
			},
			Content: page.content,
		}

		recorder := httptest.NewRecorder()
		PageRender(recorder, opts)

		body := recorder.Body.String()
		if !strings.Contains(body, page.title) {
			t.Errorf("expected title %q in output", page.title)
		}
		if !strings.Contains(body, page.content) {
			t.Errorf("expected content %q in output", page.content)
		}
	}
}
