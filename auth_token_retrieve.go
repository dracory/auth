package auth

// import (
// 	"net/http"

// 	"github.com/dracory/auth/utils"
// 	"github.com/dracory/req"
// )

// // authTokenRetrieve retrieves the auth token from the request
// // Several attempts are made:
// //  1. From cookie
// //  2. Authorization header (aka Bearer token)
// //  3. Request param "api_key"
// //  4. Request param "token"
// func AuthTokenRetrieve(r *http.Request, useCookies bool) string {
// 	// 1. Token from cookie
// 	if useCookies {
// 		return utils.AuthCookieGet(r)
// 	}

// 	// 2. Bearer token
// 	authTokenFromBearerToken := utils.BearerTokenFromHeader(r.Header.Get("Authorization"))

// 	if authTokenFromBearerToken != "" {
// 		return authTokenFromBearerToken
// 	}

// 	// 3. API key
// 	apiKeyFromRequest := req.GetStringTrimmed(r, "api_key")

// 	if apiKeyFromRequest != "" {
// 		return apiKeyFromRequest
// 	}

// 	// 4. Token
// 	tokenFromRequest := req.GetStringTrimmed(r, "token")

// 	if tokenFromRequest != "" {
// 		return tokenFromRequest
// 	}

// 	return ""
// }
