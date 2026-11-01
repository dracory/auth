package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// tempKeyStore is an in-memory temporary-key store shared by a test's
// FuncTemporaryKeyGet/Set hooks — it lets the test read back stored
// payloads (codes, tokens) the way an email recipient would.
type tempKeyStore struct {
	values map[string]string
}

func newTempKeyStore() *tempKeyStore {
	return &tempKeyStore{values: map[string]string{}}
}

func (s *tempKeyStore) get(key string) (string, error) {
	return s.values[key], nil
}

func (s *tempKeyStore) set(key string, value string, _ int) error {
	s.values[key] = value
	return nil
}

// payloadField returns a top-level string field from the JSON payload
// stored under the first key carrying the given prefix.
func (s *tempKeyStore) payloadField(t *testing.T, prefix, field string) string {
	t.Helper()
	for key, raw := range s.values {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		payload := map[string]any{}
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			t.Fatalf("payload under %q is not JSON: %v", key, err)
		}
		v, _ := payload[field].(string)
		return v
	}
	return ""
}

// tokenByPrefix returns the key suffix (nonce/token) of the first stored
// key carrying the given prefix.
func (s *tempKeyStore) tokenByPrefix(t *testing.T, prefix string) string {
	t.Helper()
	for key := range s.values {
		if strings.HasPrefix(key, prefix) {
			return strings.TrimPrefix(key, prefix)
		}
	}
	t.Fatalf("no temp key with prefix %q (have %v)", prefix, s.values)
	return ""
}

func get(mux *http.ServeMux, path string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", path, nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func postForm(mux *http.ServeMux, path string, form url.Values, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

// responseField extracts a field from the JSON body, supporting both the
// flat form and api.SuccessWithData's nested {"data": {...}} envelope.
func responseField(t *testing.T, rec *httptest.ResponseRecorder, field string) string {
	t.Helper()
	payload := map[string]any{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("response is not JSON: %v\nbody: %s", err, rec.Body.String())
	}
	if data, ok := payload["data"].(map[string]any); ok {
		if v, _ := data[field].(string); v != "" {
			return v
		}
	}
	v, _ := payload[field].(string)
	return v
}
