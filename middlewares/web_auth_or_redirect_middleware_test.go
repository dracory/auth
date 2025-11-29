package middlewares

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/types"
)

func TestWebAuthOrRedirectMiddleware_NoToken_RedirectsToLogin(t *testing.T) {
	authInstance := testutils.NewAuthSharedForTest()
	testutils.SetUseCookiesForTest(authInstance, true)
	testutils.SetLoginURLForTest(authInstance, "/auth/login")

	req, err := http.NewRequest("GET", "/", bytes.NewBuffer([]byte("")))
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called when token is missing")
	})

	handler := WebAuthOrRedirectMiddleware(next, authInstance)
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, status)
	}

	location := recorder.Header().Get("Location")
	expectedLocation := authInstance.LinkLogin()
	if location != expectedLocation {
		t.Fatalf("expected redirect to %q, got %q", expectedLocation, location)
	}
}

func TestWebAuthOrRedirectMiddleware_InvalidToken_RedirectsToLogin(t *testing.T) {
	authInstance := testutils.NewAuthSharedForTest()
	testutils.SetUseCookiesForTest(authInstance, true)
	testutils.SetLoginURLForTest(authInstance, "/auth/login")

	// Simulate an error in token lookup
	testutils.SetFuncUserFindByAuthTokenForTest(authInstance, func(ctx context.Context, token string, options types.UserAuthOptions) (userID string, err error) {
		return "", http.ErrNoCookie
	})

	req, err := http.NewRequest("GET", "/", bytes.NewBuffer([]byte("")))
	if err != nil {
		t.Fatal(err)
	}

	req.AddCookie(&http.Cookie{Name: types.CookieName, Value: "invalid-token"})

	recorder := httptest.NewRecorder()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called when token is invalid")
	})

	handler := WebAuthOrRedirectMiddleware(next, authInstance)
	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusTemporaryRedirect {
		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, status)
	}

	location := recorder.Header().Get("Location")
	expectedLocation := authInstance.LinkLogin()
	if location != expectedLocation {
		t.Fatalf("expected redirect to %q, got %q", expectedLocation, location)
	}
}

func TestWebAuthOrRedirectMiddleware_ValidToken_AppendsUserIDToContext(t *testing.T) {
	authInstance := testutils.NewAuthSharedForTest()
	testutils.SetUseCookiesForTest(authInstance, true)
	testutils.SetLoginURLForTest(authInstance, "/auth/login")

	// Valid token returns a userID
	testutils.SetFuncUserFindByAuthTokenForTest(authInstance, func(ctx context.Context, token string, options types.UserAuthOptions) (userID string, err error) {
		if token == "123456" {
			return "234567", nil
		}
		return "", nil
	})

	req, err := http.NewRequest("GET", "/", bytes.NewBuffer([]byte("")))
	if err != nil {
		t.Fatal(err)
	}

	req.AddCookie(&http.Cookie{Name: types.CookieName, Value: "123456"})

	recorder := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := r.Context().Value(types.AuthenticatedUserID{})
		expectedUserID := "234567"
		if value != expectedUserID {
			t.Fatalf("expected user ID %q in context, got %v", expectedUserID, value)
		}
	})

	handler := WebAuthOrRedirectMiddleware(next, authInstance)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}
