package helpers

import "github.com/dracory/hb"

// Layout builds the default HTML layout for auth pages.
func Layout(content string) string {
	style := hb.NewStyle(`
	html, body {
		background: #f8fafc;
		font-family: Nunito, sans-serif;
	}
	`).ToHTML()
	h := hb.NewSection().
		Style("padding:120px 0px").
		AddChild(hb.NewHTML(content))
	return style + h.ToHTML()
}
