package types

// LinkTarget selects how a legal link opens.
type LinkTarget string

const (
	// LinkTargetSameTab navigates the current tab (default).
	LinkTargetSameTab LinkTarget = "_self"
	// LinkTargetNewTab opens the link in a new tab (rel="noopener").
	LinkTargetNewTab LinkTarget = "_blank"
	// LinkTargetModal opens the link in an iframe overlay. Caveat: the
	// target page must permit framing — the host's X-Frame-Options
	// header and CSP frame-ancestors directive can block the iframe, so
	// consumers choosing modal must serve frameable legal pages.
	LinkTargetModal LinkTarget = "modal"
)

// LegalLink configures a Terms/Privacy link: label, destination, and
// how it opens. Reusable for future link slots (footer links,
// alternatives).
type LegalLink struct {
	// Label is the anchor text. Defaults to "Terms" / "Privacy Policy"
	// depending on the slot when empty.
	Label string
	// URL is the link destination. Empty hides the clause entirely.
	URL string
	// Target selects how the link opens. Empty defaults to
	// LinkTargetSameTab; unrecognized values behave like
	// LinkTargetSameTab.
	Target LinkTarget
}

// LegalLinks carries the site-level legal links rendered as the "By
// continuing, you agree to {Terms} and {Privacy Policy}." line under the
// form on login/register-type built-in pages. Identical on every page,
// so they live globally on ConfigShared rather than per page.
type LegalLinks struct {
	Terms   LegalLink
	Privacy LegalLink
}
