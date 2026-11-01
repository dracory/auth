// Background example: full-page photo backdrop behind the auth panel.
// Demonstrates that a host-controlled background image never breaks the
// built-in pages — the whole page content lives inside .auth-panel, so
// text stays readable regardless of the photo. A light scrim via
// body::before keeps the optional footer links legible.
// The image comes from picsum.photos (free placeholder image service).
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/dracory/auth"
	"github.com/dracory/auth/types"
)

type bgUser struct {
	ID        string
	Username  string
	Password  string
	FirstName string
	LastName  string
}

type bgMemoryStore struct {
	mu        sync.Mutex
	users     map[string]*bgUser
	sessions  map[string]string
	tempKeys  map[string]string
	nextIndex int
}

var store = &bgMemoryStore{
	users:     make(map[string]*bgUser),
	sessions:  make(map[string]string),
	tempKeys:  make(map[string]string),
	nextIndex: 1,
}

func (s *bgMemoryStore) userLogin(_ context.Context, username, password string, _ types.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[username]
	if !ok || u.Password != password {
		return "", errors.New("invalid credentials")
	}
	return u.ID, nil
}

func (s *bgMemoryStore) userFindByUsername(_ context.Context, username, _, _ string, _ types.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if u, ok := s.users[username]; ok {
		return u.ID, nil
	}
	return "", errors.New("user not found")
}

func (s *bgMemoryStore) storeAuthToken(_ context.Context, token, userID string, _ types.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[token] = userID
	return nil
}

func (s *bgMemoryStore) findByAuthToken(_ context.Context, token string, _ types.UserAuthOptions) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id, ok := s.sessions[token]
	if !ok {
		return "", errors.New("invalid token")
	}
	return id, nil
}

func (s *bgMemoryStore) logout(_ context.Context, userID string, _ types.UserAuthOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for t, id := range s.sessions {
		if id == userID {
			delete(s.sessions, t)
		}
	}
	return nil
}

func (s *bgMemoryStore) tempKeySet(key, value string, _ int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tempKeys[key] = value
	return nil
}

func (s *bgMemoryStore) tempKeyGet(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.tempKeys[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return v, nil
}

func (s *bgMemoryStore) displayName(userID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.users {
		if u.ID == userID {
			full := strings.TrimSpace(u.FirstName + " " + u.LastName)
			if full != "" {
				return full
			}
			return u.Username
		}
	}
	return userID
}

func emailSend(_ context.Context, userID, subject, body string) error {
	log.Printf("[background email] to=%s subject=%s\n", userID, subject)
	return nil
}

// backgroundCSS covers the viewport with a free picsum.photos image and a
// light scrim, and clears the package shell's bg-light so the photo shows
// through. .auth-panel is left untouched — Bootstrap's card styling (and
// dark/light mode via data-bs-theme) applies as usual; a host could
// instead override it, e.g. `.auth-panel { background: rgba(255,255,255,.9) }`.
const backgroundCSS = `body {
	background: url('https://picsum.photos/seed/auth/1920/1080') center center / cover no-repeat fixed;
}
body::before {
	content: '';
	position: fixed;
	inset: 0;
	background: rgba(255, 255, 255, 0.45);
	z-index: -1;
}
.bg-light {
	background-color: transparent !important;
}`

// newAuth seeds the demo user and builds the configured auth instance.
func newAuth() (types.AuthPasswordInterface, error) {
	demoUser := &bgUser{ID: "usr-1", Username: "user@demo.local", Password: "demopassword", FirstName: "Demo", LastName: "User"}
	store.users[demoUser.Username] = demoUser

	return auth.NewUsernameAndPasswordAuth(types.ConfigUsernameAndPassword{
		ConfigShared: types.ConfigShared{
			Endpoint:             "/auth",
			UrlRedirectOnSuccess: "/dashboard",
			UseCookies:           true,
			EnableRegistration:   false,
			BrandName:            "DEMO",
			CustomCSS:            backgroundCSS,
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

		FuncEmailSend:          emailSend,
		FuncUserFindByUsername: store.userFindByUsername,
		FuncUserLogin:          store.userLogin,
	})
}

// newMux builds the example's HTTP handler tree: the auth mux under
// /auth, a public home page, and a middleware-protected dashboard.
func newMux(authInstance types.AuthPasswordInterface) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/", authInstance.Router().ServeHTTP)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<h1>Background example</h1><p><a href="/auth/login">Login over a photo backdrop</a> (user@demo.local / demopassword)</p>`)
	})

	mux.Handle("/dashboard", authInstance.WebAuthOrRedirectMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := authInstance.GetCurrentUserID(r)
		fmt.Fprintf(w, `<h1>Dashboard</h1><p>Welcome, %s (ID: %s)</p><p><a href="/auth/logout">Logout</a></p>`, store.displayName(userID), userID)
	})))

	return mux
}

func main() {
	authInstance, err := newAuth()
	if err != nil {
		log.Fatal(err)
	}

	mux := newMux(authInstance)

	fmt.Println("Background example running on http://localhost:8088")
	if err := http.ListenAndServe(":8088", mux); err != nil {
		log.Fatal(err)
	}
}
