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

	// Tokens in URL query strings are not accepted: they end up in access
	// logs, proxy logs, browser history and Referer headers. A parameter is
	// only honoured when it comes from the request body — i.e. the request
	// is not GET/HEAD and the name does not appear in the URL query.
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		return ""
	}
	query := r.URL.Query()

	// 3. API key
	if !query.Has("api_key") {
		if apiKeyFromRequest := req.GetStringTrimmed(r, "api_key"); apiKeyFromRequest != "" {
			return apiKeyFromRequest
		}
	}

	// 4. Token
	if !query.Has("token") {
		if tokenFromRequest := req.GetStringTrimmed(r, "token"); tokenFromRequest != "" {
			return tokenFromRequest
		}
	}

	return ""
}
