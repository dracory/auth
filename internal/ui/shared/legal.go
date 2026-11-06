package shared

import (
	"github.com/dracory/auth/types"
	"github.com/dracory/hb"
)

// LegalLine renders the site-level "By continuing, you agree to {Terms}
// and {Privacy Policy}." line appended inside the form card on
// login/register-type built-in pages. Returns "" when both links are
// unconfigured (empty URL). A single configured link renders a
// single-link sentence. LinkTargetModal links carry data-legal-modal —
// legal_modal.js intercepts the click and opens an iframe overlay
// (requires the target page to permit framing).
func LegalLine(links types.LegalLinks) string {
	terms := legalAnchor(links.Terms, "Terms")
	privacy := legalAnchor(links.Privacy, "Privacy Policy")

	sentence := ""
	switch {
	case terms != "" && privacy != "":
		sentence = "By continuing, you agree to " + terms + " and " + privacy + "."
	case terms != "":
		sentence = "By continuing, you agree to " + terms + "."
	case privacy != "":
		sentence = "By continuing, you agree to " + privacy + "."
	default:
		return ""
	}

	return hb.NewDiv().Class("text-muted small text-center mt-3").
		HTML(sentence).ToHTML()
}

// legalAnchor renders one link, or "" when unconfigured.
func legalAnchor(link types.LegalLink, defaultLabel string) string {
	if link.URL == "" {
		return ""
	}
	a := hb.NewHyperlink().Href(link.URL).
		Class("text-decoration-none fw-bold").
		Text(OrDefault(link.Label, defaultLabel))
	switch link.Target {
	case types.LinkTargetNewTab:
		a.Target("_blank").Attr("rel", "noopener")
	case types.LinkTargetModal:
		a.Attr("data-legal-modal", "true")
	}
	return a.ToHTML()
}
