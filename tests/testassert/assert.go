package testassert

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Nil reports a test failure if value is not nil.
func Nil(t *testing.T, value interface{}) {
	t.Helper()
	if value != nil {
		t.Fatalf("expected nil, got %#v", value)
	}
}

// NotNil reports a test failure if value is nil.
func NotNil(t *testing.T, value interface{}) {
	t.Helper()
	if value == nil {
		t.Fatalf("expected non-nil value")
	}
}

// Equal reports a test failure if expected != actual (using Go's ==).
func Equal(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
}

// HTTPBodyContainsf sends a request through handler and checks that the
// response body contains the expected substring.
func HTTPBodyContainsf(t *testing.T, handler http.HandlerFunc, method, urlStr string, values url.Values, expected string, msgAndArgs ...interface{}) {
	t.Helper()

	req := newRequestForAssert(t, method, urlStr, values)

	recorder := httptest.NewRecorder()
	handler(recorder, req)

	body := recorder.Body.String()
	if !strings.Contains(body, expected) {
		_ = msgAndArgs // msgAndArgs are ignored; only format is used in original tests
		t.Fatalf("expected body to contain %q, got %q", expected, body)
	}
}

// HTTPBodyNotContainsf sends a request through handler and checks that the
// response body does not contain the expected substring.
func HTTPBodyNotContainsf(t *testing.T, handler http.HandlerFunc, method, urlStr string, values url.Values, expected string, msgAndArgs ...interface{}) {
	t.Helper()

	req := newRequestForAssert(t, method, urlStr, values)

	recorder := httptest.NewRecorder()
	handler(recorder, req)

	body := recorder.Body.String()
	if strings.Contains(body, expected) {
		_ = msgAndArgs // msgAndArgs are ignored; only format is used in original tests
		t.Fatalf("expected body not to contain %q, got %q", expected, body)
	}
}

// newRequestForAssert is a small helper used by HTTPBodyContainsf and
// HTTPBodyNotContainsf. It uses only the standard library.
func newRequestForAssert(t *testing.T, method, urlStr string, values url.Values) *http.Request {
	t.Helper()

	var bodyStr string
	if method == http.MethodPost && values != nil {
		bodyStr = values.Encode()
	}

	req, err := http.NewRequest(method, urlStr, strings.NewReader(bodyStr))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	if method == http.MethodPost && values != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// For GET requests, attach values as query parameters instead.
	if method == http.MethodGet && values != nil {
		q := req.URL.Query()
		for k, vs := range values {
			for _, v := range vs {
				q.Add(k, v)
			}
		}
		req.URL.RawQuery = q.Encode()
	}

	return req
}
