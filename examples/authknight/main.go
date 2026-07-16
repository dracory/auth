package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	auth "github.com/dracory/auth"
	"github.com/dracory/auth/types"
)

type authKnightUser struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
}

type authKnightMemoryStore struct {
	mu           sync.Mutex
	usersByEmail map[string]*authKnightUser
	sessions     map[string]string // token -> userID
	tempKeys     map[string]string // ak_key -> email
}

var store = &authKnightMemoryStore{
	usersByEmail: make(map[string]*authKnightUser),
	sessions:     make(map[string]string),
	tempKeys:     make(map[string]string),
}

func (s *authKnightMemoryStore) findUserByEmail(_ context.Context, email string, _ types.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if u, ok := s.usersByEmail[email]; ok {
		return u.ID, nil
	}

	return "", fmt.Errorf("user with email %s not found", email)
}

func (s *authKnightMemoryStore) registerUser(_ context.Context, email, firstName, lastName string, _ types.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.usersByEmail[email]; exists {
		return "", fmt.Errorf("user with email %s already exists", email)
	}

	id := fmt.Sprintf("user-%d", len(s.usersByEmail)+1)
	s.usersByEmail[email] = &authKnightUser{
		ID:        id,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}

	return id, nil
}

func (s *authKnightMemoryStore) logout(_ context.Context, userID string, _ types.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for token, id := range s.sessions {
		if id == userID {
			delete(s.sessions, token)
		}
	}

	return nil
}

func (s *authKnightMemoryStore) storeAuthToken(_ context.Context, token, userID string, _ types.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[token] = userID
	return nil
}

func (s *authKnightMemoryStore) findByAuthToken(_ context.Context, token string, _ types.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, ok := s.sessions[token]
	if !ok {
		return "", fmt.Errorf("invalid auth token")
	}
	return id, nil
}

func (s *authKnightMemoryStore) tempKeySet(key, value string, expiresSeconds int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = expiresSeconds
	s.tempKeys[key] = value
	return nil
}

func (s *authKnightMemoryStore) tempKeyGet(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.tempKeys[key]
	if !ok {
		return "", fmt.Errorf("temporary key not found")
	}
	return v, nil
}

func (s *authKnightMemoryStore) displayName(userID string) string {
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

func main() {
	authInstance, err := auth.NewAuthKnightAuth(types.ConfigAuthKnight{
		ConfigShared: types.ConfigShared{
			Endpoint:             "/auth",
			UrlRedirectOnSuccess: "/dashboard",
			UseCookies:           true,
			EnableRegistration:   true,
			CookieConfig: &types.CookieConfig{
				Secure:   types.CookieInsecure,
				HttpOnly: types.CookieHttpOnly,
				SameSite: http.SameSiteLaxMode,
				Path:     "/",
				MaxAge:   2 * 60 * 60,
			},

			FuncUserFindByAuthToken: store.findByAuthToken,
			FuncUserLogout:          store.logout,
			FuncUserStoreAuthToken:  store.storeAuthToken,
			FuncTemporaryKeyGet:     store.tempKeyGet,
			FuncTemporaryKeySet:     store.tempKeySet,
		},

		FuncUserFindByEmail: store.findUserByEmail,
		FuncUserRegister:    store.registerUser,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	mux := http.NewServeMux()

	// Mount auth under /auth
	mux.HandleFunc("/auth/", authInstance.Router().ServeHTTP)

	// Public home page with AuthKnight login link
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		loginURL := authInstance.LinkAuthKnightRedirect(r)
		fmt.Printf("[HOME] loginURL='%s'\n", loginURL)
		if _, err := fmt.Fprintf(w, "<h1>Home</h1><p><a href='%s'>Login with AuthKnight</a></p>", loginURL); err != nil {
			log.Printf("failed to write home page response: %v", err)
		}
	})

	// Protected dashboard page
	mux.Handle("/dashboard", authInstance.WebAuthOrRedirectMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := authInstance.GetCurrentUserID(r)
		displayName := store.displayName(userID)
		if _, err := fmt.Fprintf(w, "<h1>Dashboard</h1><p>Welcome, %s (id: %s)!</p><p><a href='%s'>Logout</a></p>", displayName, userID, authInstance.LinkLogout()); err != nil {
			log.Printf("failed to write dashboard page response: %v", err)
		}
	})))

	// Catch-all debug logger for all other requests
	mux.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[DEBUG] Request URL: %s\n", r.URL.String())
		fmt.Printf("[DEBUG] Request URI: %s\n", r.RequestURI)
		fmt.Fprintf(w, "Check console for debug info")
	})

	fmt.Println("AuthKnight auth example running on http://localhost:8084")
	if err := http.ListenAndServe(":8084", mux); err != nil {
		fmt.Println(err)
	}
}
