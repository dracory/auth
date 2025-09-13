package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/dracory/req"
	"github.com/dracory/str"
)

func (a Auth) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.AuthHandler(w, r)
	})
}

func (a Auth) Router() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.AuthHandler)
	return mux
}

// Router routes the requests
func (a Auth) AuthHandler(w http.ResponseWriter, r *http.Request) {
	path := req.GetStringOr(r, "path", "home")
	uri := r.RequestURI

	if r.RequestURI == "" && r.URL.Path != "" {
		uri = r.URL.Path // Attempt to take from URL path (empty RequestURI occurs during testing)
	}

	uri = strings.TrimSuffix(uri, "/") // Remove trailing slash

	if strings.Contains(uri, "?") {
		uri = str.LeftFrom(uri, "?")
	}

	if strings.HasSuffix(uri, PathApiLogin) {
		path = PathApiLogin
	} else if strings.HasSuffix(uri, PathApiLoginCodeVerify) {
		path = PathApiLoginCodeVerify
	} else if strings.HasSuffix(uri, PathApiLogout) {
		path = PathApiLogout
	} else if strings.HasSuffix(uri, PathApiResetPassword) {
		path = PathApiResetPassword
	} else if strings.HasSuffix(uri, PathApiRestorePassword) {
		path = PathApiRestorePassword
	} else if strings.HasSuffix(uri, PathApiRegister) {
		path = PathApiRegister
	} else if strings.HasSuffix(uri, PathApiRegisterCodeVerify) {
		path = PathApiRegisterCodeVerify
	} else if strings.HasSuffix(uri, PathLogin) {
		path = PathLogin
	} else if strings.HasSuffix(uri, PathLoginCodeVerify) {
		path = PathLoginCodeVerify
	} else if strings.HasSuffix(uri, PathLogout) {
		path = PathLogout
	} else if strings.HasSuffix(uri, PathRegister) {
		path = PathRegister
	} else if strings.HasSuffix(uri, PathRegisterCodeVerify) {
		path = PathRegisterCodeVerify
	} else if strings.HasSuffix(uri, PathPasswordRestore) {
		path = PathPasswordRestore
	} else if strings.HasSuffix(uri, PathPasswordReset) {
		path = PathPasswordReset
	}

	ctx := context.WithValue(r.Context(), keyEndpoint, r.URL.Path)

	routeFunc := a.getRoute(path)

	routeFunc(w, r.WithContext(ctx))
}

// getRoute finds a route
func (a Auth) getRoute(route string) func(w http.ResponseWriter, r *http.Request) {
	routes := map[string]func(w http.ResponseWriter, r *http.Request){
		PathApiLogin:              a.apiLogin,
		PathApiLoginCodeVerify:    a.apiLoginCodeVerify,
		PathApiLogout:             a.apiLogout,
		PathApiRegister:           a.apiRegister,
		PathApiRegisterCodeVerify: a.apiRegisterCodeVerify,
		PathApiResetPassword:      a.apiPasswordReset,
		PathApiRestorePassword:    a.apiPasswordRestore,
		PathLogin:                 a.pageLogin,
		PathLoginCodeVerify:       a.pageLoginCodeVerify,
		PathLogout:                a.pageLogout,
		PathPasswordReset:         a.pagePasswordReset,
		PathPasswordRestore:       a.pagePasswordRestore,
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

func (a Auth) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, a.LinkLogin(), http.StatusTemporaryRedirect)
}
