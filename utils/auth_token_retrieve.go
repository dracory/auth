package utils

import (
	"net/http"

	"github.com/dracory/req"
)

// AuthTokenRetrieve retrieves the auth token from the request using the default cookie name.
// Several attempts are made:
//  1. From cookie
//  2. Authorization header (aka Bearer token)
//  3. Request param "api_key"
//  4. Request param "token"
func AuthTokenRetrieve(r *http.Request, useCookies bool) string {
	return AuthTokenRetrieveWithName(r, useCookies, "")
}

// AuthTokenRetrieveWithName retrieves the auth token from the request using the provided cookie name.
// If cookieName is empty, the default types.CookieName is used.
func AuthTokenRetrieveWithName(r *http.Request, useCookies bool, cookieName string) string {
	// 1. Token from cookie
	if useCookies {
		return AuthCookieGetWithName(r, cookieName)
	}

	// 2. Bearer token
	authTokenFromBearerToken := BearerTokenFromHeader(r.Header.Get("Authorization"))

	if authTokenFromBearerToken != "" {
		return authTokenFromBearerToken
	}

	// 3. API key
	apiKeyFromRequest := req.GetStringTrimmed(r, "api_key")

	if apiKeyFromRequest != "" {
		return apiKeyFromRequest
	}

	// 4. Token
	tokenFromRequest := req.GetStringTrimmed(r, "token")

	if tokenFromRequest != "" {
		return tokenFromRequest
	}

	return ""
}
