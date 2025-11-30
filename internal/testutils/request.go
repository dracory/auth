package testutils

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// MakePostRequest is a small helper for tests that need to construct a
// POST request with form values and an httptest.ResponseRecorder.
func MakePostRequest(t *testing.T, path string, values url.Values) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()
	body := strings.NewReader(values.Encode())
	req, err := http.NewRequest(http.MethodPost, path, body)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	return recorder, req
}
