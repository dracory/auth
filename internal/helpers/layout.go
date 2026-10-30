package helpers

import "github.com/dracory/hb"

// Layout wraps a built-in auth page's card content in the shared outer
// shell — a vertically centered, narrow column on a light background.
// Stock Bootstrap classes only, so consumer themes restyle it cleanly.
func Layout(content string) string {
	return hb.NewDiv().Class("min-vh-100 d-flex align-items-center justify-content-center bg-light py-5").
		Child(hb.NewDiv().Class("container").
			Child(hb.NewDiv().Class("row justify-content-center").
				Child(hb.NewDiv().Class("col-12 col-sm-10 col-md-8 col-lg-5 col-xl-4").
					HTML(content)))).
		ToHTML()
}
