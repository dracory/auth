package utils

import "strings"

// BearerTokenFromHeader extracts the bearer token from the
// passed authorization header value. If a bearer token
// is not found, an empty string is returned.
//
// Parameters:
//   - authHeader: a string representing the authorization header
//
// Returns:
//   - a string representing the extracted bearer token
//
// Example:
//
//	authHeader := r.Header.Get("Authorization")
//	authTokenFromBearerToken := BearerTokenFromHeader(authHeader)
//
//  or simplified
//
//	authTokenFromBearerToken := BearerTokenFromHeader(r.Header.Get("Authorization"))
func BearerTokenFromHeader(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, "Bearer")
	if len(parts) != 2 {
		return ""
	}

	token := strings.TrimSpace(parts[1])
	if len(token) < 1 {
		return ""
	}

	return token
}
