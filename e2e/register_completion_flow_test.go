package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth"
	"github.com/dracory/auth/internal/testutils"
	"github.com/dracory/auth/types"
)

// TestRegisterCompletionFlow exercises the whole authenticated-completion
// path end to end through a real auth implementation:
//
//  1. GET /auth/register with a valid session cookie → completion form
//     (verified email read-only, no editable email input, no password).
//  2. POST /auth/api/register with the same cookie + names → the host's
//     UserProfileUpdate runs against the session user (a forged email in
//     the form body is ignored) and the response carries the redirect.
//  3. The same requests without a cookie → the public sign-up form.
func TestRegisterCompletionFlow(t *testing.T) {
	var updatedUserID, updatedFirst, updatedLast string

	cfg := testutils.NewPasswordlessConfigForTest()
	cfg.FuncUserFindByAuthToken = func(ctx context.Context, token string, options types.UserAuthOptions) (string, error) {
		if token == "session-abc" {
			return "user-1", nil
		}
		return "", nil
	}

	a, err := auth.NewPasswordlessAuth(cfg)
	if err != nil {
		t.Fatalf("NewPasswordlessAuth: %v", err)
	}

	impl, ok := a.(interface {
		SetFuncUserProfileUpdate(func(ctx context.Context, userID string, firstName string, lastName string, options types.UserAuthOptions) error)
		SetFuncUserEmailGet(func(ctx context.Context, userID string, options types.UserAuthOptions) (string, error))
	})
	if !ok {
		t.Fatal("impl does not expose the profile hooks")
	}
	impl.SetFuncUserEmailGet(func(ctx context.Context, userID string, options types.UserAuthOptions) (string, error) {
		return "otp-user@example.com", nil
	})
	impl.SetFuncUserProfileUpdate(func(ctx context.Context, userID string, firstName string, lastName string, options types.UserAuthOptions) error {
		updatedUserID, updatedFirst, updatedLast = userID, firstName, lastName
		return nil
	})

	mux := a.Router()
	sessionCookie := &http.Cookie{Name: a.GetCookieName(), Value: "session-abc"}

	// 1. GET /auth/register as the OTP-logged-in user.
	req := httptest.NewRequest("GET", "/auth/register", nil)
	req.AddCookie(sessionCookie)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"Complete your registration", "Email (verified)", "otp-user@example.com", "form-control-plaintext"} {
		if !strings.Contains(body, want) {
			t.Fatalf("completion page missing %q", want)
		}
	}
	for _, forbidden := range []string{`name="email"`, `name="password"`} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("completion page must not contain %s", forbidden)
		}
	}

	// 2. POST /auth/api/register as the same user — forged email ignored.
	form := url.Values{
		"first_name": {"Jane"},
		"last_name":  {"Doe"},
		"email":      {"attacker@example.com"},
	}
	post := httptest.NewRequest("POST", "/auth/api/register", strings.NewReader(form.Encode()))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	post.AddCookie(sessionCookie)
	postRec := httptest.NewRecorder()
	mux.ServeHTTP(postRec, post)

	postBody := postRec.Body.String()
	if !strings.Contains(postBody, `"status":"success"`) {
		t.Fatalf("expected success from api/register, got %q", postBody)
	}
	if updatedUserID != "user-1" || updatedFirst != "Jane" || updatedLast != "Doe" {
		t.Fatalf("UserProfileUpdate got (%q, %q, %q)", updatedUserID, updatedFirst, updatedLast)
	}
	if !strings.Contains(postBody, `"redirect":"http://localhost/dashboard"`) {
		t.Fatalf("expected redirect to dashboard, got %q", postBody)
	}

	// 3. Same GET without the cookie → public sign-up form.
	guestReq := httptest.NewRequest("GET", "/auth/register", nil)
	guestRec := httptest.NewRecorder()
	mux.ServeHTTP(guestRec, guestReq)

	guestBody := guestRec.Body.String()
	if !strings.Contains(guestBody, "Create your account") || !strings.Contains(guestBody, `name="email"`) {
		t.Fatalf("expected public register form for guest, got %q", guestBody)
	}
}
