package utils

import (
	"net/url"
	"strings"
)

// MagicLinkPayload is the JSON value stored under MagicLinkKeyPrefix+token.
// Email is stored in the value (the key is the random token, not PII);
// Token is duplicated into the value for constant-time comparison at verify
// time; IP optionally binds the link to the requester's IP; Return is a
// safe-relative post-login path; Remember marks remember-me intent.
type MagicLinkPayload struct {
	Email    string `json:"email"`
	IP       string `json:"ip"`
	Token    string `json:"token"`
	Return   string `json:"return,omitempty"`
	Remember bool   `json:"remember,omitempty"`
}

// IsSafeReturnPath reports whether p is a safe post-auth redirect target:
// a relative path starting with "/" but not "//" (protocol-relative URLs
// would escape the origin). Backslashes and control characters are
// rejected — several browsers treat "/\evil.com" as protocol-relative.
func IsSafeReturnPath(p string) bool {
	if !strings.HasPrefix(p, "/") || strings.HasPrefix(p, "//") {
		return false
	}
	return !strings.ContainsAny(p, "\\\r\n\t")
}

// IsSafeBackURL reports whether backURL is safe to redirect to: a relative
// path (not "//..."), or an absolute URL whose host equals the request's
// host. Parsed with net/url — substring matching would admit
// "://example.com.evil.com", userinfo tricks, and query-string smuggling.
func IsSafeBackURL(backURL string, host string) bool {
	if backURL == "" {
		return false
	}
	if strings.HasPrefix(backURL, "/") {
		return IsSafeReturnPath(backURL)
	}
	u, err := url.Parse(backURL)
	if err != nil || u.Hostname() == "" {
		return false
	}
	return strings.EqualFold(u.Hostname(), host)
}
