package shared

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracory/auth/types"
)

func newRequest() *http.Request {
	return httptest.NewRequest(http.MethodGet, "/auth/login", nil)
}

// TestBuildPage_BasicStructure tests that buildPage creates a valid HTML document
// with the expected structure when no custom layout is provided.
func TestBuildPage_BasicStructure(t *testing.T) {
	opts := PageOptions{
		Title:   "Test Page",
		Content: "<h1>Hello World</h1>",
	}

	html := buildPage(nil, nil, opts)

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
		Title:   "My Custom Title",
		Content: "<p>Content</p>",
	}

	html := buildPage(nil, nil, opts)

	if !strings.Contains(html, "<title>My Custom Title</title>") {
		t.Errorf("expected title 'My Custom Title' in HTML, got: %s", html)
	}
}

// TestBuildPage_Favicon tests that the favicon is included.
func TestBuildPage_Favicon(t *testing.T) {
	opts := PageOptions{
		Title:   "Test",
		Content: "<p>Content</p>",
	}

	html := buildPage(nil, nil, opts)

	if !strings.Contains(html, "data:image/x-icon;base64") {
		t.Error("expected favicon data URL")
	}
}

// TestBuildPage_BootstrapInclusion tests that Bootstrap CSS and JS are included.
func TestBuildPage_BootstrapInclusion(t *testing.T) {
	opts := PageOptions{
		Title:   "Test",
		Content: "<p>Content</p>",
	}

	html := buildPage(nil, nil, opts)

	if !strings.Contains(html, "bootstrap") {
		t.Error("expected Bootstrap CSS reference")
	}
}

// TestBuildPage_CustomStyles tests that the default document injects no
// package-owned opinionated CSS — pages must be restylable with a plain
// Bootstrap theme — and that the bootstrap-icons font stylesheet is loaded
// from the CDN (its font URLs are relative, so it cannot be vendored inline).
func TestBuildPage_CustomStyles(t *testing.T) {
	opts := PageOptions{
		Title:   "Test",
		Content: "<p>Content</p>",
	}

	html := buildPage(nil, nil, opts)

	if !strings.Contains(html, "bootstrap-icons") {
		t.Error("expected bootstrap-icons stylesheet")
	}
	for _, stale := range []string{"Nunito", "#f8fafc"} {
		if strings.Contains(html, stale) {
			t.Errorf("package-owned base styles must not be injected: %s", stale)
		}
	}
}

// TestBuildPage_DefaultBodyDecoration tests that content is wrapped in the
// default padded section even when a custom layout is absent.
func TestBuildPage_DefaultBodyDecoration(t *testing.T) {
	opts := PageOptions{
		Title:   "Test",
		Content: "<p>Inner Content</p>",
	}

	html := buildPage(nil, nil, opts)

	if !strings.Contains(html, "min-vh-100") {
		t.Error("expected default padded section wrapper")
	}
	if !strings.Contains(html, "<p>Inner Content</p>") {
		t.Error("expected inner content")
	}
}

// TestBuildPage_ContentInclusion tests that the content is included in the output.
func TestBuildPage_ContentInclusion(t *testing.T) {
	testContent := "<h1>Test Heading</h1><p>Test paragraph with unique text xyz123</p>"
	opts := PageOptions{
		Title:   "Test",
		Content: testContent,
	}

	html := buildPage(nil, nil, opts)

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
		Title:   "Test",
		Content: "<p>Content</p>",
		Scripts: customScript,
	}

	html := buildPage(nil, nil, opts)

	if !strings.Contains(html, "console.log('custom script')") {
		t.Error("expected custom script in output")
	}
}

// TestBuildPage_EmptyContent tests that buildPage handles empty content gracefully.
func TestBuildPage_EmptyContent(t *testing.T) {
	opts := PageOptions{
		Title:   "Empty Page",
		Content: "",
	}

	html := buildPage(nil, nil, opts)

	if !strings.Contains(html, "<html") {
		t.Error("expected valid HTML even with empty content")
	}
	if !strings.Contains(html, "<title>Empty Page</title>") {
		t.Error("expected title even with empty content")
	}
}

// TestBuildPage_CustomLayoutDocument tests that a host-provided Layout
// produces the entire document — the package's shell is not used.
func TestBuildPage_CustomLayoutDocument(t *testing.T) {
	var gotTitle string
	var gotOptions types.LayoutOptions
	opts := PageOptions{
		Title:   "Host Title",
		Content: "<p>Card</p>",
		Scripts: "<script>pageJS();</script>",
		Layout: func(w http.ResponseWriter, r *http.Request, title string, body string, options types.LayoutOptions) string {
			gotTitle = title
			gotOptions = options
			return "<html><head><title>" + title + "</title></head><body><nav>HostNav</nav>" + body + "</body></html>"
		},
	}

	html := buildPage(httptest.NewRecorder(), newRequest(), opts)

	if gotTitle != "Host Title" {
		t.Errorf("expected layout to receive title, got %q", gotTitle)
	}
	if !strings.Contains(html, "HostNav") {
		t.Error("expected host document chrome in output")
	}
	if !strings.Contains(html, "<p>Card</p>") {
		t.Error("expected page body embedded in host document")
	}
	if strings.Contains(html, "data:image/x-icon;base64") {
		t.Error("package favicon should not appear when host layout renders the document")
	}
	// The package's required assets must be passed through options so the
	// host document can inject them.
	if len(gotOptions.StyleURLs) == 0 || !strings.Contains(strings.Join(gotOptions.StyleURLs, "\n"), "bootstrap-icons") {
		t.Error("expected bootstrap-icons URL in layout options")
	}
	if len(gotOptions.Scripts) == 0 {
		t.Error("expected required scripts in layout options")
	}
	joined := strings.Join(gotOptions.Scripts, "\n")
	if !strings.Contains(joined, "pageJS()") {
		t.Error("expected page scripts in layout options")
	}
}

// TestBuildPage_LayoutEmptyFallback tests that a Layout returning "" falls
// back to the package's standalone document.
func TestBuildPage_LayoutEmptyFallback(t *testing.T) {
	opts := PageOptions{
		Title:   "Fallback",
		Content: "<p>Content</p>",
		Layout: func(w http.ResponseWriter, r *http.Request, title string, body string, options types.LayoutOptions) string {
			return ""
		},
	}

	html := buildPage(nil, nil, opts)

	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("expected default document when layout returns empty")
	}
	if !strings.Contains(html, "<title>Fallback</title>") {
		t.Error("expected default document title")
	}
}

// TestPageRender_StatusCode tests that PageRender sets the correct status code.
func TestPageRender_StatusCode(t *testing.T) {
	opts := PageOptions{
		Title:   "Test",
		Content: "<p>Content</p>",
	}

	recorder := httptest.NewRecorder()
	PageRender(recorder, newRequest(), opts)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, recorder.Code)
	}
}

// TestPageRender_ContentType tests that PageRender sets the correct content type.
func TestPageRender_ContentType(t *testing.T) {
	opts := PageOptions{
		Title:   "Test",
		Content: "<p>Content</p>",
	}

	recorder := httptest.NewRecorder()
	PageRender(recorder, newRequest(), opts)

	contentType := recorder.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("expected Content-Type 'text/html', got %q", contentType)
	}
}

// TestPageRender_HTMLOutput tests that PageRender writes HTML to the response.
func TestPageRender_HTMLOutput(t *testing.T) {
	opts := PageOptions{
		Title:   "Render Test",
		Content: "<p>Rendered Content</p>",
	}

	recorder := httptest.NewRecorder()
	PageRender(recorder, newRequest(), opts)

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
		Title:      "Test",
		Content:    "<p>Content</p>",
		Logger:     logger,
		LogMessage: "Failed to render page",
	}

	recorder := httptest.NewRecorder()
	PageRender(recorder, newRequest(), opts)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, recorder.Code)
	}
}

// TestPageRender_WithoutLogger tests that PageRender works without a logger.
func TestPageRender_WithoutLogger(t *testing.T) {
	opts := PageOptions{
		Title:   "Test",
		Content: "<p>Content</p>",
		Logger:  nil,
	}

	recorder := httptest.NewRecorder()
	PageRender(recorder, newRequest(), opts)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, recorder.Code)
	}
}

// TestPageRender_ComplexLayout tests PageRender with a host layout that
// embeds the page inside site chrome.
func TestPageRender_ComplexLayout(t *testing.T) {
	opts := PageOptions{
		Title:   "Complex Page",
		Content: "<article><h2>Article Title</h2><p>Article content</p></article>",
		Scripts: "<script>console.log('page loaded');</script>",
		Layout: func(w http.ResponseWriter, r *http.Request, title string, body string, options types.LayoutOptions) string {
			scripts := strings.Join(options.Scripts, "\n")
			return `<html><body><div class="container">
				<header><h1>Site Header</h1></header>
				<main>` + body + `</main>
				<footer><p>Site Footer</p></footer>
				` + scripts + `
			</div></body></html>`
		},
	}

	recorder := httptest.NewRecorder()
	PageRender(recorder, newRequest(), opts)

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
			Title:   page.title,
			Content: page.content,
		}

		recorder := httptest.NewRecorder()
		PageRender(recorder, newRequest(), opts)

		body := recorder.Body.String()
		if !strings.Contains(body, page.title) {
			t.Errorf("expected title %q in output", page.title)
		}
		if !strings.Contains(body, page.content) {
			t.Errorf("expected content %q in output", page.content)
		}
	}
}
