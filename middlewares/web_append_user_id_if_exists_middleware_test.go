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

func TestWebAppendUserIDIfExistsMiddleware(t *testing.T) {
	auth := testutils.NewAuthSharedForTest()
	testutils.SetUseCookiesForTest(auth, true)

	// Configure token lookup to return a user for the expected token.
	testutils.SetFuncUserFindByAuthTokenForTest(auth, func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
		if sessionID == "123456" {
			return "234567", nil
		}
		return "", nil
	})
	req, err := http.NewRequest("GET", "/", bytes.NewBuffer([]byte("")))
	if err != nil {
		t.Fatal(err)
	}

	req.AddCookie(&http.Cookie{Name: types.CookieName, Value: "123456"})

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := r.Context().Value(types.AuthenticatedUserID{})

		expectedUserID := "234567"
		if value != expectedUserID {
			t.Fatal("Response SHOULD BE '"+expectedUserID+"' but found ", value)
		}
	})

	recorder := httptest.NewRecorder()
	handler := WebAppendUserIdIfExistsMiddleware(testHandler, auth)
	handler.ServeHTTP(recorder, req)

	if recorder.Body.String() != "" {
		t.Fatal("Response SHOULD BE empty but found ", recorder.Body.String())
	}
}
