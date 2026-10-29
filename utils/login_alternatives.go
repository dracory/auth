package utils

import (
	"html"
	"strings"

	"github.com/dracory/auth/types"
)

// LoginAlternatives renders the "OR — use another method" fragment shown on
// login pages when more than one method is mounted. Returns "" for fewer
// than two methods or when the current method can't be resolved.
// Labels come from Method.AlternativeLabel(); paths are derived from the
// mount rules (primary at <endpoint>/login, others at
// <endpoint>/login-<name>). Output is HTML-escaped — no user input.
func LoginAlternatives(endpoint string, currentMethod string, methods []types.Method) string {
	if len(methods) < 2 {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<div class="login-alternatives"><hr/><p>OR — use another method:</p>`)
	rendered := 0
	for i, m := range methods {
		if m == nil || m.Name() == currentMethod {
			continue
		}
		// Mount rule (see NewMultiMethodAuth): methods[0] mounts at
		// <endpoint>/login, the rest at <endpoint>/login-<name>.
		href := endpoint + "/login-" + m.Name()
		if i == 0 {
			href = endpoint + "/login"
		}
		label := m.AlternativeLabel()
		if label == "" {
			label = m.Name()
		}
		b.WriteString(`<a class="btn btn-outline-secondary" href="`)
		b.WriteString(html.EscapeString(href))
		b.WriteString(`">`)
		b.WriteString(html.EscapeString(label))
		b.WriteString(`</a>`)
		rendered++
	}
	b.WriteString(`</div>`)
	if rendered == 0 {
		return ""
	}
	return b.String()
}
