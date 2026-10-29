package shared

import (
	"log/slog"
	"net/http"

	"github.com/dracory/auth/internal/helpers"
	"github.com/dracory/auth/types"
	"github.com/dracory/hb"
	"github.com/dracory/uncdn"
)

// LayoutFunc renders a complete HTML document around a page's body
// content. It matches blogadmin/shopadmin's FuncLayout signature so one
// host layout can serve all dracory packages. Returning "" falls back
// to the package's standalone document.
type LayoutFunc = func(w http.ResponseWriter, r *http.Request, title string, body string, options types.LayoutOptions) string

type PageOptions struct {
	Title      string
	Layout     LayoutFunc
	Content    string
	Scripts    string
	Logger     *slog.Logger
	LogMessage string

	// Auth is optional; when set, ThemeCSS/CustomCSS are pulled from the
	// implementation via optional interfaces (nil-safe for tests).
	Auth types.AuthSharedInterface
}

// layoutOptionsFor assembles the page-level assets a built-in auth page
// depends on — passed to a custom FuncLayout so the host document can
// inject them. Vendor assets (Bootstrap, jQuery, web.js) are NOT included:
// the document owner supplies those — host layouts bring their own, and
// defaultDocument injects the vendored copies itself. Passing vendor
// scripts here would double-load Bootstrap and break data-API components
// (each delegated click toggles twice).
func layoutOptionsFor(opts PageOptions) types.LayoutOptions {
	options := types.LayoutOptions{
		Styles: []string{
			baseStyles,
		},
	}
	if theme := themeCSSOf(opts.Auth); theme != "" {
		options.StyleURLs = append(options.StyleURLs, theme)
	}
	if custom := customCSSOf(opts.Auth); custom != "" {
		options.Styles = append(options.Styles, custom)
	}
	if opts.Scripts != "" {
		options.Scripts = append(options.Scripts, opts.Scripts)
	}
	return options
}

// buildPage composes the page body (default padded-section decoration
// around the card content), then either delegates the full document to a
// host-provided layout or falls back to the standalone shell.
func buildPage(w http.ResponseWriter, r *http.Request, opts PageOptions) string {
	body := helpers.Layout(opts.Content)
	options := layoutOptionsFor(opts)

	if opts.Layout != nil {
		if html := opts.Layout(w, r, opts.Title, body, options); html != "" {
			return html
		}
	}

	return defaultDocument(opts.Title, body, options)
}

// defaultDocument renders the package's standalone HTML shell — used
// when no FuncLayout is configured or it declines by returning "".
func defaultDocument(title string, body string, options types.LayoutOptions) string {
	faviconImgCms := `data:image/x-icon;base64,AAABAAEAEBAQAAEABAAoAQAAFgAAACgAAAAQAAAAIAAAAAEABAAAAAAAgAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAmzKzAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABEQEAAQERAAEAAQABAAEAAQABAQEBEQABAAEREQEAAAERARARAREAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD//wAA//8AAP//AAD//wAA//8AAP//AAD//wAAi6MAALu7AAC6owAAuC8AAIkjAAD//wAA//8AAP//AAD//wAA`

	webpage := hb.NewWebpage()
	webpage.SetTitle(title)
	webpage.SetFavicon(faviconImgCms)
	webpage.AddStyles(append([]string{uncdn.BootstrapCss521()}, options.Styles...))
	webpage.AddStyleURLs(options.StyleURLs)
	webpage.AddScripts(append([]string{
		uncdn.Jquery360(),
		uncdn.BootstrapJs521(),
		uncdn.WebJs260(),
	}, options.Scripts...))
	webpage.AddScriptURLs(options.ScriptURLs)
	webpage.AddChild(hb.NewHTML(body))

	return webpage.ToHTML()
}

// PageRender writes the provided HTML to the ResponseWriter using a standard
// status code and content type. If writing fails and a logger is provided, it
// logs the supplied error message together with the error.
func PageRender(
	w http.ResponseWriter,
	r *http.Request,
	opts PageOptions,
) {
	html := buildPage(w, r, opts)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(html)); err != nil {
		if opts.Logger != nil {
			opts.Logger.Error(opts.LogMessage, "error", err)
		}
	}
}

const baseStyles = `html,body{height:100%;font-family: Ubuntu, sans-serif;}
body {
	font-family: "Nunito", sans-serif;
	font-size: 0.9rem;
	font-weight: 400;
	line-height: 1.6;
	color: #212529;
	text-align: left;
	background-color: #f8fafc;
}
.form-select {
	display: block;
	width: 100%;
	padding: .375rem 2.25rem .375rem .75rem;
	font-size: 1rem;
	font-weight: 400;
	line-height: 1.5;
	color: #212529;
	background-color: #fff;
	background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 16 16'%3e%3cpath fill='none' stroke='%23343a40' stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M2 5l6 6 6-6'/%3e%3c/svg%3e");
	background-repeat: no-repeat;
	background-position: right .75rem center;
	background-size: 16px 12px;
	border: 1px solid #ced4da;
	border-radius: .25rem;
	-webkit-appearance: none;
	-moz-appearance: none;
	appearance: none;
}`

func themeCSSOf(a types.AuthSharedInterface) string {
	if t, ok := a.(interface{ GetThemeCSS() string }); ok {
		return t.GetThemeCSS()
	}
	return ""
}

func customCSSOf(a types.AuthSharedInterface) string {
	if t, ok := a.(interface{ GetCustomCSS() string }); ok {
		return t.GetCustomCSS()
	}
	return ""
}

type authKnightChecker interface {
	IsAuthKnight() bool
	LinkAuthKnightRedirect(r *http.Request) string
}

// RedirectAuthKnightIfActive checks if the auth instance is in AuthKnight mode
// and redirects to AuthKnight's hosted login if so. Returns true if a redirect
// was performed, false otherwise.
func RedirectAuthKnightIfActive(w http.ResponseWriter, r *http.Request, a types.AuthSharedInterface) bool {
	checker, ok := a.(authKnightChecker)
	if !ok || !checker.IsAuthKnight() {
		return false
	}
	http.Redirect(w, r, checker.LinkAuthKnightRedirect(r), http.StatusTemporaryRedirect)
	return true
}
