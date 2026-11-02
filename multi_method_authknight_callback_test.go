package auth

import (
	"net/http"
	"testing"

	"github.com/dracory/auth/types"
)

// TestMultiMethod_AuthKnightCallbackReachable proves the callback URL the
// package advertises to AuthKnight (LinkAuthKnightCallback ->
// <endpoint>/api/authknight/callback) is actually served by the
// multi-method mux. The mux mounts the callback handler at
// <endpoint>/authknight-callback instead, so the advertised URL 404s and
// every AuthKnight login dead-ends after the hosted redirect.
func TestMultiMethod_AuthKnightCallbackReachable(t *testing.T) {
	cfg := newValidAuthKnightConfig()
	a, err := NewAuthKnightAuth(cfg)
	if err != nil {
		t.Fatalf("NewAuthKnightAuth: %v", err)
	}

	mma, err := NewMultiMethodAuth("/auth", []types.Method{NewAuthKnightMethod(a)})
	if err != nil {
		t.Fatalf("NewMultiMethodAuth: %v", err)
	}

	callbackPath := a.LinkAuthKnightCallback()
	rec := serve(t, mma.Router(), http.MethodGet, callbackPath+"?once=TESTTOKEN")

	if rec.Code == http.StatusNotFound {
		t.Fatalf("advertised AuthKnight callback %q returns 404 on the "+
			"multi-method mux (handler is only mounted at "+
			"/auth/authknight-callback)", callbackPath)
	}
}
