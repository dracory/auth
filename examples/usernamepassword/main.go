package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	auth "github.com/dracory/auth"
	authtypes "github.com/dracory/auth/types"
	"golang.org/x/crypto/bcrypt"
)

type passwordUser struct {
	ID           string
	Username     string
	FirstName    string
	LastName     string
	PasswordHash []byte
}

type passwordMemoryStore struct {
	mu            sync.Mutex
	usersByName   map[string]*passwordUser
	sessions      map[string]string // token -> userID
	tempKeys      map[string]string // key -> username
	nextUserIndex int
}

var passwordStore = &passwordMemoryStore{
	usersByName:   make(map[string]*passwordUser),
	sessions:      make(map[string]string),
	tempKeys:      make(map[string]string),
	nextUserIndex: 1,
}

func (s *passwordMemoryStore) userLogin(_ context.Context, username, password string, _ authtypes.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.usersByName[username]
	if !ok {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	return u.ID, nil
}

func (s *passwordMemoryStore) userRegister(_ context.Context, username, password, firstName, lastName string, _ authtypes.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.usersByName[username]; exists {
		return fmt.Errorf("user %s already exists", username)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	id := fmt.Sprintf("user-%d", s.nextUserIndex)
	s.nextUserIndex++

	s.usersByName[username] = &passwordUser{
		ID:           id,
		Username:     username,
		FirstName:    firstName,
		LastName:     lastName,
		PasswordHash: hash,
	}

	return nil
}

func (s *passwordMemoryStore) userFindByUsername(_ context.Context, username, firstName, lastName string, _ authtypes.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.usersByName[username]
	if !ok || u.FirstName != firstName || u.LastName != lastName {
		return "", errors.New("user not found")
	}

	return u.ID, nil
}

func (s *passwordMemoryStore) userPasswordChange(_ context.Context, username, newPassword string, _ authtypes.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.usersByName[username]
	if !ok {
		return errors.New("user not found")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.PasswordHash = hash
	return nil
}

func (s *passwordMemoryStore) logout(_ context.Context, userID string, _ authtypes.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for token, id := range s.sessions {
		if id == userID {
			delete(s.sessions, token)
		}
	}

	return nil
}

func (s *passwordMemoryStore) storeAuthToken(_ context.Context, token, userID string, _ authtypes.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[token] = userID
	return nil
}

func (s *passwordMemoryStore) findByAuthToken(_ context.Context, token string, _ authtypes.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, ok := s.sessions[token]
	if !ok {
		return "", errors.New("invalid auth token")
	}

	return id, nil
}

func (s *passwordMemoryStore) tempKeySet(key, value string, expiresSeconds int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// For simplicity, this example ignores expiration and just stores the value in memory.
	_ = expiresSeconds
	s.tempKeys[key] = value
	return nil
}

func (s *passwordMemoryStore) tempKeyGet(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.tempKeys[key]
	if !ok {
		return "", errors.New("temporary key not found")
	}
	return v, nil
}

func (s *passwordMemoryStore) displayName(userID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.usersByName {
		if u.ID == userID {
			full := strings.TrimSpace(u.FirstName + " " + u.LastName)
			if full != "" {
				return full
			}
			if u.Username != "" {
				return u.Username
			}
			break
		}
	}

	return userID
}

func exampleEmailSend(_ context.Context, email, subject, body string) error {
	fmt.Printf("[username/password] Sending email to %s: %s\n%s", email, subject, body)
	return nil
}

func main() {
	authInstance, err := auth.NewUsernameAndPasswordAuth(auth.ConfigUsernameAndPassword{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/dashboard",
		UseCookies:           true,

		FuncTemporaryKeyGet:     passwordStore.tempKeyGet,
		FuncTemporaryKeySet:     passwordStore.tempKeySet,
		FuncUserStoreAuthToken:  passwordStore.storeAuthToken,
		FuncUserFindByAuthToken: passwordStore.findByAuthToken,
		FuncUserFindByUsername:  passwordStore.userFindByUsername,
		FuncUserLogin:           passwordStore.userLogin,
		FuncUserLogout:          passwordStore.logout,
		FuncEmailSend:           exampleEmailSend,

		EnableRegistration:     true,
		FuncUserRegister:       passwordStore.userRegister,
		FuncUserPasswordChange: passwordStore.userPasswordChange,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	// Mount auth under /auth
	mux.HandleFunc("/auth/", authInstance.Router().ServeHTTP)

	// Public home page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8" />
  <title>Auth Example - Username &amp; Password</title>
</head>
<body>
  <h1>Username &amp; Password Auth Example</h1>

  <p>This example uses an <strong>in-memory store</strong> for users, sessions and reset tokens.</p>

  <h2>How to use</h2>
  <ol>
    <li>Click <a href="%s">Register</a> and create a new user with an email (used as username) and password.</li>
    <li>After registering, go back and click <a href="%s">Login</a> to sign in with the same credentials.</li>
    <li>On successful login you will be redirected to <code>/dashboard</code>, which is protected by middleware.</li>
  </ol>

  <h2>What this demonstrates</h2>
  <ul>
    <li>Using <code>NewUsernameAndPasswordAuth</code> with callback-based storage.</li>
    <li>Protecting routes with <code>WebAuthOrRedirectMiddleware</code>.</li>
    <li>Using cookies for auth token storage.</li>
  </ul>

  <p><strong>Note:</strong> Data is not persisted. Restarting the example clears all users and sessions.</p>
</body>
</html>`,
			authInstance.LinkRegister(), authInstance.LinkLogin()); err != nil {
			log.Printf("failed to write home page response: %v", err)
		}
	})

	// Protected dashboard page
	mux.Handle("/dashboard", authInstance.WebAuthOrRedirectMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := authInstance.GetCurrentUserID(r)
		displayName := passwordStore.displayName(userID)
		if _, err := fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8" />
  <title>Dashboard - Auth Example</title>
</head>
<body>
  <h1>Dashboard</h1>
  <p>You are logged in as: <strong>%s</strong> (id: %s)</p>

  <p>This page is protected by <code>WebAuthOrRedirectMiddleware</code>. If you clear your cookies and
  refresh, you will be redirected back to the login page.</p>

  <p><a href="%s">Logout</a></p>
</body>
</html>`, displayName, userID, authInstance.LinkLogout()); err != nil {
			log.Printf("failed to write dashboard page response: %v", err)
		}
	})))

	fmt.Println("Username/password auth example running on http://localhost:8082")
	if err := http.ListenAndServe(":8082", mux); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
