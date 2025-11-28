package auth

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebAuthOrRedirectMiddleware_NoToken_RedirectsToLogin(t *testing.T) {
	authInstance, err := newPasswordlessAuthForMiddlewareTests(true)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/", bytes.NewBuffer([]byte("")))
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called when token is missing")
	})

	handler := authInstance.WebAuthOrRedirectMiddleware(next)
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
	authInstance, err := newPasswordlessAuthForMiddlewareTests(true)
	if err != nil {
		t.Fatal(err)
	}

	// Simulate an error in token lookup
	authInstance.funcUserFindByAuthToken = func(ctx context.Context, token string, options UserAuthOptions) (userID string, err error) {
		return "", http.ErrNoCookie
	}

	req, err := http.NewRequest("GET", "/", bytes.NewBuffer([]byte("")))
	if err != nil {
		t.Fatal(err)
	}

	req.AddCookie(&http.Cookie{Name: CookieName, Value: "invalid-token"})

	recorder := httptest.NewRecorder()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called when token is invalid")
	})

	handler := authInstance.WebAuthOrRedirectMiddleware(next)
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
	authInstance, err := newPasswordlessAuthForMiddlewareTests(true)
	if err != nil {
		t.Fatal(err)
	}

	// Valid token returns a userID
	authInstance.funcUserFindByAuthToken = func(ctx context.Context, token string, options UserAuthOptions) (userID string, err error) {
		if token == "123456" {
			return "234567", nil
		}
		return "", nil
	}

	req, err := http.NewRequest("GET", "/", bytes.NewBuffer([]byte("")))
	if err != nil {
		t.Fatal(err)
	}

	req.AddCookie(&http.Cookie{Name: CookieName, Value: "123456"})

	recorder := httptest.NewRecorder()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := r.Context().Value(AuthenticatedUserID{})
		expectedUserID := "234567"
		if value != expectedUserID {
			t.Fatalf("expected user ID %q in context, got %v", expectedUserID, value)
		}
	})

	handler := authInstance.WebAuthOrRedirectMiddleware(next)
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}
}
