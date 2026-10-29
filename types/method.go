package types

import "net/http"

// Method is the open-extensibility contract for login methods. Built-ins
// (otp, magiclink, password, authknight) are just implementations of this
// interface — projects can register TOTP, WebAuthn, OAuth, etc. methods
// without forking the library.
type Method interface {
	// Name returns the method slug used in routes and alternative-method
	// links, e.g. "otp", "magiclink", "password".
	Name() string

	// AlternativeLabel returns the label rendered by the
	// login-alternatives fragment ("use another method").
	AlternativeLabel() string

	// PageHandler serves the GET login page for the method.
	PageHandler() http.HandlerFunc

	// AjaxHandler serves the POST handler (page ajax / JSON API). May be
	// nil — authknight has no POST handler.
	AjaxHandler() http.HandlerFunc

	// CallbackRoute returns a path and handler for methods with a callback
	// (magiclink, authknight). ok=false means no callback. The path may
	// embed an HTTP method ("POST /x-callback") — defaults to GET.
	CallbackRoute() (path string, handler http.HandlerFunc, ok bool)
}

// RouteProvider is an optional interface methods can implement to expose
// additional routes when they are mounted — e.g. the password method
// mounts recovery routes only when it is present in the method list.
type RouteProvider interface {
	// ExtraRoutes returns additional path→handler mappings mounted under
	// the auth endpoint, only when this method is in the list.
	ExtraRoutes() map[string]http.HandlerFunc
}
