package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	auth "github.com/dracory/auth"
	"github.com/jordan-wright/email"
)

type passwordlessUser struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
}

type passwordlessMemoryStore struct {
	mu           sync.Mutex
	usersByEmail map[string]*passwordlessUser
	sessions     map[string]string // token -> userID
	tempKeys     map[string]string // code -> email
}

var passwordlessStore = &passwordlessMemoryStore{
	usersByEmail: make(map[string]*passwordlessUser),
	sessions:     make(map[string]string),
	tempKeys:     make(map[string]string),
}

func (s *passwordlessMemoryStore) findUserByEmail(_ context.Context, email string, _ auth.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if u, ok := s.usersByEmail[email]; ok {
		return u.ID, nil
	}

	return "", fmt.Errorf("user with email %s not found", email)
}

func (s *passwordlessMemoryStore) registerUser(_ context.Context, email, firstName, lastName string, _ auth.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.usersByEmail[email]; exists {
		return fmt.Errorf("user with email %s already exists", email)
	}

	id := fmt.Sprintf("user-%d", len(s.usersByEmail)+1)
	s.usersByEmail[email] = &passwordlessUser{
		ID:        id,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}

	return nil
}

func (s *passwordlessMemoryStore) logout(_ context.Context, userID string, _ auth.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for token, id := range s.sessions {
		if id == userID {
			delete(s.sessions, token)
		}
	}

	return nil
}

func (s *passwordlessMemoryStore) storeAuthToken(_ context.Context, token, userID string, _ auth.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[token] = userID
	return nil
}

func (s *passwordlessMemoryStore) findByAuthToken(_ context.Context, token string, _ auth.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, ok := s.sessions[token]
	if !ok {
		return "", fmt.Errorf("invalid auth token")
	}
	return id, nil
}

func (s *passwordlessMemoryStore) tempKeySet(key, value string, expiresSeconds int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// For simplicity, this example ignores expiration and just stores the value in memory.
	_ = expiresSeconds
	s.tempKeys[key] = value
	return nil
}

func (s *passwordlessMemoryStore) tempKeyGet(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.tempKeys[key]
	if !ok {
		return "", fmt.Errorf("temporary key not found")
	}
	return v, nil
}

func (s *passwordlessMemoryStore) displayName(userID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.usersByEmail {
		if u.ID == userID {
			full := strings.TrimSpace(u.FirstName + " " + u.LastName)
			if full != "" {
				return full
			}
			if u.Email != "" {
				return u.Email
			}
			break
		}
	}

	return userID
}

func passwordlessEmailSend(_ context.Context, to, subject, body string) error {
	log.Printf("[passwordless] Sending email to %s via localhost:1025: %s", to, subject)

	e := email.NewEmail()
	e.From = "no-reply@example.com"
	e.To = []string{to}
	e.Subject = subject
	e.HTML = []byte(body)

	// For local dev mail server listening on :1025 with no auth
	return e.Send("localhost:1025", nil)
}

func main() {
	authInstance, err := auth.NewPasswordlessAuth(auth.ConfigPasswordless{
		Endpoint:             "/auth",
		UrlRedirectOnSuccess: "/dashboard",
		UseCookies:           true,

		FuncUserFindByAuthToken: passwordlessStore.findByAuthToken,
		FuncUserFindByEmail:     passwordlessStore.findUserByEmail,
		FuncUserLogout:          passwordlessStore.logout,
		FuncUserStoreAuthToken:  passwordlessStore.storeAuthToken,
		FuncEmailSend:           passwordlessEmailSend,
		FuncTemporaryKeyGet:     passwordlessStore.tempKeyGet,
		FuncTemporaryKeySet:     passwordlessStore.tempKeySet,

		EnableRegistration: true,
		FuncUserRegister:   passwordlessStore.registerUser,
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
		if _, err := fmt.Fprintf(w, "<h1>Home</h1><p><a href='%s'>Login</a> | <a href='%s'>Register</a></p>",
			authInstance.LinkLogin(), authInstance.LinkRegister()); err != nil {
			log.Printf("failed to write home page response: %v", err)
		}
	})

	// Protected dashboard page
	mux.Handle("/dashboard", authInstance.WebAuthOrRedirectMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := authInstance.GetCurrentUserID(r)
		displayName := passwordlessStore.displayName(userID)
		if _, err := fmt.Fprintf(w, "<h1>Dashboard</h1><p>Welcome, %s (id: %s)!</p><p><a href='%s'>Logout</a></p>", displayName, userID, authInstance.LinkLogout()); err != nil {
			log.Printf("failed to write dashboard page response: %v", err)
		}
	})))

	fmt.Println("Passwordless auth example running on http://localhost:8083")
	if err := http.ListenAndServe(":8083", mux); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
