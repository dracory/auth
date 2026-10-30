package shared

import (
	"strings"

	"github.com/dracory/hb"
)

// The v3 built-in pages use stock Bootstrap 5 markup only — no custom CSS —
// so consumers can re-skin everything with a plain Bootstrap theme (e.g.
// Bootlegance) or their own stylesheet. The builders below keep the markup
// uniform across pages: title + subtitle above a centered shadow card,
// uppercase small labels, lg inputs/buttons, an "or" divider before
// alternative-method links, and centered muted footer links.

// PageHeader renders the page title and muted helper line shown above the
// card on every built-in page.
func PageHeader(title, subtitle string) hb.TagInterface {
	return hb.NewDiv().Children([]hb.TagInterface{
		hb.NewHeading1().Class("h3 fw-bold mb-1").Text(title),
		hb.NewParagraph().Class("text-muted mb-4").Text(subtitle),
	})
}

// AlertGroup returns the success/danger alert pair the page scripts
// address via $('div.alert-success') / $('div.alert-danger').
func AlertGroup() hb.TagInterface {
	return hb.NewDiv().Class("alert-group").Children([]hb.TagInterface{
		hb.NewDiv().Class("alert alert-success").Style("display:none"),
		hb.NewDiv().Class("alert alert-danger").Style("display:none"),
	})
}

// Card wraps children in the shared auth card (subtle shadow, padded body).
func Card(children ...hb.TagInterface) hb.TagInterface {
	return hb.NewDiv().Class("card shadow-sm").
		Child(hb.NewDiv().Class("card-body p-4").Children(children))
}

// FieldGroup renders an uppercase small label above a form control.
func FieldGroup(label string, input hb.TagInterface) hb.TagInterface {
	return hb.NewDiv().Class("mb-3").Children([]hb.TagInterface{
		hb.NewLabel().Class("form-label small fw-bold text-uppercase text-secondary").Text(label),
		input,
	})
}

// NameRow renders first/last name inputs side by side. required adds the
// HTML5 required attribute — the registration flows enforce names client-
// side, password restore sends them but does not gate on them.
func NameRow(required bool) hb.TagInterface {
	nameField := func(label, name, placeholder string) hb.TagInterface {
		input := hb.NewInput().Class("form-control form-control-lg").Name(name).Placeholder(placeholder)
		if required {
			input.Attr("required", "required")
		}
		return hb.NewDiv().Class("col-6").Child(hb.NewDiv().Class("mb-0").Children([]hb.TagInterface{
			hb.NewLabel().Class("form-label small fw-bold text-uppercase text-secondary").Text(label),
			input,
		}))
	}
	return hb.NewDiv().Class("row g-3 mb-3").Children([]hb.TagInterface{
		nameField("First name", "first_name", "Your first name"),
		nameField("Last name", "last_name", "Your last name"),
	})
}

// PrimaryButton renders the main submit button. extraClass carries the
// legacy marker classes the page scripts rely on (e.g. "ButtonLogin",
// "ButtonContinue").
func PrimaryButton(extraClass, icon, label, onClick string) hb.TagInterface {
	return hb.NewButton().
		Class(strings.TrimSpace(extraClass + " btn btn-primary btn-lg w-100 fw-bold mt-2")).
		Children([]hb.TagInterface{
			hb.NewI().Class("bi " + icon).Class("me-2"),
			hb.NewSpan().Text(label),
			hb.NewDiv().Class("ImgLoading spinner-border spinner-border-sm text-light").Style("display:none;margin-left:10px;"),
		}).
		OnClick(onClick)
}

// OrDivider renders the muted "or" separator used before alternative
// login-method links.
func OrDivider() hb.TagInterface {
	return hb.NewDiv().Class("d-flex align-items-center gap-3 my-4 text-secondary small text-uppercase").Children([]hb.TagInterface{
		hb.NewDiv().Class("flex-grow-1 border-top"),
		hb.NewSpan().Text("or"),
		hb.NewDiv().Class("flex-grow-1 border-top"),
	})
}

// AltLink renders an outlined full-width link button used for sibling
// login methods.
func AltLink(href, icon, label string) hb.TagInterface {
	link := hb.NewHyperlink().Class("btn btn-outline-secondary btn-lg w-100 fw-bold").Href(href)
	children := []hb.TagInterface{}
	if icon != "" {
		children = append(children, hb.NewI().Class("bi "+icon).Class("me-2"))
	}
	children = append(children, hb.NewSpan().Text(label))
	return link.Children(children)
}

// FooterText renders a centered muted "text link" line under the card
// (e.g. "Don't have an account? Register").
func FooterText(text, href, label string) hb.TagInterface {
	return hb.NewDiv().Class("text-muted small text-center mt-4").Children([]hb.TagInterface{
		hb.NewSpan().Text(text + " "),
		hb.NewHyperlink().Class("fw-bold text-decoration-none").Href(href).Text(label),
	})
}
