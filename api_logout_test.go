package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth/tests/testassert"
)

func TestLogoutEndpointNoToken(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	// No cookie or header set, so token retrieval should fail/return empty
	expectedSuccess := `"status":"success"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogout(), url.Values{}, expectedSuccess, "%")

	expectedMessage := `"message":"logout success"`
	testassert.HTTPBodyContainsf(t, authInstance.Router().ServeHTTP, "POST", authInstance.LinkApiLogout(), url.Values{}, expectedMessage, "%")
}

func TestLogoutEndpointTokenValidationError(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	testassert.Nil(t, err)
	testassert.NotNil(t, authInstance)

	// Mock token validation error
	authInstance.funcUserFindByAuthToken = func(token string, options UserAuthOptions) (userID string, err error) {
		return "", errors.New("db error")
	}

	// Manually create request to set cookie
	req, _ := http.NewRequest("POST", authInstance.LinkApiLogout(), nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: "invalid-token"})

	// We need to use a custom handler wrapper to inject the cookie because testassert helper creates a new request
	// But since testassert doesn't support custom requests easily, we might need to rely on header or just assume testassert helper can be bypassed or we use a different approach.
	// Actually, `testassert.HTTPBodyContainsf` creates its own request.
	// Let's look at `testassert` again. It doesn't seem to allow setting cookies easily.
	// However, `AuthTokenRetrieve` checks header "Authorization: Bearer <token>" too.
	// `testassert` doesn't seem to allow setting headers either in `HTTPBodyContainsf`.

	// Wait, `testassert` is very simple. I might need to write a custom test for this one or modify `testassert`?
	// Or I can just write the test using standard `httptest` here since I need custom headers/cookies.

	// Let's use standard httptest for this specific test to ensure we can set the token.
}

// Re-writing the file content to use standard httptest where needed or if I can't use testassert.
// Actually, I can just write standard Go tests using `httptest` and `testassert`'s assertions where applicable, or just standard `t.Fatal`.
// Let's try to stick to the pattern but use `httptest` directly for setup where `testassert` helper is insufficient.

func TestLogoutEndpointTokenValidationError_Custom(t *testing.T) {
	authInstance, err := testSetupUsernameAndPasswordAuth()
	testassert.Nil(t, err)

	authInstance.funcUserFindByAuthToken = func(token string, options UserAuthOptions) (userID string, err error) {
		return "", errors.New("db error")
	}

	req, _ := http.NewRequest("POST", authInstance.LinkApiLogout(), nil)
	req.Header.Set("Authorization", "Bearer some-token")

	recorder := httptest.NewRecorder()
	authInstance.Router().ServeHTTP(recorder, req)

	body := recorder.Body.String()
	expected := `"message":"logout failed"`
	if !strings.Contains(body, expected) {
		t.Fatalf("expected body to contain %q, got %q", expected, body)
	}
}
