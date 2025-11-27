package auth

// func newPasswordlessAuthForMiddlewareTests(useCookies bool) (*Auth, error) {
// 	return NewPasswordlessAuth(ConfigPasswordless{
// 		Endpoint:             "/auth",
// 		UrlRedirectOnSuccess: "/user",
// 		FuncTemporaryKeyGet:  func(key string) (value string, err error) { return "", nil },
// 		FuncTemporaryKeySet:  func(key, value string, expiresSeconds int) (err error) { return nil },
// 		FuncUserFindByAuthToken: func(sessionID string, options UserAuthOptions) (userID string, err error) {
// 			// Default: no user found
// 			return "", nil
// 		},
// 		FuncUserFindByEmail:    func(email string, options UserAuthOptions) (userID string, err error) { return "", nil },
// 		FuncUserLogout:         func(userID string, options UserAuthOptions) (err error) { return nil },
// 		FuncUserStoreAuthToken: func(sessionID, userID string, options UserAuthOptions) error { return nil },
// 		FuncEmailSend:          func(email, emailSubject, emailBody string) (err error) { return nil },
// 		UseCookies:             useCookies,
// 		UseLocalStorage:        !useCookies,
// 	})
// }

// func TestAuthMiddleware_NoToken_UseCookies_RedirectsToLogin(t *testing.T) {
// 	authInstance, err := newPasswordlessAuthForMiddlewareTests(true)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	req, err := http.NewRequest("GET", "/", bytes.NewBuffer([]byte("")))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// No cookie set -> no token

// 	recorder := httptest.NewRecorder()
// 	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		t.Fatal("next handler should not be called when token is missing")
// 	})

// 	handler := authInstance.AuthMiddleware(next)
// 	handler.ServeHTTP(recorder, req)

// 	if recorder.Code != http.StatusTemporaryRedirect {
// 		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, recorder.Code)
// 	}

// 	location := recorder.Header().Get("Location")
// 	expectedLocation := authInstance.LinkLogin()
// 	if location != expectedLocation {
// 		t.Fatalf("expected redirect to %q, got %q", expectedLocation, location)
// 	}
// }

// func TestAuthMiddleware_NoToken_NoCookies_ReturnsUnauthenticated(t *testing.T) {
// 	authInstance, err := newPasswordlessAuthForMiddlewareTests(false)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	req, err := http.NewRequest("GET", "/", bytes.NewBuffer([]byte("")))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	recorder := httptest.NewRecorder()
// 	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		t.Fatal("next handler should not be called when token is missing")
// 	})

// 	handler := authInstance.AuthMiddleware(next)
// 	handler.ServeHTTP(recorder, req)

// 	body := recorder.Body.String()
// 	if !bytes.Contains([]byte(body), []byte("auth token is required")) {
// 		t.Fatalf("expected body to contain %q, got %q", "auth token is required", body)
// 	}
// }

// func TestAuthMiddleware_InvalidToken_UseCookies_RedirectsToLogin(t *testing.T) {
// 	authInstance, err := newPasswordlessAuthForMiddlewareTests(true)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Override token lookup to simulate error
// 	authInstance.funcUserFindByAuthToken = func(sessionID string, options UserAuthOptions) (userID string, err error) {
// 		return "", http.ErrNoCookie
// 	}

// 	req, err := http.NewRequest("GET", "/", bytes.NewBuffer([]byte("")))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	req.AddCookie(&http.Cookie{Name: CookieName, Value: "invalid-token"})

// 	recorder := httptest.NewRecorder()
// 	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		t.Fatal("next handler should not be called when token is invalid")
// 	})

// 	handler := authInstance.AuthMiddleware(next)
// 	handler.ServeHTTP(recorder, req)

// 	if recorder.Code != http.StatusTemporaryRedirect {
// 		t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, recorder.Code)
// 	}

// 	location := recorder.Header().Get("Location")
// 	expectedLocation := authInstance.LinkLogin()
// 	if location != expectedLocation {
// 		t.Fatalf("expected redirect to %q, got %q", expectedLocation, location)
// 	}
// }

// func TestAuthMiddleware_InvalidToken_NoCookies_ReturnsUnauthenticated(t *testing.T) {
// 	authInstance, err := newPasswordlessAuthForMiddlewareTests(false)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Override token lookup to simulate empty userID
// 	authInstance.funcUserFindByAuthToken = func(sessionID string, options UserAuthOptions) (userID string, err error) {
// 		return "", nil
// 	}

// 	req, err := http.NewRequest("GET", "/", bytes.NewBuffer([]byte("")))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Provide a bearer token so authToken != ""
// 	req.Header.Set("Authorization", "Bearer some-token")

// 	recorder := httptest.NewRecorder()
// 	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		t.Fatal("next handler should not be called when userID is empty")
// 	})

// 	handler := authInstance.AuthMiddleware(next)
// 	handler.ServeHTTP(recorder, req)

// 	body := recorder.Body.String()
// 	if !bytes.Contains([]byte(body), []byte("auth token is required")) {
// 		t.Fatalf("expected body to contain %q, got %q", "auth token is required", body)
// 	}
// }

// func TestAuthMiddleware_ValidToken_AppendsUserIDToContext(t *testing.T) {
// 	authInstance, err := newPasswordlessAuthForMiddlewareTests(true)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Valid token returns a userID
// 	authInstance.funcUserFindByAuthToken = func(sessionID string, options UserAuthOptions) (userID string, err error) {
// 		if sessionID == "123456" {
// 			return "234567", nil
// 		}
// 		return "", nil
// 	}

// 	req, err := http.NewRequest("GET", "/", bytes.NewBuffer([]byte("")))
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	req.AddCookie(&http.Cookie{Name: CookieName, Value: "123456"})

// 	recorder := httptest.NewRecorder()

// 	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		value := r.Context().Value(AuthenticatedUserID{})
// 		expectedUserID := "234567"
// 		if value != expectedUserID {
// 			t.Fatalf("expected user ID %q in context, got %v", expectedUserID, value)
// 		}
// 	})

// 	handler := authInstance.AuthMiddleware(next)
// 	handler.ServeHTTP(recorder, req)

// 	if recorder.Code != http.StatusOK {
// 		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
// 	}
// }
