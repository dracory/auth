package utils

import (
	"strings"
)

// MagicLinkKeyPrefix prefixes the temporary keys under which magic-link
// payloads are stored. The key IS the token — it is 128+ bits of entropy,
// single-use, and expires with the payload.
const MagicLinkKeyPrefix = "magiclink:token:"

// MagicLinkExpiresSeconds is the default magic-link TTL (15 minutes).
const MagicLinkExpiresSeconds = 900

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
// would escape the origin).
func IsSafeReturnPath(p string) bool {
	return strings.HasPrefix(p, "/") && !strings.HasPrefix(p, "//")
}
