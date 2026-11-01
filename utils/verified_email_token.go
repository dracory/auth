package utils

import "github.com/dracory/str"

// VerifiedEmailTokenExpiresSeconds is the TTL of a verified-email token —
// long enough to complete the registration form, short enough to limit
// token theft windows.
const VerifiedEmailTokenExpiresSeconds = 900

// IssueVerifiedEmailToken stores a fresh random token mapped to the
// verified email in the temporary key store. The token is the unified
// "verified-email credential" every register-first flow (OTP code verify,
// magic link, AuthKnight ak_key) exchanges its proof for — the register
// page resolves it back to the email, renders it read-only, and the
// register API consumes it on submit (single use).
func IssueVerifiedEmailToken(
	tempKeySet func(key string, value string, expiresSeconds int) error,
	email string,
	expiresSeconds int,
) (token string, err error) {
	token, err = str.RandomFromGamma(32, "BCDFGHJKLMNPQRSTVXYZ")
	if err != nil {
		return "", err
	}
	if expiresSeconds <= 0 {
		expiresSeconds = VerifiedEmailTokenExpiresSeconds
	}
	if err := tempKeySet(token, email, expiresSeconds); err != nil {
		return "", err
	}
	return token, nil
}
