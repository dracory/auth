package utils

import "strings"

// BearerTokenFromHeader extracts the token from an Authorization header of
// the form "Bearer <token>" (scheme matched case-insensitively per
// RFC 9110 §11). Returns "" for any other scheme or malformed value.
func BearerTokenFromHeader(authHeader string) string {
	scheme, token, found := strings.Cut(strings.TrimSpace(authHeader), " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}

	token = strings.TrimSpace(token)
	if token == "" || strings.ContainsAny(token, " \t") {
		return ""
	}

	return token
}
