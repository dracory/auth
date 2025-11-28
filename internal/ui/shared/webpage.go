package shared

import (
	"log/slog"
	"net/http"

	"github.com/dracory/hb"
	"github.com/dracory/uncdn"
)

type PageOptions struct {
	Title      string
	Layout     func(string) string
	Content    string
	Scripts    string
	Logger     *slog.Logger
	LogMessage string
}

// buildPage composes a full HTML document using the shared UI shell
// (styles, scripts) and a caller-provided layout function that wraps the
// page-specific content.
func buildPage(opts PageOptions) string {
	// Apply outer layout to the content first.
	laidOutContent := opts.Layout(opts.Content)

	faviconImgCms := `data:image/x-icon;base64,AAABAAEAEBAQAAEABAAoAQAAFgAAACgAAAAQAAAAIAAAAAEABAAAAAAAgAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAmzKzAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABEQEAAQERAAEAAQABAAEAAQABAQEBEQABAAEREQEAAAERARARAREAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAD//wAA//8AAP//AAD//wAA//8AAP//AAD//wAAi6MAALu7AAC6owAAuC8AAIkjAAD//wAA//8AAP//AAD//wAA`
	app := ""

	webpage := hb.NewWebpage()
	webpage.SetTitle(opts.Title)
	webpage.SetFavicon(faviconImgCms)
	webpage.AddStyles([]string{
		uncdn.BootstrapCss521(),
	})
	webpage.AddScripts([]string{
		uncdn.Jquery360(),
		uncdn.BootstrapJs521(),
		uncdn.WebJs260(),
		app,
		opts.Scripts,
	})
	webpage.AddStyle(`html,body{height:100%;font-family: Ubuntu, sans-serif;}`)
	webpage.AddStyle(`body {
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
	}`)
	webpage.AddChild(hb.NewHTML(laidOutContent))

	return webpage.ToHTML()
}

// PageRender writes the provided HTML to the ResponseWriter using a standard
// status code and content type. If writing fails and a logger is provided, it
// logs the supplied error message together with the error.
func PageRender(
	w http.ResponseWriter,
	opts PageOptions,
) {
	html := buildPage(opts)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(html)); err != nil {
		if opts.Logger != nil {
			opts.Logger.Error(opts.LogMessage, "error", err)
		}
	}
}
