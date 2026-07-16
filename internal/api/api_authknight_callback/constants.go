package api_authknight_callback

import "time"

const (
	DefaultHTTPTimeout       = 10 * time.Second
	DefaultTempKeyExpiration = 1 * time.Hour
	TokenGamma               = "BCDFGHJKLMNPQRSTVXYZ"
	TokenLength              = 32
	AuthKnightWhoEndpoint    = "/api/who"
	AuthKnightDefaultBaseURL = "https://authknight.com"
)
