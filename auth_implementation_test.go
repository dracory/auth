package auth

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dracory/auth/internal/links"
	"github.com/dracory/auth/types"
)

func TestGetCurrentUserID_EmptyContextReturnsEmptyString(t *testing.T) {
	auth := &authImplementation{}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	userID := auth.GetCurrentUserID(req)

	if userID != "" {
		t.Fatalf("expected empty user ID, got %q", userID)
	}
}

func TestGetCurrentUserID_ReturnsUserIDFromContext(t *testing.T) {
	auth := &authImplementation{}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), types.AuthenticatedUserID{}, "12345")
	req = req.WithContext(ctx)

	userID := auth.GetCurrentUserID(req)

	if userID != "12345" {
		t.Fatalf("expected user ID %q, got %q", "12345", userID)
	}
}

func TestLinkAddsSlashWhenMissing(t *testing.T) {
	endpoint := "http://example.com/auth"
	uri := PathApiLogin

	got := links.Join(endpoint, uri)
	want := endpoint + "/" + uri

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestLinkKeepsTrailingSlash(t *testing.T) {
	endpoint := "http://example.com/auth/"
	uri := PathApiLogin

	got := links.Join(endpoint, uri)
	want := endpoint + uri

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestAuthLinkHelpersRespectEndpointTrailingSlash(t *testing.T) {
	authWithSlash := &authImplementation{endpoint: "http://localhost/auth/"}

	if got, want := authWithSlash.LinkApiLogin(), "http://localhost/auth/"+PathApiLogin; got != want {
		t.Fatalf("LinkApiLogin with trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authWithSlash.LinkLogin(), "http://localhost/auth/"+PathLogin; got != want {
		t.Fatalf("LinkLogin with trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authWithSlash.LinkRegister(), "http://localhost/auth/"+PathRegister; got != want {
		t.Fatalf("LinkRegister with trailing slash: expected %q, got %q", want, got)
	}
}

func TestAuthLinkHelpers_NoTrailingSlash(t *testing.T) {
	authNoSlash := &authImplementation{endpoint: "http://localhost/auth"}

	if got, want := authNoSlash.LinkApiLoginCodeVerify(), "http://localhost/auth/"+PathApiLoginCodeVerify; got != want {
		t.Fatalf("LinkApiLoginCodeVerify without trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authNoSlash.LinkApiLogout(), "http://localhost/auth/"+PathApiLogout; got != want {
		t.Fatalf("LinkApiLogout without trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authNoSlash.LinkApiRegister(), "http://localhost/auth/"+PathApiRegister; got != want {
		t.Fatalf("LinkApiRegister without trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authNoSlash.LinkApiRegisterCodeVerify(), "http://localhost/auth/"+PathApiRegisterCodeVerify; got != want {
		t.Fatalf("LinkApiRegisterCodeVerify without trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authNoSlash.LinkApiPasswordRestore(), "http://localhost/auth/"+PathApiRestorePassword; got != want {
		t.Fatalf("LinkApiPasswordRestore without trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authNoSlash.LinkApiPasswordReset(), "http://localhost/auth/"+PathApiResetPassword; got != want {
		t.Fatalf("LinkApiPasswordReset without trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authNoSlash.LinkLoginCodeVerify(), "http://localhost/auth/"+PathLoginCodeVerify; got != want {
		t.Fatalf("LinkLoginCodeVerify without trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authNoSlash.LinkLogout(), "http://localhost/auth/"+PathLogout; got != want {
		t.Fatalf("LinkLogout without trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authNoSlash.LinkPasswordRestore(), "http://localhost/auth/"+PathPasswordRestore; got != want {
		t.Fatalf("LinkPasswordRestore without trailing slash: expected %q, got %q", want, got)
	}

	token := "abc123"
	if got, want := authNoSlash.LinkPasswordReset(token), "http://localhost/auth/"+PathPasswordReset+"?t="+token; got != want {
		t.Fatalf("LinkPasswordReset without trailing slash: expected %q, got %q", want, got)
	}

	if got, want := authNoSlash.LinkRegisterCodeVerify(), "http://localhost/auth/"+PathRegisterCodeVerify; got != want {
		t.Fatalf("LinkRegisterCodeVerify without trailing slash: expected %q, got %q", want, got)
	}
}

func TestAuthLinkRedirectOnSuccess(t *testing.T) {
	authInstance := &authImplementation{urlRedirectOnSuccess: "/dashboard"}

	if got, want := authInstance.LinkRedirectOnSuccess(), "/dashboard"; got != want {
		t.Fatalf("LinkRedirectOnSuccess: expected %q, got %q", want, got)
	}
}

func TestRegistrationEnableDisableToggle(t *testing.T) {
	var authInstance authImplementation

	if authInstance.enableRegistration {
		t.Fatalf("expected enableRegistration to be false by default")
	}

	authInstance.RegistrationEnable()
	if !authInstance.enableRegistration {
		t.Fatalf("expected enableRegistration to be true after RegistrationEnable")
	}

	authInstance.RegistrationDisable()
	if authInstance.enableRegistration {
		t.Fatalf("expected enableRegistration to be false after RegistrationDisable")
	}
}

func TestSetAndGetEndpoint(t *testing.T) {
	a := &authImplementation{}
	a.SetEndpoint("http://localhost:8080")
	if got := a.GetEndpoint(); got != "http://localhost:8080" {
		t.Fatalf("expected endpoint 'http://localhost:8080', got %q", got)
	}
}

func TestSetAndGetLayout(t *testing.T) {
	a := &authImplementation{}
	fn := func(content string) string { return "<html>" + content + "</html>" }
	a.SetLayout(fn)
	if a.GetLayout() == nil {
		t.Fatal("expected non-nil layout")
	}
}

func TestSetAndGetLogger(t *testing.T) {
	a := &authImplementation{}
	if a.GetLogger() == nil {
		t.Fatal("expected non-nil default logger")
	}
	customLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	a.SetLogger(customLogger)
	if a.GetLogger() != customLogger {
		t.Fatal("expected custom logger to be set")
	}
}

func TestSetAndGetObservabilityHooks(t *testing.T) {
	a := &authImplementation{}
	if a.GetObservabilityHooks() == nil {
		t.Fatal("expected non-nil default hooks")
	}
	a.SetObservabilityHooks(nil)
	if a.GetObservabilityHooks() == nil {
		t.Fatal("expected non-nil hooks fallback")
	}
}

func TestSetAndGetUseCookies(t *testing.T) {
	a := &authImplementation{}
	a.SetUseCookies(true)
	if !a.GetUseCookies() {
		t.Fatal("expected useCookies true")
	}
}

func TestSetAndGetDisableRateLimit(t *testing.T) {
	a := &authImplementation{}
	a.SetDisableRateLimit(true)
	if !a.GetDisableRateLimit() {
		t.Fatal("expected disableRateLimit true")
	}
}

func TestSetAndGetPasswordStrength(t *testing.T) {
	a := &authImplementation{}
	cfg := &types.PasswordStrengthConfig{MinLength: 10}
	a.SetPasswordStrength(cfg)
	if a.GetPasswordStrength() != cfg {
		t.Fatal("expected password strength config to match")
	}
}

func TestSetAndGetFuncUserLogin(t *testing.T) {
	a := &authImplementation{}
	fn := func(ctx context.Context, username, password string, options types.UserAuthOptions) (string, error) {
		return "user1", nil
	}
	a.SetFuncUserLogin(fn)
	if a.GetFuncUserLogin() == nil {
		t.Fatal("expected non-nil funcUserLogin")
	}
}

func TestSetAndGetFuncUserRegister(t *testing.T) {
	a := &authImplementation{}
	fn := func(ctx context.Context, username, password, firstName, lastName string, options types.UserAuthOptions) error {
		return nil
	}
	a.SetFuncUserRegister(fn)
	if a.GetFuncUserRegister() == nil {
		t.Fatal("expected non-nil funcUserRegister")
	}
}

func TestSetAndGetFuncUserPasswordChange(t *testing.T) {
	a := &authImplementation{}
	fn := func(ctx context.Context, userID, password string, options types.UserAuthOptions) error {
		return nil
	}
	a.SetFuncUserPasswordChange(fn)
	if a.GetFuncUserPasswordChange() == nil {
		t.Fatal("expected non-nil funcUserPasswordChange")
	}
}

func TestSetAndGetFuncUserLogout(t *testing.T) {
	a := &authImplementation{}
	fn := func(ctx context.Context, userID string, options types.UserAuthOptions) error {
		return nil
	}
	a.SetFuncUserLogout(fn)
	if a.GetFuncUserLogout() == nil {
		t.Fatal("expected non-nil funcUserLogout")
	}
}

func TestSetAndGetFuncUserStoreAuthToken(t *testing.T) {
	a := &authImplementation{}
	fn := func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
		return nil
	}
	a.SetFuncUserStoreAuthToken(fn)
	if a.GetFuncUserStoreAuthToken() == nil {
		t.Fatal("expected non-nil funcUserStoreAuthToken")
	}
}

func TestSetAndGetFuncUserFindByAuthToken(t *testing.T) {
	a := &authImplementation{}
	fn := func(ctx context.Context, token string, options types.UserAuthOptions) (string, error) {
		return "user1", nil
	}
	a.SetFuncUserFindByAuthToken(fn)
	if a.GetFuncUserFindByAuthToken() == nil {
		t.Fatal("expected non-nil funcUserFindByAuthToken")
	}
}

func TestSetAndGetFuncUserFindByUsername(t *testing.T) {
	a := &authImplementation{}
	fn := func(ctx context.Context, username, firstName, lastName string, options types.UserAuthOptions) (string, error) {
		return "user1", nil
	}
	a.SetFuncUserFindByUsername(fn)
	if a.GetFuncUserFindByUsername() == nil {
		t.Fatal("expected non-nil funcUserFindByUsername")
	}
}

func TestSetAndGetFuncEmailTemplatePasswordRestore(t *testing.T) {
	a := &authImplementation{}
	fn := func(ctx context.Context, userID string, link string, options types.UserAuthOptions) string {
		return "body"
	}
	a.SetFuncEmailTemplatePasswordRestore(fn)
	if a.GetFuncEmailTemplatePasswordRestore() == nil {
		t.Fatal("expected non-nil funcEmailTemplatePasswordRestore")
	}
}

func TestSetAndGetFuncEmailTemplateRegisterCode(t *testing.T) {
	a := &authImplementation{}
	fn := func(ctx context.Context, email string, code string, options types.UserAuthOptions) string {
		return "body"
	}
	a.SetFuncEmailTemplateRegisterCode(fn)
	if a.GetFuncEmailTemplateRegisterCode() == nil {
		t.Fatal("expected non-nil funcEmailTemplateRegisterCode")
	}
}

func TestSetAndGetFuncEmailSend(t *testing.T) {
	a := &authImplementation{}
	fn := func(ctx context.Context, userID, subject, body string) error {
		return nil
	}
	a.SetFuncEmailSend(fn)
	if a.GetFuncEmailSend() == nil {
		t.Fatal("expected non-nil funcEmailSend")
	}
}

func TestSetAndGetFuncTemporaryKeyGet(t *testing.T) {
	a := &authImplementation{}
	fn := func(key string) (string, error) { return "val", nil }
	a.SetFuncTemporaryKeyGet(fn)
	if a.GetFuncTemporaryKeyGet() == nil {
		t.Fatal("expected non-nil funcTemporaryKeyGet")
	}
}

func TestSetAndGetFuncTemporaryKeySet(t *testing.T) {
	a := &authImplementation{}
	fn := func(key string, value string, expiresSeconds int) error { return nil }
	a.SetFuncTemporaryKeySet(fn)
	if a.GetFuncTemporaryKeySet() == nil {
		t.Fatal("expected non-nil funcTemporaryKeySet")
	}
}

func TestSetAndGetPasswordlessUserRegister(t *testing.T) {
	a := &authImplementation{}
	fn := func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) error {
		return nil
	}
	a.SetPasswordlessUserRegister(fn)
	if a.GetPasswordlessUserRegister() == nil {
		t.Fatal("expected non-nil passwordlessUserRegister")
	}
}

func TestSetAndGetPasswordlessUserFindByEmail(t *testing.T) {
	a := &authImplementation{}
	fn := func(ctx context.Context, email string, options types.UserAuthOptions) (string, error) {
		return "user1", nil
	}
	a.SetPasswordlessUserFindByEmail(fn)
	if a.GetPasswordlessUserFindByEmail() == nil {
		t.Fatal("expected non-nil passwordlessUserFindByEmail")
	}
}

func TestSetAndGetPasswordlessFuncEmailTemplateLoginCode(t *testing.T) {
	a := &authImplementation{}
	fn := func(ctx context.Context, email, link string, options types.UserAuthOptions) string {
		return "body"
	}
	a.SetPasswordlessFuncEmailTemplateLoginCode(fn)
	if a.GetPasswordlessFuncEmailTemplateLoginCode() == nil {
		t.Fatal("expected non-nil passwordlessFuncEmailTemplateLoginCode")
	}
}

func TestSetAndGetPasswordlessFuncEmailTemplateRegisterCode(t *testing.T) {
	a := &authImplementation{}
	fn := func(ctx context.Context, email, link string, options types.UserAuthOptions) string {
		return "body"
	}
	a.SetPasswordlessFuncEmailTemplateRegisterCode(fn)
	if a.GetPasswordlessFuncEmailTemplateRegisterCode() == nil {
		t.Fatal("expected non-nil passwordlessFuncEmailTemplateRegisterCode")
	}
}

func TestSetAndGetPasswordlessFuncEmailSend(t *testing.T) {
	a := &authImplementation{}
	fn := func(ctx context.Context, email, subject, body string) error {
		return nil
	}
	a.SetPasswordlessFuncEmailSend(fn)
	if a.GetPasswordlessFuncEmailSend() == nil {
		t.Fatal("expected non-nil passwordlessFuncEmailSend")
	}
}

func TestIsPasswordless(t *testing.T) {
	a := &authImplementation{passwordless: true}
	if !a.IsPasswordless() {
		t.Fatal("expected passwordless true")
	}
}

func TestIsVerificationEnabled(t *testing.T) {
	a := &authImplementation{enableVerification: true}
	if !a.IsVerificationEnabled() {
		t.Fatal("expected verification enabled true")
	}
}

func TestIsImpersonationEnabled(t *testing.T) {
	a := &authImplementation{enableImpersonation: true}
	if !a.IsImpersonationEnabled() {
		t.Fatal("expected impersonation enabled true")
	}
}

func TestSetAndGetFuncCanImpersonate(t *testing.T) {
	a := &authImplementation{}
	fn := func(ctx context.Context, adminUserID, targetUserID string) (bool, error) {
		return true, nil
	}
	a.SetFuncCanImpersonate(fn)
	if a.GetFuncCanImpersonate() == nil {
		t.Fatal("expected non-nil funcCanImpersonate")
	}
}

func TestSetAndGetFuncImpersonationStart(t *testing.T) {
	a := &authImplementation{}
	fn := func(ctx context.Context, adminUserID, targetUserID string) error {
		return nil
	}
	a.SetFuncImpersonationStart(fn)
	if a.GetFuncImpersonationStart() == nil {
		t.Fatal("expected non-nil funcImpersonationStart")
	}
}

func TestSetAndGetFuncImpersonationStop(t *testing.T) {
	a := &authImplementation{}
	fn := func(ctx context.Context, adminUserID, targetUserID string) error {
		return nil
	}
	a.SetFuncImpersonationStop(fn)
	if a.GetFuncImpersonationStop() == nil {
		t.Fatal("expected non-nil funcImpersonationStop")
	}
}

func TestLinkApiImpersonateStart(t *testing.T) {
	a := &authImplementation{endpoint: "http://localhost/auth"}
	expected := "http://localhost/auth/" + PathApiImpersonateStart
	if got := a.LinkApiImpersonateStart(); got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestLinkApiImpersonateStop(t *testing.T) {
	a := &authImplementation{endpoint: "http://localhost/auth"}
	expected := "http://localhost/auth/" + PathApiImpersonateStop
	if got := a.LinkApiImpersonateStop(); got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestIsRegistrationEnabled(t *testing.T) {
	a := &authImplementation{enableRegistration: true}
	if !a.IsRegistrationEnabled() {
		t.Fatal("expected registration enabled true")
	}
}

func TestGetCookieName_Default(t *testing.T) {
	a := &authImplementation{}
	if got := a.GetCookieName(); got != CookieName {
		t.Fatalf("expected default cookie name %q, got %q", CookieName, got)
	}
}

func TestGetCookieName_Custom(t *testing.T) {
	a := &authImplementation{cookieName: "custom-auth"}
	if got := a.GetCookieName(); got != "custom-auth" {
		t.Fatalf("expected cookie name %q, got %q", "custom-auth", got)
	}
}

func TestSetCookieName(t *testing.T) {
	a := &authImplementation{}
	a.SetCookieName("custom-auth")
	if got := a.GetCookieName(); got != "custom-auth" {
		t.Fatalf("expected cookie name %q, got %q", "custom-auth", got)
	}
}

func TestSetAuthCookie_UsesCookieName(t *testing.T) {
	a := &authImplementation{
		cookieName: "custom-auth",
		cookieConfig: CookieConfig{
			HttpOnly: types.CookieHttpOnly,
			Secure:   types.CookieInsecure,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   7200,
			Path:     "/",
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	a.SetAuthCookie(w, r, "test-token")

	res := w.Result()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected a cookie to be set")
	}

	if cookies[0].Name != "custom-auth" {
		t.Fatalf("expected cookie name %q, got %q", "custom-auth", cookies[0].Name)
	}

	if cookies[0].Value != "test-token" {
		t.Fatalf("expected cookie value %q, got %q", "test-token", cookies[0].Value)
	}
}

func TestRemoveAuthCookie_UsesCookieName(t *testing.T) {
	a := &authImplementation{
		cookieName: "custom-auth",
		cookieConfig: CookieConfig{
			HttpOnly: types.CookieHttpOnly,
			Secure:   types.CookieInsecure,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   7200,
			Path:     "/",
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	a.RemoveAuthCookie(w, r)

	res := w.Result()
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected a cookie to be set")
	}

	if cookies[0].Name != "custom-auth" {
		t.Fatalf("expected cookie name %q, got %q", "custom-auth", cookies[0].Name)
	}

	if cookies[0].Value != "none" {
		t.Fatalf("expected cookie value %q, got %q", "none", cookies[0].Value)
	}
}
