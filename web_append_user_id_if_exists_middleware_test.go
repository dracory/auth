package auth

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dracory/auth/types"
)

func TestBlockRobotsMiddlewareShouldPassForHomeWithoutSlash(t *testing.T) {
	auth, err := NewPasswordlessAuth(ConfigPasswordless{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/user",
		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
		FuncUserFindByAuthToken: func(ctx context.Context, sessionID string, options types.UserAuthOptions) (userID string, err error) {
			if sessionID == "123456" {
				return "234567", nil
			}
			return "", nil
		},
		FuncUserFindByEmail: func(ctx context.Context, email string, options types.UserAuthOptions) (userID string, err error) {
			return "", nil
		},
		FuncUserLogout: func(ctx context.Context, userID string, options types.UserAuthOptions) (err error) {
			return nil
		},
		FuncUserStoreAuthToken: func(ctx context.Context, sessionID, userID string, options types.UserAuthOptions) error {
			return nil
		},
		FuncEmailSend: func(ctx context.Context, email, emailSubject, emailBody string) (err error) {
			return nil
		},
		UseCookies:      true,
		UseLocalStorage: false,
	})

	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("GET", "/", bytes.NewBuffer([]byte("")))
	if err != nil {
		t.Fatal(err)
	}

	req.AddCookie(&http.Cookie{Name: CookieName, Value: "123456"})

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := r.Context().Value(AuthenticatedUserID{})

		expectedUserID := "234567"
		if value != expectedUserID {
			t.Fatal("Response SHOULD BE '"+expectedUserID+"' but found ", value)
		}
	})

	recorder := httptest.NewRecorder()
	handler := auth.WebAppendUserIdIfExistsMiddleware(testHandler)
	handler.ServeHTTP(recorder, req)

	if recorder.Body.String() != "" {
		t.Fatal("Response SHOULD BE empty but found ", recorder.Body.String())
	}
}
