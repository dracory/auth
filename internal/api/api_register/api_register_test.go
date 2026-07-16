package api_register

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dracory/auth/types"
)

func makePostRequest(t *testing.T, path string, values url.Values) (*httptest.ResponseRecorder, *http.Request) {
	body := strings.NewReader(values.Encode())
	req, err := http.NewRequest("POST", path, body)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	recorder := httptest.NewRecorder()
	return recorder, req
}

// Username and password registration tests

func TestApiRegisterUsernameAndPasswordRequiresFirstName(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		RegisterWithUsernameAndPassword: func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (string, string) {
			if firstName == "" {
				return "", "First name is required field"
			}
			return "", ""
		},
	}

	recorder, req := makePostRequest(t, "/api/register", url.Values{})
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"error\"") {
		t.Fatalf("expected status error, got %q", body)
	}
	if !strings.Contains(body, "\"message\":\"First name is required field\"") {
		t.Fatalf("expected first name required message, got %q", body)
	}
}

func TestApiRegisterUsernameAndPasswordRequiresLastName(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		RegisterWithUsernameAndPassword: func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (string, string) {
			if lastName == "" {
				return "", "Last name is required field"
			}
			return "", ""
		},
	}

	values := url.Values{
		"first_name": {"John"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Last name is required field\"") {
		t.Fatalf("expected last name required message, got %q", body)
	}
}

func TestApiRegisterUsernameAndPasswordRequiresEmail(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		RegisterWithUsernameAndPassword: func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (string, string) {
			if email == "" {
				return "", "Email is required field"
			}
			return "", ""
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Email is required field\"") {
		t.Fatalf("expected email required message, got %q", body)
	}
}

func TestApiRegisterUsernameAndPasswordRequiresPassword(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		RegisterWithUsernameAndPassword: func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (string, string) {
			if password == "" {
				return "", "Password is required field"
			}
			return "", ""
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Password is required field\"") {
		t.Fatalf("expected password required message, got %q", body)
	}
}

func TestApiRegisterUsernameAndPasswordInvalidEmail(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		RegisterWithUsernameAndPassword: func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (string, string) {
			if email == "invalid-email" {
				return "", "This is not a valid email: invalid-email"
			}
			return "", ""
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"invalid-email"},
		"password":   {"1234"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"This is not a valid email: invalid-email\"") {
		t.Fatalf("expected invalid email message, got %q", body)
	}
}

func TestApiRegisterUsernameAndPasswordFuncUserRegisterNotDefined(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		RegisterWithUsernameAndPassword: func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (string, string) {
			return "", "registration failed. FuncUserRegister function not defined"
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
		"password":   {"1234"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"registration failed. FuncUserRegister function not defined\"") {
		t.Fatalf("expected func not defined message, got %q", body)
	}
}

func TestApiRegisterUsernameAndPasswordRegistrationFailed(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		RegisterWithUsernameAndPassword: func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (string, string) {
			return "", "registration failed."
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
		"password":   {"1234"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"registration failed.\"") {
		t.Fatalf("expected registration failed message, got %q", body)
	}
}

func TestApiRegisterUsernameAndPasswordSuccess(t *testing.T) {
	deps := Dependencies{
		Passwordless: false,
		RegisterWithUsernameAndPassword: func(ctx context.Context, email, password, firstName, lastName, ip, userAgent string) (string, string) {
			return "registration success", ""
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
		"password":   {"1234"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"success\"") {
		t.Fatalf("expected success status, got %q", body)
	}
	if !strings.Contains(body, "\"message\":\"registration success\"") {
		t.Fatalf("expected registration success message, got %q", body)
	}
}

// Passwordless registration tests

func TestApiRegisterPasswordlessRequiresFirstName(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
	}

	recorder, req := makePostRequest(t, "/api/register", url.Values{})
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"error\"") {
		t.Fatalf("expected status error, got %q", body)
	}
	if !strings.Contains(body, "\"message\":\"First name is required field\"") {
		t.Fatalf("expected first name required message, got %q", body)
	}
}

func TestApiRegisterPasswordlessRequiresLastName(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
	}

	values := url.Values{
		"first_name": {"John"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Last name is required field\"") {
		t.Fatalf("expected last name required message, got %q", body)
	}
}

func TestApiRegisterPasswordlessRequiresEmail(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Email is required field\"") {
		t.Fatalf("expected email required message, got %q", body)
	}
}

func TestApiRegisterPasswordlessTokenStoreError(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		RegisterPasswordlessInitDependencies: RegisterPasswordlessInitDependencies{
			TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
				return errors.New("db error")
			},
			ExpiresSeconds: 3600,
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Failed to process request. Please try again later\"") {
		t.Fatalf("expected token store error message, got %q", body)
	}
}

func TestApiRegisterPasswordlessEmailSendError(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		RegisterPasswordlessInitDependencies: RegisterPasswordlessInitDependencies{
			TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
				return nil
			},
			ExpiresSeconds: 3600,
			EmailTemplate: func(ctx context.Context, email string, verificationCode string) string {
				return "body"
			},
			EmailSend: func(ctx context.Context, email string, subject string, body string) error {
				return errors.New("smtp error")
			},
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"message\":\"Failed to send email. Please try again later\"") {
		t.Fatalf("expected email send error message, got %q", body)
	}
}

func TestApiRegisterPasswordlessSuccess(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		RegisterPasswordlessInitDependencies: RegisterPasswordlessInitDependencies{
			TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
				return nil
			},
			ExpiresSeconds: 3600,
			EmailTemplate: func(ctx context.Context, email string, verificationCode string) string {
				return "body"
			},
			EmailSend: func(ctx context.Context, email string, subject string, body string) error {
				return nil
			},
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"success\"") {
		t.Fatalf("expected success status, got %q", body)
	}
	if !strings.Contains(body, "\"message\":\"Registration code was sent successfully\"") {
		t.Fatalf("expected success message, got %q", body)
	}
}

func TestRegisterPasswordlessInitError_Error(t *testing.T) {
	var e *RegisterPasswordlessInitError
	if e.Error() != "" {
		t.Fatalf("expected empty string for nil receiver, got %q", e.Error())
	}

	e = &RegisterPasswordlessInitError{Message: "custom message"}
	if e.Error() != "custom message" {
		t.Fatalf("expected 'custom message', got %q", e.Error())
	}

	e = &RegisterPasswordlessInitError{Err: errors.New("inner error")}
	if e.Error() != "inner error" {
		t.Fatalf("expected 'inner error', got %q", e.Error())
	}

	e = &RegisterPasswordlessInitError{Code: "some_code"}
	if e.Error() != "some_code" {
		t.Fatalf("expected 'some_code', got %q", e.Error())
	}
}

func TestApiRegisterNilRegisterWithUsernameAndPassword(t *testing.T) {
	deps := Dependencies{
		Passwordless:                    false,
		RegisterWithUsernameAndPassword: nil,
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
		"password":   {"1234"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "Registration failed") {
		t.Fatalf("expected registration failed, got %q", body)
	}
}

func TestApiRegisterPasswordlessInvalidEmail(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"invalid-email"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"error\"") {
		t.Fatalf("expected error status, got %q", body)
	}
}

func TestApiRegisterPasswordlessNilTemporaryKeySet(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		RegisterPasswordlessInitDependencies: RegisterPasswordlessInitDependencies{
			TemporaryKeySet: nil,
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "Failed to process request") {
		t.Fatalf("expected failed to process, got %q", body)
	}
}

func TestApiRegisterPasswordlessNilEmailTemplateAndSend(t *testing.T) {
	deps := Dependencies{
		Passwordless: true,
		RegisterPasswordlessInitDependencies: RegisterPasswordlessInitDependencies{
			TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
				return nil
			},
			ExpiresSeconds: 3600,
			EmailTemplate:  nil,
			EmailSend:      nil,
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "Internal server error") {
		t.Fatalf("expected internal server error, got %q", body)
	}
}

func TestApiRegisterPasswordlessDefaultExpiration(t *testing.T) {
	keySetExpires := 0
	deps := Dependencies{
		Passwordless: true,
		RegisterPasswordlessInitDependencies: RegisterPasswordlessInitDependencies{
			TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
				keySetExpires = expiresSeconds
				return nil
			},
			ExpiresSeconds: 0,
			EmailTemplate: func(ctx context.Context, email string, verificationCode string) string {
				return "body"
			},
			EmailSend: func(ctx context.Context, email string, subject string, body string) error {
				return nil
			},
		},
	}

	values := url.Values{
		"first_name": {"John"},
		"last_name":  {"Doe"},
		"email":      {"test@test.com"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	if keySetExpires != 3600 {
		t.Fatalf("expected default expiration 3600, got %d", keySetExpires)
	}
}

// AuthKnight registration tests

func TestApiRegisterAuthKnight_Success(t *testing.T) {
	var registeredEmail, registeredFirstName, registeredLastName string
	var storedToken, storedUserID string
	var cookieSet bool
	var clearedKey string
	var clearedValue string

	deps := Dependencies{
		AuthKnight: true,
		AuthKnightRegisterDependencies: AuthKnightRegisterDependencies{
			TemporaryKeyGet: func(key string) (string, error) {
				if key == "test-ak-key" {
					return "newuser@example.com", nil
				}
				return "", errors.New("key not found")
			},
			TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
				clearedKey = key
				clearedValue = value
				return nil
			},
			UserRegister: func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (string, error) {
				registeredEmail = email
				registeredFirstName = firstName
				registeredLastName = lastName
				return "user-new-1", nil
			},
			UserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
				storedToken = token
				storedUserID = userID
				return nil
			},
			UseCookies: true,
			SetAuthCookie: func(w http.ResponseWriter, r *http.Request, token string) {
				cookieSet = true
			},
		},
	}

	values := url.Values{
		"ak_key":     {"test-ak-key"},
		"first_name": {"Jane"},
		"last_name":  {"Smith"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	if registeredEmail != "newuser@example.com" {
		t.Fatalf("expected registered email 'newuser@example.com', got '%s'", registeredEmail)
	}
	if registeredFirstName != "Jane" {
		t.Fatalf("expected registered firstName 'Jane', got '%s'", registeredFirstName)
	}
	if registeredLastName != "Smith" {
		t.Fatalf("expected registered lastName 'Smith', got '%s'", registeredLastName)
	}
	if storedUserID != "user-new-1" {
		t.Fatalf("expected stored userID 'user-new-1', got '%s'", storedUserID)
	}
	if storedToken == "" {
		t.Fatal("expected non-empty stored token")
	}
	if !cookieSet {
		t.Fatal("expected cookie to be set")
	}
	if clearedKey != "test-ak-key" {
		t.Fatalf("expected ak_key to be cleared, got key='%s'", clearedKey)
	}
	if clearedValue != "" {
		t.Fatalf("expected ak_key value to be cleared to empty, got '%s'", clearedValue)
	}
}

func TestApiRegisterAuthKnight_AkKeyConsumedAfterRegistration(t *testing.T) {
	keyStore := make(map[string]string)
	keyStore["test-ak-key"] = "newuser@example.com"

	deps := Dependencies{
		AuthKnight: true,
		AuthKnightRegisterDependencies: AuthKnightRegisterDependencies{
			TemporaryKeyGet: func(key string) (string, error) {
				v, ok := keyStore[key]
				if !ok {
					return "", errors.New("key not found")
				}
				return v, nil
			},
			TemporaryKeySet: func(key string, value string, expiresSeconds int) error {
				if value == "" {
					delete(keyStore, key)
				} else {
					keyStore[key] = value
				}
				return nil
			},
			UserRegister: func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (string, error) {
				return "user-new-1", nil
			},
			UserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
				return nil
			},
		},
	}

	values := url.Values{
		"ak_key":     {"test-ak-key"},
		"first_name": {"Jane"},
		"last_name":  {"Smith"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "registration success") {
		t.Fatalf("expected registration success, got: %s", body)
	}

	// Second attempt with the same ak_key should fail
	recorder2, req2 := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder2, req2, deps)

	body2 := recorder2.Body.String()
	if !strings.Contains(body2, "Link not valid or expired") {
		t.Fatalf("expected expired link error on second use of ak_key, got: %s", body2)
	}
}

func TestApiRegisterAuthKnight_MissingAkKey(t *testing.T) {
	deps := Dependencies{
		AuthKnight: true,
		AuthKnightRegisterDependencies: AuthKnightRegisterDependencies{
			TemporaryKeyGet: func(key string) (string, error) { return "", nil },
			UserRegister: func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (string, error) {
				t.Fatal("should not call UserRegister")
				return "", nil
			},
			UserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
				return nil
			},
		},
	}

	values := url.Values{
		"first_name": {"Jane"},
		"last_name":  {"Smith"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "Once token is required") {
		t.Fatalf("expected error about missing once token, got: %s", body)
	}
}

func TestApiRegisterAuthKnight_InvalidAkKey(t *testing.T) {
	deps := Dependencies{
		AuthKnight: true,
		AuthKnightRegisterDependencies: AuthKnightRegisterDependencies{
			TemporaryKeyGet: func(key string) (string, error) {
				return "", errors.New("key not found")
			},
			UserRegister: func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (string, error) {
				t.Fatal("should not call UserRegister")
				return "", nil
			},
			UserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
				return nil
			},
		},
	}

	values := url.Values{
		"ak_key":     {"invalid-key"},
		"first_name": {"Jane"},
		"last_name":  {"Smith"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "Link not valid or expired") {
		t.Fatalf("expected error about expired link, got: %s", body)
	}
}

func TestApiRegisterAuthKnight_MissingFirstName(t *testing.T) {
	deps := Dependencies{
		AuthKnight: true,
		AuthKnightRegisterDependencies: AuthKnightRegisterDependencies{
			TemporaryKeyGet: func(key string) (string, error) {
				return "test@example.com", nil
			},
			UserRegister: func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (string, error) {
				t.Fatal("should not call UserRegister")
				return "", nil
			},
			UserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
				return nil
			},
		},
	}

	values := url.Values{
		"ak_key":    {"test-key"},
		"last_name": {"Smith"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "First name is required") {
		t.Fatalf("expected error about missing first name, got: %s", body)
	}
}

func TestApiRegisterAuthKnight_MissingLastName(t *testing.T) {
	deps := Dependencies{
		AuthKnight: true,
		AuthKnightRegisterDependencies: AuthKnightRegisterDependencies{
			TemporaryKeyGet: func(key string) (string, error) {
				return "test@example.com", nil
			},
			UserRegister: func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (string, error) {
				t.Fatal("should not call UserRegister")
				return "", nil
			},
			UserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
				return nil
			},
		},
	}

	values := url.Values{
		"ak_key":     {"test-key"},
		"first_name": {"Jane"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "Last name is required") {
		t.Fatalf("expected error about missing last name, got: %s", body)
	}
}

func TestApiRegisterAuthKnight_UserRegisterError(t *testing.T) {
	deps := Dependencies{
		AuthKnight: true,
		AuthKnightRegisterDependencies: AuthKnightRegisterDependencies{
			TemporaryKeyGet: func(key string) (string, error) {
				return "test@example.com", nil
			},
			UserRegister: func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (string, error) {
				return "", errors.New("database error")
			},
			UserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
				return nil
			},
		},
	}

	values := url.Values{
		"ak_key":     {"test-key"},
		"first_name": {"Jane"},
		"last_name":  {"Smith"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "Registration failed") {
		t.Fatalf("expected registration failed error, got: %s", body)
	}
}

func TestApiRegisterAuthKnight_NilTemporaryKeyGet(t *testing.T) {
	deps := Dependencies{
		AuthKnight: true,
		AuthKnightRegisterDependencies: AuthKnightRegisterDependencies{
			TemporaryKeyGet: nil,
			UserRegister: func(ctx context.Context, email, firstName, lastName string, options types.UserAuthOptions) (string, error) {
				t.Fatal("should not call UserRegister")
				return "", nil
			},
			UserStoreAuthToken: func(ctx context.Context, token, userID string, options types.UserAuthOptions) error {
				return nil
			},
		},
	}

	values := url.Values{
		"ak_key":     {"test-key"},
		"first_name": {"Jane"},
		"last_name":  {"Smith"},
	}
	recorder, req := makePostRequest(t, "/api/register", values)
	ApiRegister(recorder, req, deps)

	body := recorder.Body.String()
	if !strings.Contains(body, "Registration failed") {
		t.Fatalf("expected registration failed error, got: %s", body)
	}
}
