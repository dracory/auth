package auth

import "github.com/dracory/hb"

func (a Auth) layout(content string) string {
	font := hb.NewStyleURL("https://fonts.bunny.net/css?family=Nunito").ToHTML()
	style := hb.NewStyle(`
	html, body {
		background: #f8fafc;
		font-family: Nunito, sans-serif;;
	}
	`).ToHTML()
	h := hb.NewSection().
		Style("padding:120px 0px").
		AddChild(hb.NewHTML(content))
	return font + style + h.ToHTML()
}
