package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dracory/auth/internal/helpers"
	"github.com/dracory/auth/internal/middlewares"
	"github.com/dracory/req"
	"github.com/dracory/str"
)

func (a authImplementation) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.AuthHandler(w, r)
	})
}

func (a authImplementation) Router() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.AuthHandler)
	return mux
}

// uriHasPathSuffix reports whether uri ends with pathSuffix at a path
// boundary, i.e. uri is exactly pathSuffix or ends with "/"+pathSuffix.
// This avoids false positives such as "/mylogin" matching the "login" path.
func uriHasPathSuffix(uri string, pathSuffix string) bool {
	return uri == pathSuffix || strings.HasSuffix(uri, "/"+pathSuffix)
}

// Router routes the requests
func (a authImplementation) AuthHandler(w http.ResponseWriter, r *http.Request) {
	path := req.GetStringOr(r, "path", "home")
	uri := r.RequestURI

	if r.RequestURI == "" && r.URL.Path != "" {
		uri = r.URL.Path // Attempt to take from URL path (empty RequestURI occurs during testing)
	}

	uri = strings.TrimSuffix(uri, "/") // Remove trailing slash

	originalURI := uri

	if strings.Contains(uri, "?") {
		uri = str.LeftFrom(uri, "?")
	}

	fmt.Printf("[AUTHHANDLER] originalURI='%s' uri='%s' path='%s'\n", originalURI, uri, path)

	if uriHasPathSuffix(uri, PathApiLogin) {
		path = PathApiLogin
	} else if uriHasPathSuffix(uri, PathApiLoginCodeVerify) {
		path = PathApiLoginCodeVerify
	} else if uriHasPathSuffix(uri, PathApiLogout) {
		path = PathApiLogout
	} else if uriHasPathSuffix(uri, PathApiResetPassword) {
		path = PathApiResetPassword
	} else if uriHasPathSuffix(uri, PathApiRestorePassword) {
		path = PathApiRestorePassword
	} else if uriHasPathSuffix(uri, PathApiRegister) {
		path = PathApiRegister
	} else if uriHasPathSuffix(uri, PathApiRegisterCodeVerify) {
		path = PathApiRegisterCodeVerify
	} else if uriHasPathSuffix(uri, PathLogin) {
		path = PathLogin
	} else if uriHasPathSuffix(uri, PathLoginCodeVerify) {
		path = PathLoginCodeVerify
	} else if uriHasPathSuffix(uri, PathLogout) {
		path = PathLogout
	} else if uriHasPathSuffix(uri, PathRegister) {
		path = PathRegister
	} else if uriHasPathSuffix(uri, PathRegisterCodeVerify) {
		path = PathRegisterCodeVerify
	} else if uriHasPathSuffix(uri, PathPasswordRestore) {
		path = PathPasswordRestore
	} else if uriHasPathSuffix(uri, PathPasswordReset) {
		path = PathPasswordReset
	} else if uriHasPathSuffix(uri, PathApiImpersonateStart) {
		path = PathApiImpersonateStart
	} else if uriHasPathSuffix(uri, PathApiImpersonateStop) {
		path = PathApiImpersonateStop
	} else if uriHasPathSuffix(uri, PathApiAuthKnightCallback) {
		path = PathApiAuthKnightCallback
	}

	ctx := context.WithValue(r.Context(), keyEndpoint, r.URL.Path)

	routeFunc := a.getRoute(path)

	fmt.Printf("[AUTHHANDLER] resolved path='%s' routeFunc!=nil=%v\n", path, routeFunc != nil)

	routeFunc(w, r.WithContext(ctx))
}

// getRoute finds a route
func (a authImplementation) getRoute(route string) func(w http.ResponseWriter, r *http.Request) {
	csrfCfg := middlewares.CSRFConfig{
		Enabled:  a.enableCSRFProtection,
		Validate: a.funcCSRFTokenValidate,
	}

	routes := map[string]func(w http.ResponseWriter, r *http.Request){
		PathApiLogout:       a.apiLogout,
		PathLogin:           a.pageLogin,
		PathLoginCodeVerify: a.pageLoginCodeVerify,
		PathLogout:          a.pageLogout,
		PathPasswordReset:   a.pagePasswordReset,
		PathPasswordRestore: a.pagePasswordRestore,
	}

	for path, handler := range a.buildAPIRoutes(csrfCfg) {
		routes[path] = handler
	}

	if a.enableRegistration {
		routes[PathRegister] = a.pageRegister
		routes[PathRegisterCodeVerify] = a.pageRegisterCodeVerify
	}

	if val, ok := routes[route]; ok {
		return val
	}

	return a.notFoundHandler
}

func (a authImplementation) buildAPIRoutes(csrfCfg middlewares.CSRFConfig) map[string]func(http.ResponseWriter, *http.Request) {
	routes := make(map[string]func(http.ResponseWriter, *http.Request))

	apiRoutes := []struct {
		path     string
		endpoint string
		handler  func(http.ResponseWriter, *http.Request)
		useCSRF  bool
	}{
		{PathApiLogin, EndpointLogin, a.apiLogin, true},
		{PathApiLoginCodeVerify, EndpointLoginCodeVerify, a.apiLoginCodeVerify, false},
		{PathApiRegister, EndpointRegister, a.apiRegister, true},
		{PathApiRegisterCodeVerify, EndpointRegisterCodeVerify, a.apiRegisterCodeVerify, false},
		{PathApiResetPassword, EndpointPasswordReset, a.apiPasswordReset, true},
		{PathApiRestorePassword, EndpointPasswordRestore, a.apiPasswordRestore, false},
	}

	if a.enableImpersonation {
		apiRoutes = append(apiRoutes, []struct {
			path     string
			endpoint string
			handler  func(http.ResponseWriter, *http.Request)
			useCSRF  bool
		}{
			{PathApiImpersonateStart, EndpointImpersonateStart, a.apiImpersonateStart, true},
			{PathApiImpersonateStop, EndpointImpersonateStop, a.apiImpersonateStop, true},
		}...)
	}

	if a.authKnight {
		apiRoutes = append(apiRoutes, []struct {
			path     string
			endpoint string
			handler  func(http.ResponseWriter, *http.Request)
			useCSRF  bool
		}{
			{PathApiAuthKnightCallback, EndpointAuthKnightCallback, a.apiAuthKnightCallback, false},
		}...)
	}

	for _, cfg := range apiRoutes {
		h := cfg.handler
		if cfg.useCSRF {
			h = middlewares.WithCSRF(csrfCfg, h)
		}

		routes[cfg.path] = middlewares.WithRateLimit(
			middlewares.RateLimitConfig{
				Check: func(w http.ResponseWriter, r *http.Request, endpoint string) bool {
					return helpers.CheckRateLimit(w, r, endpoint, a.disableRateLimit, a.funcCheckRateLimit, a.rateLimiter, a.GetObservabilityHooks())
				},
				Endpoint: cfg.endpoint,
			},
			h,
		)
	}

	return routes
}

func (a authImplementation) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, a.LinkLogin(), http.StatusTemporaryRedirect)
}
