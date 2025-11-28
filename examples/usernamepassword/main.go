package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"

	auth "github.com/dracory/auth"
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

func (s *passwordMemoryStore) userLogin(_ context.Context, username, password string, _ auth.UserAuthOptions) (string, error) {
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

func (s *passwordMemoryStore) userRegister(_ context.Context, username, password, firstName, lastName string, _ auth.UserAuthOptions) error {
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

func (s *passwordMemoryStore) userFindByUsername(_ context.Context, username, firstName, lastName string, _ auth.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.usersByName[username]
	if !ok || u.FirstName != firstName || u.LastName != lastName {
		return "", errors.New("user not found")
	}

	return u.ID, nil
}

func (s *passwordMemoryStore) userPasswordChange(_ context.Context, username, newPassword string, _ auth.UserAuthOptions) error {
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

func (s *passwordMemoryStore) logout(_ context.Context, userID string, _ auth.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for token, id := range s.sessions {
		if id == userID {
			delete(s.sessions, token)
		}
	}

	return nil
}

func (s *passwordMemoryStore) storeAuthToken(_ context.Context, token, userID string, _ auth.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[token] = userID
	return nil
}

func (s *passwordMemoryStore) findByAuthToken(_ context.Context, token string, _ auth.UserAuthOptions) (string, error) {
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
		fmt.Fprintf(w, "<h1>Home</h1><p><a href='%s'>Login</a> | <a href='%s'>Register</a></p>",
			authInstance.LinkLogin(), authInstance.LinkRegister())
	})

	// Protected dashboard page
	mux.Handle("/dashboard", authInstance.WebAuthOrRedirectMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := authInstance.GetCurrentUserID(r)
		fmt.Fprintf(w, "<h1>Dashboard</h1><p>Welcome, user %s!</p><p><a href='%s'>Logout</a></p>", userID, authInstance.LinkLogout())
	})))

	fmt.Println("Username/password auth example running on http://localhost:8082")
	if err := http.ListenAndServe(":8082", mux); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
