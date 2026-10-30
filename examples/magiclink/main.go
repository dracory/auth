// Magic link example: standalone magic link authentication flow.
// Users enter their email and receive a single-use sign-in link.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/dracory/auth"
	"github.com/dracory/auth/types"
)

type magicLinkUser struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
}

type magicLinkMemoryStore struct {
	mu           sync.Mutex
	usersByEmail map[string]*magicLinkUser
	sessions     map[string]string // token -> userID
	tempKeys     map[string]string // key -> value
}

var store = &magicLinkMemoryStore{
	usersByEmail: make(map[string]*magicLinkUser),
	sessions:     make(map[string]string),
	tempKeys:     make(map[string]string),
}

func (s *magicLinkMemoryStore) findUserByEmail(_ context.Context, email string, _ types.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if u, ok := s.usersByEmail[email]; ok {
		return u.ID, nil
	}
	return "", fmt.Errorf("user with email %s not found", email)
}

func (s *magicLinkMemoryStore) registerUser(_ context.Context, email, firstName, lastName string, _ types.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.usersByEmail[email]; exists {
		return fmt.Errorf("user with email %s already exists", email)
	}

	id := fmt.Sprintf("user-%d", len(s.usersByEmail)+1)
	s.usersByEmail[email] = &magicLinkUser{
		ID:        id,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
	}
	return nil
}

func (s *magicLinkMemoryStore) logout(_ context.Context, userID string, _ types.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for token, id := range s.sessions {
		if id == userID {
			delete(s.sessions, token)
		}
	}
	return nil
}

func (s *magicLinkMemoryStore) storeAuthToken(_ context.Context, token, userID string, _ types.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[token] = userID
	return nil
}

func (s *magicLinkMemoryStore) findByAuthToken(_ context.Context, token string, _ types.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, ok := s.sessions[token]
	if !ok {
		return "", fmt.Errorf("invalid auth token")
	}
	return id, nil
}

func (s *magicLinkMemoryStore) tempKeySet(key, value string, expiresSeconds int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = expiresSeconds
	s.tempKeys[key] = value
	return nil
}

func (s *magicLinkMemoryStore) tempKeyGet(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.tempKeys[key]
	if !ok {
		return "", fmt.Errorf("temporary key not found")
	}
	return v, nil
}

func (s *magicLinkMemoryStore) displayName(userID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.usersByEmail {
		if u.ID == userID {
			full := strings.TrimSpace(u.FirstName + " " + u.LastName)
			if full != "" {
				return full
			}
			return u.Email
		}
	}
	return userID
}

func emailSend(_ context.Context, email, subject, body string) error {
	log.Printf("[magiclink email] to=%s subject=%s\n%s\n", email, subject, body)
	return nil
}

func main() {
	// Seed demo user
	_ = store.registerUser(context.Background(), "user@example.com", "Magic", "User", types.UserAuthOptions{})

	shared := types.ConfigShared{
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
	}

	passwordlessAuth, err := auth.NewPasswordlessAuth(types.ConfigPasswordless{
		ConfigShared:        shared,
		FuncUserFindByEmail: store.findUserByEmail,
		FuncEmailSend:       emailSend,
		FuncUserRegister:    store.registerUser,
	})
	if err != nil {
		log.Fatal(err)
	}

	magicLinkMethod := auth.NewMagicLinkMethodFromAuth(passwordlessAuth, nil, func(token string) string {
		return "/auth/magiclink-callback?token=" + token
	})

	multi, err := auth.NewMultiMethodAuth("/auth", []types.Method{magicLinkMethod})
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/auth/", multi.Handler())

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<h1>Magic Link Demo</h1><p><a href="/auth/login">Login via Magic Link</a></p>`)
	})

	mux.Handle("/dashboard", passwordlessAuth.WebAuthOrRedirectMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := passwordlessAuth.GetCurrentUserID(r)
		fmt.Fprintf(w, `<h1>Dashboard</h1><p>Welcome, %s (id: %s)!</p><p><a href="/auth/logout">Logout</a></p>`, store.displayName(userID), userID)
	})))

	fmt.Println("Magic Link auth example running on http://localhost:8085")
	if err := http.ListenAndServe(":8085", mux); err != nil {
		log.Fatal(err)
	}
}
